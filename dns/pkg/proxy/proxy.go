// / Proxies to cloudflare, should later on retry if fail
package proxy

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"slices"
)

type Client struct {
	serverAddress string
	port          int
	serverName    string // SNI for TLS
}

func NewTLSClient(address string, port int, serverName string) *Client {
	return &Client{serverAddress: address, port: port, serverName: serverName}
}

func (c *Client) isIPV4() (bool, error) {
	ip := net.ParseIP(c.serverAddress)
	if ip.To4() != nil {
		return true, nil
	} else if ip.To16() != nil {
		return false, nil
	}
	return false, fmt.Errorf("invalid IP address: %s", c.serverAddress)
}

func (c *Client) addr() (string, error) {
	isIPV4, err := c.isIPV4()
	if err != nil {
		return "", err
	}
	if isIPV4 {
		return fmt.Sprintf("%s:%d", c.serverAddress, c.port), nil
	}
	return fmt.Sprintf("[%s]:%d", c.serverAddress, c.port), nil
}

func (c *Client) SendQuery(query []byte) ([]byte, error) {
	addr, err := c.addr()
	if err != nil {
		log.Printf("Error determining address: %v\n", err)
		return nil, err
	}

	return c.sendQueryTLS(addr, query)
}

func (c *Client) sendQueryTLS(addr string, query []byte) ([]byte, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: c.serverName,
	})
	if err != nil {
		return nil, fmt.Errorf("dial tls: %w", err)
	}
	defer conn.Close()

	// TCP DNS: 2-byte big-endian length prefix per RFC 1035 §4.2.2
	prefix := make([]byte, 2)
	binary.BigEndian.PutUint16(prefix, uint16(len(query)))
	if _, err = conn.Write(append(prefix, query...)); err != nil {
		return nil, fmt.Errorf("write tls: %w", err)
	}

	var respLen uint16
	if err := binary.Read(conn, binary.BigEndian, &respLen); err != nil {
		return nil, fmt.Errorf("read tls length prefix: %w", err)
	}
	response := make([]byte, respLen)
	if _, err := io.ReadFull(conn, response); err != nil {
		return nil, fmt.Errorf("read tls response: %w", err)
	}

	if !hasTheSameID(query, response) {
		return nil, fmt.Errorf("response ID mismatch")
	}
	return response, nil
}

func hasTheSameID(query, response []byte) bool {
	return slices.Equal(query[:2], response[:2])
}
