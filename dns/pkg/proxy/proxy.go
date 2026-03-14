package proxy

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

const poolSize = 8

type conn struct {
	c *tls.Conn
}

type pool struct {
	mu    sync.Mutex
	conns []*conn
	cfg   *tls.Config
	addr  string
}

func newPool(addr string, cfg *tls.Config) *pool {
	return &pool{addr: addr, cfg: cfg}
}

func (p *pool) get() (*conn, error) {
	p.mu.Lock()
	if n := len(p.conns); n > 0 {
		c := p.conns[n-1]
		p.conns = p.conns[:n-1]
		p.mu.Unlock()
		return c, nil
	}
	p.mu.Unlock()
	return p.dial()
}

func (p *pool) put(c *conn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.conns) >= poolSize {
		c.c.Close()
		return
	}
	p.conns = append(p.conns, c)
}

func (p *pool) discard(c *conn) {
	c.c.Close()
}

func (p *pool) dial() (*conn, error) {
	c, err := tls.Dial("tcp", p.addr, p.cfg)
	if err != nil {
		return nil, fmt.Errorf("dial tls: %w", err)
	}
	return &conn{c: c}, nil
}

type Client struct {
	pool *pool
}

func NewTLSClient(address string, port int, serverName string) *Client {
	ip := net.ParseIP(address)
	var addr string
	if ip != nil && ip.To4() == nil && ip.To16() != nil {
		addr = fmt.Sprintf("[%s]:%d", address, port)
	} else {
		addr = fmt.Sprintf("%s:%d", address, port)
	}
	cfg := &tls.Config{ServerName: serverName}
	return &Client{pool: newPool(addr, cfg)}
}

func (c *Client) SendQuery(query []byte) ([]byte, error) {
	for range 2 {
		cn, err := c.pool.get()
		if err != nil {
			return nil, err
		}

		resp, err := sendOn(cn.c, query)
		if err != nil {
			c.pool.discard(cn)
			continue
		}

		c.pool.put(cn)
		return resp, nil
	}
	return nil, fmt.Errorf("upstream query failed after retry")
}

func sendOn(c *tls.Conn, query []byte) ([]byte, error) {
	prefix := make([]byte, 2)
	binary.BigEndian.PutUint16(prefix, uint16(len(query)))
	if _, err := c.Write(append(prefix, query...)); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	var respLen uint16
	if err := binary.Read(c, binary.BigEndian, &respLen); err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}
	resp := make([]byte, respLen)
	if _, err := io.ReadFull(c, resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(resp) < 2 || resp[0] != query[0] || resp[1] != query[1] {
		return nil, fmt.Errorf("response ID mismatch")
	}
	return resp, nil
}
