package proxy

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"testing"
	"time"
)

func TestNewTLSClient_AddrFormatting(t *testing.T) {
	tests := []struct {
		name    string
		address string
		port    int
		want    string
	}{
		{"ipv4", "1.1.1.1", 853, "1.1.1.1:853"},
		{"ipv6", "2606:4700:4700::1111", 853, "[2606:4700:4700::1111]:853"},
		{"hostname", "dns.example.com", 853, "dns.example.com:853"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := NewTLSClient(tc.address, tc.port, "sni.example")
			if c.pool.addr != tc.want {
				t.Errorf("addr = %q, want %q", c.pool.addr, tc.want)
			}
			if c.pool.cfg.ServerName != "sni.example" {
				t.Errorf("ServerName = %q, want sni.example", c.pool.cfg.ServerName)
			}
		})
	}
}

// echoServer starts a TLS server that reads a length-prefixed message and
// writes back a length-prefixed reply produced by reply(msg). It returns the
// listen address and a *tls.Config a client can use to trust it.
func echoServer(t *testing.T, reply func([]byte) []byte) (string, *tls.Config) {
	t.Helper()
	cert, pool := genCert(t)
	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				var n uint16
				if err := binary.Read(c, binary.BigEndian, &n); err != nil {
					return
				}
				msg := make([]byte, n)
				if _, err := io.ReadFull(c, msg); err != nil {
					return
				}
				out := reply(msg)
				prefix := make([]byte, 2)
				binary.BigEndian.PutUint16(prefix, uint16(len(out)))
				_, _ = c.Write(append(prefix, out...))
			}(c)
		}
	}()

	return ln.Addr().String(), &tls.Config{ServerName: "localhost", RootCAs: pool}
}

func TestSendQuery_RoundTrip(t *testing.T) {
	addr, cfg := echoServer(t, func(msg []byte) []byte { return msg }) // echo verbatim
	c := &Client{pool: newPool(addr, cfg)}

	query := []byte{0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00}
	resp, err := c.SendQuery(query)
	if err != nil {
		t.Fatalf("SendQuery: %v", err)
	}
	if string(resp) != string(query) {
		t.Errorf("resp = %x, want %x", resp, query)
	}
}

func TestSendQuery_IDMismatch(t *testing.T) {
	// Server flips the transaction ID so every reply mismatches.
	addr, cfg := echoServer(t, func(msg []byte) []byte {
		out := append([]byte(nil), msg...)
		out[0] ^= 0xFF
		return out
	})
	c := &Client{pool: newPool(addr, cfg)}

	if _, err := c.SendQuery([]byte{0x12, 0x34, 0x00, 0x00}); err == nil {
		t.Fatal("want error on persistent ID mismatch, got nil")
	}
}

func TestSendQuery_PooledConnectionReused(t *testing.T) {
	addr, cfg := echoServer(t, func(msg []byte) []byte { return msg })
	c := &Client{pool: newPool(addr, cfg)}

	for i := 0; i < 3; i++ {
		if _, err := c.SendQuery([]byte{0x00, byte(i), 0x00, 0x00}); err != nil {
			t.Fatalf("query %d: %v", i, err)
		}
	}
	// After successful queries the connection should be returned to the pool.
	if got := len(c.pool.conns); got != 1 {
		t.Errorf("pooled conns = %d, want 1", got)
	}
}

func genCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "localhost"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("keypair: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(certPEM) {
		t.Fatal("append cert to pool")
	}
	return cert, pool
}
