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
	"time"

	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/parser"
	"github.com/will-x86/bdns/dns/pkg/proxy"
	"github.com/will-x86/bdns/dns/pkg/rcache"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

type DNSUpstream interface {
	SendQuery([]byte) ([]byte, error)
}

// handler holds stuff for serving a single connection.
type handler struct {
	upstream DNSUpstream
	cache    *rcache.Cache
	write    func([]byte) error
	engine   *rule.Engine
	stores   rule.Stores
	userID   string
}

func (h *handler) handle(ctx context.Context, requestBytes []byte, remoteAddr string) {
	log.Printf("Received request from %s\n", remoteAddr)

	msg := parser.Message()
	if err := msg.Parse(requestBytes); err != nil {
		log.Printf("Error parsing DNS message from %s: %v\n", remoteAddr, err)
		return
	}
	log.Printf("Parsed request: %s", msg.Header.String())

	if len(msg.Questions) == 0 {
		log.Printf("Request from %s has no questions — dropping\n", remoteAddr)
		return
	}
	if len(msg.Questions) > 1 {
		log.Printf("Request from %s has %d questions\n", remoteAddr, len(msg.Questions))
	}

	q := msg.Questions[0]
	qtypeStr, ok := parser.TypeToString[q.QType]
	if !ok {
		qtypeStr = "UNKNOWN"
	}

	// Refuse non-authed users
	if h.userID == "" {
		log.Printf("No SNI from %s — refusing\n", remoteAddr)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}
	if userExists, err := db.UserExists(h.userID); err != nil {
		log.Printf("DB error checking user %s: %v\n", h.userID, err)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	} else if !userExists {
		log.Printf("User %s not found in DB\n", h.userID)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}

	decision, ruleErr := h.engine.Evaluate(ctx, &rule.RuleContext{
		Domain: q.QName,
		UserID: h.userID,
		Now:    time.Now(),
		Stores: h.stores,
	})
	if ruleErr != nil {
		log.Printf("Rule engine error for %s %s: %v\n", q.QName, qtypeStr, ruleErr)
		// Fail open
	} else if decision.Verdict == rule.VerdictBlock {
		log.Printf("Blocked %s %s for user %s: %s\n", q.QName, qtypeStr, h.userID, decision.Reason)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}

	var (
		responseBytes []byte
		err           error
	)

	if h.cache != nil {
		responseBytes, err = h.cache.QueryRace(
			ctx,
			requestBytes,
			q.QName,
			qtypeStr,
			func(ctx context.Context) ([]byte, error) {
				return h.upstream.SendQuery(requestBytes)
			},
		)
	} else {
		responseBytes, err = h.upstream.SendQuery(requestBytes)
	}

	if err != nil {
		log.Printf("Error resolving %s %s: %v\n", q.QName, qtypeStr, err)
		return
	}

	if err := h.write(responseBytes); err != nil {
		log.Printf("Error sending response to client %s: %v\n", remoteAddr, err)
		return
	}
	log.Printf("Sent response to client %s\n", remoteAddr)

	resMsg := parser.Message()
	if err := resMsg.Parse(responseBytes); err != nil {
		log.Printf("Error parsing DNS response: %v\n", err)
		return
	}
	log.Printf("Parsed response %s", resMsg.Header.String())
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
		Whitelist: stores,
		Category:  stores,
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
			var userID string
			if strings.Contains(fullSNI, ".") {
				parts := strings.SplitN(fullSNI, ".", 2)
				userID = parts[0]
			}

			log.Printf("Client SNI: %s\n", userID)

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
				engine: engine,
				stores: ruleStores,
				userID: userID,
			}
			h.handle(ctx, buf, c.RemoteAddr().String())
		}(conn)
	}
}
