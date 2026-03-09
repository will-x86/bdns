package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"

	"github.com/will-x86/bdns/dns/pkg/parser"
	"github.com/will-x86/bdns/dns/pkg/proxy"
)

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
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

	log.Println("Listening on TLS :8533")
	go func() {
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
				handleDNSClient(buf, func(response []byte) error {
					// Prepend 2-byte length prefix on the response
					prefix := make([]byte, 2)
					binary.BigEndian.PutUint16(prefix, uint16(len(response)))
					_, err := c.Write(append(prefix, response...))
					return err
				}, c.RemoteAddr().String())
			}(conn)
		}
	}()

	// non-TLS DNS server
	serverAddr, err := net.ResolveUDPAddr("udp", ":1053")
	if err != nil {
		log.Println("Error resolving UDP address: ", err.Error())
		os.Exit(1)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	log.Println("Listening on UDP :1053")
	defer serverConn.Close()

	for {
		requestBytes := make([]byte, 4096)
		_, clientAddr, err := serverConn.ReadFromUDP(requestBytes)
		if err != nil {
			log.Println("Error receiving: ", err.Error())
		} else {
			go handleDNSClient(requestBytes, func(response []byte) error {
				_, err := serverConn.WriteToUDP(response, clientAddr)
				return err
			}, clientAddr.String())
		}
	}
}

func handleDNSClient(requestBytes []byte, write func([]byte) error, remoteAddr string) {
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

	proxy := proxy.NewTLSClient("1.1.1.1", 853, "cloudflare-dns.com")
	responseBytes, err := proxy.SendQuery(requestBytes)
	if err != nil {
		log.Printf("Error sending query to upstream DNS server: %v\n", err)
		return
	}

	if err := write(responseBytes); err != nil {
		log.Printf("Error sending response to client %s: %v\n", remoteAddr, err)
		return
	}
	log.Printf("Sent response to client %s\n", remoteAddr)
}
