package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/api"
	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/proxy"
	"github.com/will-x86/bdns/dns/pkg/rcache"
	"github.com/will-x86/bdns/dns/pkg/rule"
	"github.com/will-x86/bdns/dns/pkg/store"
)

type DNSUpstream interface {
	SendQuery([]byte) ([]byte, error)
}

// Builds refused response (RCODE=5)
// Preserves tID and qSection
func buildRefusedResponse(query []byte) []byte {
	if len(query) < 12 {
		return nil
	}
	resp := make([]byte, len(query))
	copy(resp, query)

	// Flags: QR=1 (response), Opcode=0, AA=0, TC=0, RD = copy from query,
	// RA=0, Z=0, RCODE=5 (REFUSED).
	rdBit := query[2] & 0x01 // RD from query
	resp[2] = 0x80 | rdBit   // QR=1, Opcode=0, AA=0, TC=0, RD=original
	resp[3] = 0x05           // RA=0, Z=0, RCODE=5

	// Zero out answer/authority/additional counts.
	resp[6] = 0
	resp[7] = 0
	resp[8] = 0
	resp[9] = 0
	resp[10] = 0
	resp[11] = 0

	return resp
}

type ServerConfig struct {
	Port       int
	PrivateKey string
	SignedKey  string
	ValkeyAddr string
	APIAddr    string // e.g. ":8080"; empty disables the management API
}

// Print all files in cert dir & panic, to hopefully be useful to user
func tlsNiceExitNoCert(ctx context.Context, dir string, err error) {
	log := zerolog.Ctx(ctx)
	if dir == "" {
		log.Fatal().Msg("cert/privkey dir env var is empty, cannot read tls certificate")
	}
	directory := strings.Split(dir, "/")
	// Assume /dir/dir/example.{pem/crt}
	if len(directory) > 0 {
		entires, dirErr := os.ReadDir(strings.Join(directory[0:len(directory)-1], ""))
		if dirErr != nil {
			log.Fatal().Err(err).Msg("error reading tls cert")
		}
		for _, v := range entires {
			log.Info().Str("entry in cert dir", v.Name()).Send()
		}
	} else {
		log.Debug().Any("directory", directory).Send()
	}
	log.Fatal().Err(err).Msg("tls cert direcotry error, path cannot be parsed either")
}
func RunServer(ctx context.Context, c *ServerConfig) {
	log := zerolog.Ctx(ctx).With().Str("component", "server").Logger()
	cert, err := tls.LoadX509KeyPair(c.SignedKey, c.PrivateKey)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			tlsNiceExitNoCert(ctx, c.SignedKey, err)
		}
		log.Fatal().Err(err).Msg("cannot load certificate")
	}
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", c.Port), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Fatal().Err(err).Int("port", c.Port).Msg("failed to listen on port")
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		log.Info().Msg("Shutting down server")
		listener.Close()
	}()

	cache, err := rcache.New(c.ValkeyAddr)
	if err != nil {
		log.Warn().Err(err).Str("valkey-addr", c.ValkeyAddr).Msg("could not connect to valkey, continuing without cache")
		cache = nil
	}

	stores := db.NewStores(db.GetDB())
	var poolCacheStore store.Pool
	poolCacheStore, err = store.NewValkey(ctx, c.ValkeyAddr, stores)
	if err != nil {
		log.Warn().Err(err).Str("valkey-addr", c.ValkeyAddr).Msg("valkey had an error, default to memory storage for pool limits")
		poolCacheStore = store.NewMemory()

	}
	resetter := store.NewResetter(stores, poolCacheStore)
	resetter.StartResetJob(ctx)

	// Management API (profiles/whitelists/pools/etc.) shares this process and DB.
	if c.APIAddr != "" {
		repo := db.NewRepo(db.GetDB())
		go func() {
			if err := api.Serve(ctx, c.APIAddr, repo, poolCacheStore); err != nil {
				log.Error().Err(err).Str("addr", c.APIAddr).Msg("management API stopped")
			}
		}()
	}
	/*members, err := stores.GetAllPoolMembersWithTimezones(ctx)
	if err != nil {
		panic(err)
	}
	log.Info().Int("member-count", len(members)).Send()*/
	ruleStores := rule.Stores{
		Profile:   stores,
		Whitelist: stores,
		Category:  stores,
		TimeBlock: stores,
		Resolve:   stores.ResolveCategory,
		PoolCache: poolCacheStore,
		PoolDB:    stores,
	}
	engine := proxy.BuildEngine(ruleStores)

	upstream := proxy.NewDoHClient("https://cloudflare-dns.com/dns-query")
	log.Info().Int("port", c.Port).Msg("listening...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("error on accepting listener for a conn")
			continue
		}
		go func(c net.Conn) {
			log := log.With().Str("remote", c.RemoteAddr().String()).Logger()
			connCtx := log.WithContext(ctx)
			defer c.Close()
			tlsConn := c.(*tls.Conn)
			if err := tlsConn.Handshake(); err != nil {
				log.Warn().Err(err).Msg("tls handkshake error")
				return
			}
			fullSNI := tlsConn.ConnectionState().ServerName
			var profileID string
			if strings.Contains(fullSNI, ".") {
				parts := strings.SplitN(fullSNI, ".", 2)
				profileID = parts[0]
			}
			if profileID == "" {
				log.Warn().Msg("no sni/profileID")
				return
			}

			log.Debug().Str("sni", fullSNI).Str("profileID", profileID).Send()

			var mu sync.Mutex
			h := &handler{
				upstream: upstream,
				cache:    cache,
				write: func(response []byte) error {
					prefix := make([]byte, 2)
					binary.BigEndian.PutUint16(prefix, uint16(len(response)))
					mu.Lock()
					defer mu.Unlock()
					_, err := c.Write(append(prefix, response...))
					if err != nil {
						return fmt.Errorf("error writing to response: %w", err)
					}
					return nil
				},
				engine:    engine,
				stores:    ruleStores,
				profileID: profileID,
			}

			for {
				var msgLen uint16
				if err := binary.Read(c, binary.BigEndian, &msgLen); err != nil {
					if err != io.EOF {
						log.Warn().Err(err).Msg("Error reading length prefix")
					}
					return
				}

				buf := make([]byte, msgLen)
				if _, err := io.ReadFull(c, buf); err != nil {
					log.Warn().Err(err).Msg("Error reading DNS message")
					return
				}

				go h.handle(connCtx, buf, c.RemoteAddr().String())
			}
		}(conn)
	}
}
