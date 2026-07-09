package parser

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type DNSMessage struct {
	Header      Header
	Questions   []Question
	Answers     []ResourceRecord
	Authorities []ResourceRecord
	Additionals []ResourceRecord
}

type Header struct {
	TransactionID  uint16
	Flags          uint16
	NumQuestions   uint16
	NumAnswers     uint16
	NumAuthorities uint16
	NumAdditionals uint16
}

type Question struct {
	QName  string
	QType  uint16
	QClass uint16
}

type ResourceRecord struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    []byte
}

// Header flag accessors

func (h Header) QR() bool      { return h.Flags>>15&1 == 1 } // true = response
func (h Header) Opcode() uint8 { return uint8(h.Flags >> 11 & 0xF) }
func (h Header) AA() bool      { return h.Flags>>10&1 == 1 }
func (h Header) TC() bool      { return h.Flags>>9&1 == 1 }
func (h Header) RD() bool      { return h.Flags>>8&1 == 1 }
func (h Header) RA() bool      { return h.Flags>>7&1 == 1 }
func (h Header) RCode() uint8  { return uint8(h.Flags & 0xF) }

var opcodeNames = map[uint8]string{
	0: "QUERY", 2: "STATUS", 4: "NOTIFY", 5: "UPDATE", 6: "DSO",
}

var rcodeNames = map[uint8]string{
	0: "NOERROR", 1: "FORMERR", 2: "SERVFAIL", 3: "NXDOMAIN",
	4: "NOTIMP", 5: "REFUSED",
}

func (h Header) OpcodeString() string {
	if s, ok := opcodeNames[h.Opcode()]; ok {
		return s
	}
	return fmt.Sprintf("OPCODE%d", h.Opcode())
}

func (h Header) RCodeString() string {
	if s, ok := rcodeNames[h.RCode()]; ok {
		return s
	}
	return fmt.Sprintf("RCODE%d", h.RCode())
}

// Type / Class names

var TypeToString = map[uint16]string{
	1: "A", 2: "NS", 5: "CNAME", 6: "SOA", 12: "PTR",
	15: "MX", 16: "TXT", 28: "AAAA", 33: "SRV", 255: "ANY",
}

var ClassToString = map[uint16]string{
	1: "IN", 3: "CH", 4: "HS", 255: "ANY",
}

func typeName(t uint16) string {
	if s, ok := TypeToString[t]; ok {
		return s
	}
	return fmt.Sprintf("TYPE%d", t)
}

func className(c uint16) string {
	if s, ok := ClassToString[c]; ok {
		return s
	}
	return fmt.Sprintf("CLASS%d", c)
}

// RDataString returns pretty print representation of the RData field.
// raw bytes on RData directly.
func (rr ResourceRecord) RDataString() string {
	switch rr.Type {
	case 1: // A — 4-byte IPv4
		if len(rr.RData) == 4 {
			return net.IP(rr.RData).String()
		}
	case 28: // AAAA — 16-byte IPv6
		if len(rr.RData) == 16 {
			return net.IP(rr.RData).String()
		}
	case 2, 5, 12: // NS, CNAME, PTR — encoded domain name
		name, _, err := readDomainName(rr.RData, 0)
		if err == nil {
			return name
		}
	case 15: // MX — uint16 preference + encoded domain name
		if len(rr.RData) > 2 {
			pref := binary.BigEndian.Uint16(rr.RData[:2])
			name, _, err := readDomainName(rr.RData, 2)
			if err == nil {
				return fmt.Sprintf("%d %s", pref, name)
			}
		}
	case 16: // TXT — one or more length-prefixed strings
		var parts []string
		data := rr.RData
		for len(data) > 0 {
			l := int(data[0])
			if 1+l > len(data) {
				break
			}
			parts = append(parts, string(data[1:1+l]))
			data = data[1+l:]
		}
		return strings.Join(parts, " ")
	}
	return fmt.Sprintf("%x", rr.RData)
}

// Typed accessors

func (rr ResourceRecord) IPv4() net.IP {
	if rr.Type == 1 && len(rr.RData) == 4 {
		return net.IP(rr.RData)
	}
	return nil
}

func (rr ResourceRecord) IPv6() net.IP {
	if rr.Type == 28 && len(rr.RData) == 16 {
		return net.IP(rr.RData)
	}
	return nil
}

// For NS, CNAME, PTR types, return the decoded domain name. For MX, return the domain part (without preference).
func (rr ResourceRecord) DomainName() string {
	if rr.Type == 2 || rr.Type == 5 || rr.Type == 12 {
		name, _, err := readDomainName(rr.RData, 0)
		if err == nil {
			return name
		}
	}
	return ""
}

// Pretty format Header
func (h Header) String() string {
	qr := "QUERY"
	if h.QR() {
		qr = "RESPONSE"
	}
	flags := []string{}
	if h.AA() {
		flags = append(flags, "AA")
	}
	if h.TC() {
		flags = append(flags, "TC")
	}
	if h.RD() {
		flags = append(flags, "RD")
	}
	if h.RA() {
		flags = append(flags, "RA")
	}
	flagStr := ""
	if len(flags) > 0 {
		flagStr = " [" + strings.Join(flags, " ") + "]"
	}
	return fmt.Sprintf("id=%d %s op=%s%s rcode=%s qdcount=%d ancount=%d nscount=%d arcount=%d",
		h.TransactionID, qr, h.OpcodeString(), flagStr, h.RCodeString(),
		h.NumQuestions, h.NumAnswers, h.NumAuthorities, h.NumAdditionals)
}

// Pretty format for question
func (q Question) String() string {
	return fmt.Sprintf("%s %s %s", q.QName, className(q.QClass), typeName(q.QType))
}

// Pretty format for RR
func (rr ResourceRecord) String() string {
	return fmt.Sprintf("%s %s %s ttl=%d %s",
		rr.Name, className(rr.Class), typeName(rr.Type), rr.TTL, rr.RDataString())
}

// Pretty print message similar to dig
func (m *DNSMessage) String() string {
	var sb strings.Builder
	fmt.Fprintln(&sb, ";; HEADER")
	fmt.Fprintln(&sb, " ", m.Header.String())
	if len(m.Questions) > 0 {
		fmt.Fprintln(&sb, ";; QUESTION")
		for _, q := range m.Questions {
			fmt.Fprintln(&sb, " ", q.String())
		}
	}
	printSection := func(title string, rrs []ResourceRecord) {
		if len(rrs) == 0 {
			return
		}
		fmt.Fprintf(&sb, ";; %s\n", title)
		for _, rr := range rrs {
			fmt.Fprintln(&sb, " ", rr.String())
		}
	}
	printSection("ANSWER", m.Answers)
	printSection("AUTHORITY", m.Authorities)
	printSection("ADDITIONAL", m.Additionals)
	return sb.String()
}

// Parsing functionality
// Parses both queries & responses
func Message() DNSMessage {
	return DNSMessage{}
}

func (m *DNSMessage) Parse(buf []byte) error {
	if len(buf) < 12 {
		return fmt.Errorf("buffer too short for header")
	}
	m.Header.TransactionID = binary.BigEndian.Uint16(buf[0:2])
	m.Header.Flags = binary.BigEndian.Uint16(buf[2:4])
	m.Header.NumQuestions = binary.BigEndian.Uint16(buf[4:6])
	m.Header.NumAnswers = binary.BigEndian.Uint16(buf[6:8])
	m.Header.NumAuthorities = binary.BigEndian.Uint16(buf[8:10])
	m.Header.NumAdditionals = binary.BigEndian.Uint16(buf[10:12])
	offset := 12

	m.Questions = make([]Question, m.Header.NumQuestions)
	for k := range m.Header.NumQuestions {
		var err error
		m.Questions[k].QName, offset, err = readDomainName(buf, offset)
		if err != nil {
			return fmt.Errorf("question %d name: %w", k, err)
		}
		if offset+4 > len(buf) {
			return fmt.Errorf("question %d: buffer too short for type/class", k)
		}
		m.Questions[k].QType = binary.BigEndian.Uint16(buf[offset:])
		m.Questions[k].QClass = binary.BigEndian.Uint16(buf[offset+2:])
		offset += 4
	}

	readRRs := func(count uint16) ([]ResourceRecord, error) {
		rrs := make([]ResourceRecord, count)
		for k := range count {
			var err error
			rrs[k].Name, offset, err = readDomainName(buf, offset)
			if err != nil {
				return nil, fmt.Errorf("rr %d name: %w", k, err)
			}
			if offset+10 > len(buf) {
				return nil, fmt.Errorf("rr %d: buffer too short for fixed fields", k)
			}
			rrs[k].Type = binary.BigEndian.Uint16(buf[offset:])
			rrs[k].Class = binary.BigEndian.Uint16(buf[offset+2:])
			rrs[k].TTL = binary.BigEndian.Uint32(buf[offset+4:])
			rrs[k].RDLength = binary.BigEndian.Uint16(buf[offset+8:])
			offset += 10
			end := offset + int(rrs[k].RDLength)
			if end > len(buf) {
				return nil, fmt.Errorf("rr %d: rdata exceeds buffer", k)
			}
			rrs[k].RData = buf[offset:end]
			offset = end
		}
		return rrs, nil
	}

	var err error
	if m.Answers, err = readRRs(m.Header.NumAnswers); err != nil {
		return err
	}
	if m.Authorities, err = readRRs(m.Header.NumAuthorities); err != nil {
		return err
	}
	if m.Additionals, err = readRRs(m.Header.NumAdditionals); err != nil {
		return err
	}
	return nil
}

// Readining domain name compression
func readDomainName(buf []byte, offset int) (string, int, error) {
	// visited is shared across pointer recursion so a self-referential or
	// cyclic compression pointer is caught instead of recursing forever.
	return readDomainNameVisited(buf, offset, make(map[int]bool))
}

func readDomainNameVisited(buf []byte, offset int, visited map[int]bool) (string, int, error) {
	var parts []string

	for {
		if offset >= len(buf) {
			return "", offset, fmt.Errorf("offset out of bounds")
		}
		if visited[offset] {
			return "", offset, fmt.Errorf("compression pointer loop")
		}
		visited[offset] = true

		length := int(buf[offset])

		if length == 0 {
			offset++
			break
		}

		if length&0xC0 == 0xC0 { // compression pointer
			if offset+1 >= len(buf) {
				return "", offset, fmt.Errorf("truncated compression pointer")
			}
			ptr := int(binary.BigEndian.Uint16([]byte{buf[offset] & 0x3F, buf[offset+1]}))
			suffix, _, err := readDomainNameVisited(buf, ptr, visited)
			if err != nil {
				return "", offset, err
			}
			parts = append(parts, suffix)
			offset += 2
			break
		}

		offset++
		if offset+length > len(buf) {
			return "", offset, fmt.Errorf("label exceeds buffer")
		}
		parts = append(parts, string(buf[offset:offset+length]))
		offset += length
	}

	return strings.Join(parts, "."), offset, nil
}
