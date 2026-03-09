package main

// https://github.com/cmol/dns
// https://harshagarwal29.hashnode.dev/building-a-dns-resolver-in-golang-a-step-by-step-guide
import (
	"crypto/tls"
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
			go handleTLSClient(conn)
		}
	}()
	// Run non-TLS DNS server on UDP port 1053
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
		requestBytes := make([]byte, 512)
		_, clientAddr, err := serverConn.ReadFromUDP(requestBytes)
		if err != nil {
			log.Println("Error receiving: ", err.Error())
		} else {
			log.Println("Received request from ", clientAddr)
			go handleDNSClient(requestBytes, serverConn, clientAddr) // array is value type (call-by-value), i.e. copied
		}
	}

}

func handleTLSClient(conn net.Conn) {
	defer conn.Close()

	// Read DNS query
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	// Process and respond (use your existing logic here)
	_ = n
	log.Printf("Received request from %s\n", conn.RemoteAddr())
}

/*
	func main() {
		serverAddr, err := net.ResolveUDPAddr("udp", ":1053")

		if err != nil {
			log.Println("errr resolving UDP address: ", err.Error())
			os.Exit(1)
		}

		serverConn, err := net.ListenUDP("udp", serverAddr)

		if err != nil {
			log.Println("Error listening: ", err.Error())
			os.Exit(1)
		}

		log.Println("listen at: ", serverAddr)

		defer serverConn.Close()

		for {
			requestBytes := make([]byte, 512)

			_, clientAddr, err := serverConn.ReadFromUDP(requestBytes)

			if err != nil {
				log.Println("Error receiving: ", err.Error())
			} else {
				log.Println("Received request from ", clientAddr)
				go handleDNSClient(requestBytes, serverConn, clientAddr) // array is value type (call-by-value), i.e. copied
			}
		}
	}
*/
func handleDNSClient(requestBytes []byte, serverConn *net.UDPConn, clientAddr *net.UDPAddr) {
	message := parser.Message()
	err := message.Parse(requestBytes)
	if err != nil {
		panic(err)
	}
	for i, q := range message.Questions {
		log.Printf("Question %d: QName=%s, QType=%d, QClass=%d\n", i+1, q.QName, q.QType, q.QClass)
	}
	proxy := proxy.NewClient("1.1.1.1", 53)
	responseBytes, err := proxy.SendQuery(requestBytes)
	if err != nil {
		log.Printf("Error sending query to upstream DNS server: %v\n", err)
		return
	}
	_, err = serverConn.WriteToUDP(responseBytes, clientAddr)
	if err != nil {
		log.Printf("Error sending response to client: %v\n", err)
		return
	}
	log.Printf("Sent response to client %s\n", clientAddr.String())

}
