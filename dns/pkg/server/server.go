package server

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"

	"github.com/will-x86/bdns/dns/pkg/parser"
	"github.com/will-x86/bdns/dns/pkg/proxy"
	//"github.com/will-x86/bdns/dns/pkg/rcache"
)

type DNSUpstream interface {
	SendQuery([]byte) ([]byte, error)
}

func RunServer(ctx context.Context, certFile, keyFile string) {
	/*cache, err := rcache.New("127.0.0.1:6379")
	if err != nil {
		log.Fatalf("Failed to initialize valkey client: %v\n", err)
	}*/

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", ":8533", tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		log.Println("Shutting down server...")
		listener.Close()
	}()
	log.Println("Listening on TLS :8533")
	upstream := proxy.NewTLSClient("1.1.1.1", 853, "cloudflare-dns.com")
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
			sni := tlsConn.ConnectionState().ServerName
			log.Printf("Client SNI: %s\n", sni)

			// TCP DNS: 2-byte big-endian length prefix per RFC 1035 §4.2.2
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
			handleDNSClient(buf, upstream, func(response []byte) error {
				// Prepend 2-byte length prefix on the response
				prefix := make([]byte, 2)
				binary.BigEndian.PutUint16(prefix, uint16(len(response)))
				_, err := c.Write(append(prefix, response...))
				return err
			}, c.RemoteAddr().String())
		}(conn)
	}

}
func handleDNSClient(requestBytes []byte, upstream DNSUpstream, write func([]byte) error, remoteAddr string) {
	log.Printf("Received request from %s\n", remoteAddr)

	message := parser.Message()
	err := message.Parse(requestBytes)
	if err != nil {
		log.Printf("Error parsing DNS message: %v\n", err)
		return
	}
	for i, q := range message.Questions {
		log.Printf("Question %d: QName=%s, QType=%d, QClass=%d\n", i+1, q.QName, q.QType, q.QClass)
	}

	responseBytes, err := upstream.SendQuery(requestBytes)
	//proxy := proxy.NewTLSClient("1.1.1.1", 853, "cloudflare-dns.com")

	if err := write(responseBytes); err != nil {
		log.Printf("Error sending response to client %s: %v\n", remoteAddr, err)
		return
	}
	log.Printf("Sent response to client %s\n", remoteAddr)
}
