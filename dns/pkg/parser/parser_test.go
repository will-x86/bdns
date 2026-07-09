package parser

import (
	"net"
	"testing"
)

// encodeName builds a length-prefixed DNS domain name terminated by a zero octet.
func encodeName(name string) []byte {
	var out []byte
	if name != "" {
		for _, label := range splitLabels(name) {
			out = append(out, byte(len(label)))
			out = append(out, label...)
		}
	}
	return append(out, 0)
}

func splitLabels(name string) []string {
	var labels []string
	start := 0
	for i := 0; i < len(name); i++ {
		if name[i] == '.' {
			labels = append(labels, name[start:i])
			start = i + 1
		}
	}
	return append(labels, name[start:])
}

// exampleQuery is a well-formed query for "example.com" A IN, id 0x1234, RD set.
func exampleQuery() []byte {
	buf := []byte{
		0x12, 0x34, // id
		0x01, 0x00, // flags: RD
		0x00, 0x01, // qdcount
		0x00, 0x00, // ancount
		0x00, 0x00, // nscount
		0x00, 0x00, // arcount
	}
	buf = append(buf, encodeName("example.com")...)
	buf = append(buf, 0x00, 0x01, 0x00, 0x01) // type A, class IN
	return buf
}

// exampleResponse is exampleQuery answered with A 93.184.216.34, using a
// compression pointer (0xC00C) for the answer name.
func exampleResponse() []byte {
	buf := []byte{
		0x12, 0x34, // id
		0x81, 0x80, // flags: QR, RD, RA
		0x00, 0x01, // qdcount
		0x00, 0x01, // ancount
		0x00, 0x00, // nscount
		0x00, 0x00, // arcount
	}
	buf = append(buf, encodeName("example.com")...)
	buf = append(buf, 0x00, 0x01, 0x00, 0x01) // question type/class
	buf = append(buf,
		0xC0, 0x0C, // name pointer -> offset 12
		0x00, 0x01, // type A
		0x00, 0x01, // class IN
		0x00, 0x00, 0x01, 0x2C, // ttl 300
		0x00, 0x04, // rdlength
		0x5D, 0xB8, 0xD8, 0x22, // 93.184.216.34
	)
	return buf
}

func TestParse_Query(t *testing.T) {
	m := Message()
	if err := m.Parse(exampleQuery()); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Header.TransactionID != 0x1234 {
		t.Errorf("id = %#x, want 0x1234", m.Header.TransactionID)
	}
	if m.Header.QR() {
		t.Error("QR should be false for a query")
	}
	if !m.Header.RD() {
		t.Error("RD should be set")
	}
	if m.Header.NumQuestions != 1 {
		t.Fatalf("qdcount = %d, want 1", m.Header.NumQuestions)
	}
	q := m.Questions[0]
	if q.QName != "example.com" {
		t.Errorf("qname = %q, want example.com", q.QName)
	}
	if q.QType != 1 || q.QClass != 1 {
		t.Errorf("qtype/qclass = %d/%d, want 1/1", q.QType, q.QClass)
	}
}

func TestParse_ResponseWithAnswer(t *testing.T) {
	m := Message()
	if err := m.Parse(exampleResponse()); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !m.Header.QR() || !m.Header.RA() {
		t.Error("QR and RA should be set on the response")
	}
	if m.Header.NumAnswers != 1 {
		t.Fatalf("ancount = %d, want 1", m.Header.NumAnswers)
	}
	a := m.Answers[0]
	if a.Name != "example.com" {
		t.Errorf("answer name = %q, want example.com (compression)", a.Name)
	}
	if a.Type != 1 || a.TTL != 300 {
		t.Errorf("type/ttl = %d/%d, want 1/300", a.Type, a.TTL)
	}
	if ip := a.IPv4(); ip == nil || ip.String() != "93.184.216.34" {
		t.Errorf("IPv4 = %v, want 93.184.216.34", ip)
	}
	if s := a.RDataString(); s != "93.184.216.34" {
		t.Errorf("RDataString = %q, want 93.184.216.34", s)
	}
}

func TestParse_Errors(t *testing.T) {
	tests := map[string][]byte{
		"short header":        {0x12, 0x34, 0x00},
		"truncated question":  append([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}, encodeName("x.com")...),
		"rdata beyond buffer": rdataOverflow(),
	}
	for name, buf := range tests {
		t.Run(name, func(t *testing.T) {
			m := Message()
			if err := m.Parse(buf); err == nil {
				t.Fatal("want parse error, got nil")
			}
		})
	}
}

// rdataOverflow builds a response whose answer claims more rdata than present.
func rdataOverflow() []byte {
	buf := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	buf = append(buf, encodeName("x.com")...)
	buf = append(buf,
		0x00, 0x01, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x3C,
		0x00, 0xFF, // rdlength 255 but no rdata follows
	)
	return buf
}

func TestReadDomainName_Compression(t *testing.T) {
	// header(12) + "com\0" at 12..16 + pointer to 12 at 17
	buf := make([]byte, 12)
	buf = append(buf, encodeName("com")...) // indices 12..16
	buf = append(buf, 0xC0, 0x0C)           // pointer -> 12

	name, _, err := readDomainName(buf, 17)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "com" {
		t.Errorf("name = %q, want com", name)
	}
}

func TestReadDomainName_PointerLoop(t *testing.T) {
	buf := make([]byte, 14)
	buf[12] = 0xC0 // pointer...
	buf[13] = 0x0C // ...to offset 12 (itself)
	if _, _, err := readDomainName(buf, 12); err == nil {
		t.Fatal("want error for self-referential compression pointer")
	}
}

func TestHeaderStrings(t *testing.T) {
	if got := (Header{Flags: 0x0003}).RCodeString(); got != "NXDOMAIN" {
		t.Errorf("RCodeString = %q, want NXDOMAIN", got)
	}
	if got := (Header{Flags: 0x0005}).RCodeString(); got != "REFUSED" {
		t.Errorf("RCodeString = %q, want REFUSED", got)
	}
	if got := (Header{Flags: 0}).OpcodeString(); got != "QUERY" {
		t.Errorf("OpcodeString = %q, want QUERY", got)
	}
	// opcode 5 = UPDATE occupies bits 11-14: 5 << 11 = 0x2800
	if got := (Header{Flags: 0x2800}).OpcodeString(); got != "UPDATE" {
		t.Errorf("OpcodeString = %q, want UPDATE", got)
	}
	// unknown rcode falls back to RCODE<n>
	if got := (Header{Flags: 0x000F}).RCodeString(); got != "RCODE15" {
		t.Errorf("RCodeString = %q, want RCODE15", got)
	}
}

func TestRDataString(t *testing.T) {
	t.Run("CNAME", func(t *testing.T) {
		rr := ResourceRecord{Type: 5, RData: encodeName("cdn.example.net")}
		if got := rr.RDataString(); got != "cdn.example.net" {
			t.Errorf("got %q", got)
		}
		if got := rr.DomainName(); got != "cdn.example.net" {
			t.Errorf("DomainName = %q", got)
		}
	})
	t.Run("AAAA", func(t *testing.T) {
		rr := ResourceRecord{Type: 28, RData: net.ParseIP("2001:db8::1").To16()}
		if got := rr.RDataString(); got != "2001:db8::1" {
			t.Errorf("got %q", got)
		}
		if rr.IPv6() == nil {
			t.Error("IPv6() returned nil")
		}
	})
	t.Run("MX", func(t *testing.T) {
		rr := ResourceRecord{Type: 15, RData: append([]byte{0x00, 0x0A}, encodeName("mail.example.com")...)}
		if got := rr.RDataString(); got != "10 mail.example.com" {
			t.Errorf("got %q", got)
		}
	})
	t.Run("TXT", func(t *testing.T) {
		rr := ResourceRecord{Type: 16, RData: []byte{0x05, 'h', 'e', 'l', 'l', 'o', 0x05, 'w', 'o', 'r', 'l', 'd'}}
		if got := rr.RDataString(); got != "hello world" {
			t.Errorf("got %q", got)
		}
	})
}

func TestTypedAccessors_WrongType(t *testing.T) {
	if ip := (ResourceRecord{Type: 5, RData: []byte{1, 2, 3, 4}}).IPv4(); ip != nil {
		t.Error("IPv4 should be nil for non-A record")
	}
	if ip := (ResourceRecord{Type: 1, RData: []byte{1, 2, 3, 4}}).IPv6(); ip != nil {
		t.Error("IPv6 should be nil for non-AAAA record")
	}
	if dn := (ResourceRecord{Type: 1, RData: []byte{1, 2, 3, 4}}).DomainName(); dn != "" {
		t.Errorf("DomainName should be empty for A record, got %q", dn)
	}
}

func TestMessageString_DoesNotPanic(t *testing.T) {
	m := Message()
	if err := m.Parse(exampleResponse()); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s := m.String(); s == "" {
		t.Error("String() returned empty output")
	}
}
