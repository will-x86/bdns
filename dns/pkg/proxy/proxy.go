package proxy

import (
	"fmt"
	"log"
	"net"
	"os"
	"slices"
)

type Client struct {
	serverAddress string
	port          int
}

func NewClient(address string, port int) *Client {
	return &Client{serverAddress: address, port: port}
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
func (c *Client) SendQuery(query []byte) ([]byte, error) {
	isIPV4, err := c.isIPV4()
	if err != nil {
		log.Printf("Error determining IP type: %v\n", err)
		return nil, err
	}

	var addr string
	if isIPV4 {
		addr = fmt.Sprintf("%s:%d", c.serverAddress, c.port)
	} else if !isIPV4 {
		addr = fmt.Sprintf("[%s]:%d", c.serverAddress, c.port)
	}
	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Printf("Dial err %v\n", err)
		return nil, err
	}
	defer conn.Close()

	if _, err = conn.Write(query); err != nil {
		log.Printf("Write err %v\n", err)
		return nil, err
	}

	response := make([]byte, 1024)
	lengthOfTheResponse, err := conn.Read(response)
	if err != nil {
		fmt.Printf("Read err %v\n", err)
		os.Exit(-1)
	}

	if !hasTheSameID(query, response) {
		log.Printf("Response doesn't have the same ID of the query q:%v, r:%v\n", query, response)
		return nil, fmt.Errorf("response doesn't have the same ID of the query")
	}

	return response[:lengthOfTheResponse], nil
}
func hasTheSameID(query, response []byte) bool {
	return slices.Equal(query[:2], response[:2])
}
