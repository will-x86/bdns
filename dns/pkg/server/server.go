package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"

	"github.com/will-x86/bdns/dns/pkg/parser"
	"github.com/will-x86/bdns/dns/pkg/proxy"
	"github.com/will-x86/bdns/dns/pkg/rcache"
)

type DNSUpstream interface {
	SendQuery([]byte) ([]byte, error)
}

// handler holds stuff for serving a single Connection
type handler struct {
	upstream DNSUpstream
	cache    *rcache.Cache
	write    func([]byte) error
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
	// Should only process first
	if len(msg.Questions) > 1 {
		log.Printf("Request from %s has %d questions\n", remoteAddr, len(msg.Questions))
	}

	q := msg.Questions[0]
	qtypeStr, ok := parser.TypeToString[q.QType]
	if !ok {
		qtypeStr = "UNKNOWN"
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

func RunServer(ctx context.Context, certFile, keyFile string) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := tls.Listen("tcp", ":8533", &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		listener.Close()
	}()

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "localhost:6379"
	}
	cache, err := rcache.New(valkeyAddr)
	if err != nil {
		log.Printf("could not connect to Valkey at %s: %v -continuing without cache", valkeyAddr, err)
		cache = nil
	}

	upstream := proxy.NewTLSClient("1.1.1.1", 853, "cloudflare-dns.com")
	log.Println("Listening on TLS :8533")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			tlsConn := c.(*tls.Conn)
			if err := tlsConn.Handshake(); err != nil {
				log.Printf("TLS handshake error: %v\n", err)
				return
			}
			log.Printf("Client SNI: %s\n", tlsConn.ConnectionState().ServerName)

			// TCP DNS: 2-byte big-endian length prefix per RFC 1035 4.2.2 - https://datatracker.ietf.org/doc/html/rfc1035#section-4.2.2
			var msgLen uint16
			if err := binary.Read(c, binary.BigEndian, &msgLen); err != nil {
				log.Printf("Error reading TCP length prefix: %v\n", err)
				return
			}
			buf := make([]byte, msgLen)
			if _, err := io.ReadFull(c, buf); err != nil {
				log.Printf("Error reading TCP DNS message: %v\n", err)
				return
			}

			h := &handler{
				upstream: upstream,
				cache:    cache,
				write: func(response []byte) error {
					prefix := make([]byte, 2)
					binary.BigEndian.PutUint16(prefix, uint16(len(response)))
					_, err := c.Write(append(prefix, response...))
					return err
				},
			}
			h.handle(ctx, buf, c.RemoteAddr().String())
		}(conn)
	}
}
