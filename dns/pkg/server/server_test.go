package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"testing"

	dns "codeberg.org/miekg/dns"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

type fakeUpstream struct {
	response []byte
	err      error
}

func (f *fakeUpstream) SendQuery(_ []byte) ([]byte, error) {
	return f.response, f.err
}

type fakeProfileStore struct {
	exists bool
	err    error
}

func (f fakeProfileStore) ProfileExists(_ context.Context, _ string) (bool, error) {
	return f.exists, f.err
}

func profileStores(exists bool) rule.Stores {
	return rule.Stores{Profile: fakeProfileStore{exists: exists}}
}

func passthroughEngine() *rule.Engine {
	return rule.NewEngine()
}

// countingUpstream lets tests assert how many times the upstream was called.
type countingUpstream struct {
	calls    int
	response []byte
	err      error
}

func (c *countingUpstream) SendQuery(_ []byte) ([]byte, error) {
	c.calls++
	return c.response, c.err
}

// DNS query for name & type
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

// flips QR=1 on the query
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

// transaction ID from bytes (first 2 bytes, big-endian).
func txID(b []byte) uint16 {
	return binary.BigEndian.Uint16(b[0:2])
}

// withTxID - copy of the bytes with the transaction ID replaced.
func withTxID(b []byte, id uint16) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	binary.BigEndian.PutUint16(out[0:2], id)
	return out
}

func TestHandle_HappyPath(t *testing.T) {
	query := buildQuery(t, "google.com", dns.TypeA)
	response := buildResponse(t, query)

	var written []byte
	h := &handler{
		upstream:  &fakeUpstream{response: response},
		write:     func(b []byte) error { written = b; return nil },
		engine:    passthroughEngine(),
		stores:    profileStores(true),
		profileID: "testprofile",
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")

	if !bytes.Equal(written, response) {
		t.Errorf("written bytes don't match response\ngot:  %x\nwant: %x", written, response)
	}
}

func TestHandle_UpstreamError_NoWrite(t *testing.T) {
	query := buildQuery(t, "example.com", dns.TypeAAAA)
	up := &countingUpstream{err: errors.New("upstream timeout")}

	writeCount := 0
	h := &handler{
		upstream:  up,
		write:     func(b []byte) error { writeCount++; return nil },
		engine:    passthroughEngine(),
		stores:    profileStores(true),
		profileID: "testprofile",
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")

	if writeCount != 0 {
		t.Errorf("write should not be called on upstream error, was called %d time(s)", writeCount)
	}
}

func TestHandle_ParseError_NeitherUpstreamNorWrite(t *testing.T) {
	garbage := []byte{0xCA, 0xFE} // only 2 bytes; header requires 12

	up := &countingUpstream{}
	writeCount := 0
	h := &handler{
		upstream: up,
		write:    func(b []byte) error { writeCount++; return nil },
	}
	h.handle(context.Background(), garbage, "127.0.0.1:1234")

	if up.calls != 0 {
		t.Errorf("upstream should not be called on parse error, was called %d time(s)", up.calls)
	}
	if writeCount != 0 {
		t.Errorf("write should not be called on parse error, was called %d time(s)", writeCount)
	}
}

func TestHandle_WriteError_NoPanic(t *testing.T) {
	query := buildQuery(t, "example.org", dns.TypeMX)
	response := buildResponse(t, query)

	h := &handler{
		upstream:  &fakeUpstream{response: response},
		write:     func(b []byte) error { return errors.New("client disconnected") },
		engine:    passthroughEngine(),
		stores:    profileStores(true),
		profileID: "testprofile",
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")
	// pass if no panic
}

func TestHandle_QueryTypes(t *testing.T) {
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
			h := &handler{
				upstream:  &fakeUpstream{response: response},
				write:     func(b []byte) error { written = b; return nil },
				engine:    passthroughEngine(),
				stores:    profileStores(true),
				profileID: "testprofile",
			}
			h.handle(context.Background(), query, "127.0.0.1:1234")

			if !bytes.Equal(written, response) {
				t.Errorf("%s %s: response mismatch", tc.name, dns.TypeToString[tc.qtype])
			}
		})
	}
}

func TestHandle_ResponseTxIDMatchesQuery(t *testing.T) {
	query := buildQuery(t, "example.com", dns.TypeA)
	queryID := txID(query)

	// Simulate an upstream that returns a response with a *different* ID
	differentID := queryID + 1
	response := buildResponse(t, query)
	responseWithWrongID := withTxID(response, differentID)

	var written []byte
	h := &handler{
		upstream:  &fakeUpstream{response: responseWithWrongID},
		write:     func(b []byte) error { written = b; return nil },
		engine:    passthroughEngine(),
		stores:    profileStores(true),
		profileID: "testprofile",
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")

	if txID(written) != differentID {
		t.Errorf("no-cache path must not rewrite IDs: got %d, want %d", txID(written), differentID)
	}
}

func TestHandle_UnknownUser_Refused(t *testing.T) {
	query := buildQuery(t, "google.com", dns.TypeA)

	up := &countingUpstream{}
	var written []byte
	h := &handler{
		upstream:  up,
		write:     func(b []byte) error { written = b; return nil },
		engine:    passthroughEngine(),
		stores:    profileStores(false),
		profileID: "unknownprofile",
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")

	if up.calls != 0 {
		t.Errorf("upstream must not be called for unknown user, was called %d time(s)", up.calls)
	}
	if written == nil {
		t.Fatal("expected a REFUSED response, got nil")
	}
	if rcode := written[3] & 0x0F; rcode != 5 {
		t.Errorf("expected RCODE=5 (REFUSED), got %d", rcode)
	}
}

func TestHandle_NilUserStore_Refused(t *testing.T) {
	query := buildQuery(t, "google.com", dns.TypeA)

	up := &countingUpstream{}
	var written []byte
	h := &handler{
		upstream:  up,
		write:     func(b []byte) error { written = b; return nil },
		engine:    passthroughEngine(),
		stores:    rule.Stores{},
		profileID: "someprofile",
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")

	if up.calls != 0 {
		t.Errorf("upstream must not be called with nil userChecker, was called %d time(s)", up.calls)
	}
	if written == nil {
		t.Fatal("expected a REFUSED response, got nil")
	}
	if rcode := written[3] & 0x0F; rcode != 5 {
		t.Errorf("expected RCODE=5 (REFUSED), got %d", rcode)
	}
}

func TestHandle_NoSNI_Refused(t *testing.T) {
	query := buildQuery(t, "google.com", dns.TypeA)

	up := &countingUpstream{}
	var written []byte
	h := &handler{
		upstream:  up,
		write:     func(b []byte) error { written = b; return nil },
		engine:    passthroughEngine(),
		profileID: "", // no SNI
	}
	h.handle(context.Background(), query, "127.0.0.1:1234")

	if up.calls != 0 {
		t.Errorf("upstream must not be called for unauthenticated client, was called %d time(s)", up.calls)
	}
	if written == nil {
		t.Fatal("expected a REFUSED response to be written, got nil")
	}
	// Byte 3 (low flags byte): RCODE is the low 4 bits — must be 5 (REFUSED).
	rcode := written[3] & 0x0F
	if rcode != 5 {
		t.Errorf("expected RCODE=5 (REFUSED), got %d", rcode)
	}
	// Transaction ID must be preserved from the query.
	if txID(written) != txID(query) {
		t.Errorf("REFUSED response must preserve query transaction ID: got %d, want %d", txID(written), txID(query))
	}
}
