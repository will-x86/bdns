package server

import (
	"bytes"
	"errors"
	"testing"

	dns "codeberg.org/miekg/dns"
)

type fakeUpstream struct {
	response []byte
	err      error
}

func (f *fakeUpstream) SendQuery(_ []byte) ([]byte, error) {
	return f.response, f.err
}

// counts the calls
type countingUpstream struct {
	calls    int
	response []byte
	err      error
}

func (c *countingUpstream) SendQuery(_ []byte) ([]byte, error) {
	c.calls++
	return c.response, c.err
}

// buildQuery constructs a DNS message for name & type, returns wire bytes.
func buildQuery(t *testing.T, name string, qtype uint16) []byte {
	t.Helper()
	m := dns.NewMsg(name, qtype)
	if m == nil {
		t.Fatalf("buildQuery: NewMsg returned nil for name=%q qtype=%d", name, qtype)
	}
	if err := m.Pack(); err != nil {
		t.Fatalf("buildQuery: Pack failed: %v", err)
	}
	out := make([]byte, len(m.Data))
	copy(out, m.Data)
	return out
}

// buildResponse bytes, flips QR=1, and returns response wire bytes.
// Unpack() reads from m.Data
func buildResponse(t *testing.T, queryWire []byte) []byte {
	t.Helper()
	var m dns.Msg
	m.Data = queryWire
	if err := m.Unpack(); err != nil {
		t.Fatalf("buildResponse: Unpack failed: %v", err)
	}
	m.Response = true
	if err := m.Pack(); err != nil {
		t.Fatalf("buildResponse: Pack failed: %v", err)
	}
	out := make([]byte, len(m.Data))
	copy(out, m.Data)
	return out
}

// TestHandleDNSClient_HappyPath verifies that a valid query is forwarded
// upstream and the response is passed verbatim to the write function.
func TestHandleDNSClient_HappyPath(t *testing.T) {
	query := buildQuery(t, "google.com", dns.TypeA)
	response := buildResponse(t, query)

	var written []byte
	handleDNSClient(query, &fakeUpstream{response: response}, func(b []byte) error {
		written = b
		return nil
	}, "127.0.0.1:1234")

	if !bytes.Equal(written, response) {
		t.Errorf("written bytes don't match response\ngot:  %x\nwant: %x", written, response)
	}
}

// TestHandleDNSClient_UpstreamError verifies that an upstream failure is
// handled gracefully — write must never be called.
func TestHandleDNSClient_UpstreamError(t *testing.T) {
	query := buildQuery(t, "example.com", dns.TypeAAAA)

	up := &countingUpstream{err: errors.New("upstream timeout")}
	writeCount := 0
	handleDNSClient(query, up, func(b []byte) error {
		writeCount++
		return nil
	}, "127.0.0.1:1234")

	if writeCount != 0 {
		t.Errorf("write should not be called on upstream error, was called %d time(s)", writeCount)
	}
}

// TestHandleDNSClient_ParseError verifies that a malformed DNS message is
func TestHandleDNSClient_ParseError(t *testing.T) {
	// Truncated to 2 bytes — header requires 12, so Parse will return an error.
	garbage := []byte{0xCA, 0xFE}

	up := &countingUpstream{}
	writeCount := 0
	handleDNSClient(garbage, up, func(b []byte) error {
		writeCount++
		return nil
	}, "127.0.0.1:1234")

	if up.calls != 0 {
		t.Errorf("upstream should not be called on parse error, was called %d time(s)", up.calls)
	}
	if writeCount != 0 {
		t.Errorf("write should not be called on parse error, was called %d time(s)", writeCount)
	}
}

// TestHandleDNSClient_WriteError verifies that a write failure is handled with no panic
func TestHandleDNSClient_WriteError(t *testing.T) {
	query := buildQuery(t, "example.org", dns.TypeMX)
	response := buildResponse(t, query)

	handleDNSClient(query, &fakeUpstream{response: response}, func(b []byte) error {
		return errors.New("client disconnected")
	}, "127.0.0.1:1234")
	// pass if no panic
}

func TestHandleDNSClient_QueryTypes(t *testing.T) {
	cases := []struct {
		name  string
		qtype uint16
	}{
		{"google.com", dns.TypeA},
		{"google.com", dns.TypeAAAA},
		{"gmail.com", dns.TypeMX},
		{"example.com", dns.TypeTXT},
		{"www.example.com", dns.TypeCNAME},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/"+dns.TypeToString[tc.qtype], func(t *testing.T) {
			query := buildQuery(t, tc.name, tc.qtype)
			response := buildResponse(t, query)

			var written []byte
			handleDNSClient(query, &fakeUpstream{response: response}, func(b []byte) error {
				written = b
				return nil
			}, "127.0.0.1:1234")

			if !bytes.Equal(written, response) {
				t.Errorf("%s %s: response mismatch", tc.name, dns.TypeToString[tc.qtype])
			}
		})
	}
}
