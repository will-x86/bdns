package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"strings"

	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/proxy"
	"github.com/will-x86/bdns/dns/pkg/rcache"
	"github.com/will-x86/bdns/dns/pkg/rule"
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
}

// Print all files in cert dir & panic, to hopefully be useful to user
func tlsNiceExitNoCert(dir string, err error) {
	if dir == "" {
		log.Fatal("cert dir is \"\", cannot read tls certificate")
	}
	directory := strings.Split(dir, "/")
	// Assume /dir/dir/example.{pem/crt}
	if len(directory) > 0 {
		entires, dirErr := os.ReadDir(strings.Join(directory[0:len(directory)-1], ""))
		if dirErr != nil {
			log.Fatalf("error reading tls cert: %+v", err)
		}
		for _, v := range entires {
			log.Printf("Entry in cert dir: %s", v.Name())
		}
	} else {
		log.Printf("Directory: %v", directory)
	}
	log.Fatalf("tls cert directory error, path cannot be parsed either %v", err) // Exit
}
func RunServer(ctx context.Context, c *ServerConfig) {

	cert, err := tls.LoadX509KeyPair(c.SignedKey, c.PrivateKey)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			tlsNiceExitNoCert(c.SignedKey, err)
		}
		log.Fatalf("cannot load certificate %v", err)
	}
	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", c.Port), &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Fatalf("failed to listen on port %d, with error: %v", c.Port, err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		listener.Close()
	}()

	cache, err := rcache.New(c.ValkeyAddr)
	if err != nil {
		log.Printf("could not connect to Valkey at %s: %v — continuing without cache", c.ValkeyAddr, err)
		cache = nil
	}

	stores := db.NewStores(db.GetDB())
	ruleStores := rule.Stores{
		Profile:   stores,
		Whitelist: stores,
		Category:  stores,
		TimeBlock: stores,
		Resolve:   stores.ResolveCategory,
	}
	engine := proxy.BuildEngine(ruleStores)

	upstream := proxy.NewTLSClient("1.1.1.1", 853, "cloudflare-dns.com")
	log.Printf("Listening on TLS: %d", c.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error on accepting listender for a connection :", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			tlsConn := c.(*tls.Conn)
			if err := tlsConn.Handshake(); err != nil {
				log.Printf("TLS handshake error: %v\n", err)
				return
			}
			fullSNI := tlsConn.ConnectionState().ServerName
			var profileID string
			if strings.Contains(fullSNI, ".") {
				parts := strings.SplitN(fullSNI, ".", 2)
				profileID = parts[0]
			}

			log.Printf("Client SNI(profileID): %s\n", profileID)

			// TCP DNS: 2-byte big-endian length prefix per RFC 1035 4.2.2
			var msgLen uint16
			if err := binary.Read(c, binary.BigEndian, &msgLen); err != nil {
				log.Printf("Error reading TCP length prefix: %+v\n", err)
				return
			}
			buf := make([]byte, msgLen)
			if _, err := io.ReadFull(c, buf); err != nil {
				log.Printf("Error reading TCP DNS message: %+v\n", err)
				return
			}

			h := &handler{
				upstream: upstream,
				cache:    cache,
				write: func(response []byte) error {
					prefix := make([]byte, 2)
					binary.BigEndian.PutUint16(prefix, uint16(len(response)))
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
			h.handle(ctx, buf, c.RemoteAddr().String())
		}(conn)
	}
}
