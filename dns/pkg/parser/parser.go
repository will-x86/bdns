package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
)

func Message() DNSMessage {
	return DNSMessage{}
}
func (m *DNSMessage) Parse(buf []byte) error {
	reader := bytes.NewBuffer(buf)
	var questions []Question
	err := binary.Read(reader, binary.BigEndian, &m.Header)
	if err != nil {
		return err
	}
	questions = make([]Question, m.Header.NumQuestions)

	for k := range m.Header.NumQuestions {
		questions[k].QName, err = readDomainName(reader)
		if err != nil {
			return fmt.Errorf("error decoding label %d: %v", k, err)
		}
		log.Printf("Decoded domain name for question %d: %s\n", k, questions[k].QName)
		err = binary.Read(reader, binary.BigEndian, &questions[k].QType)
		if err != nil {
			return fmt.Errorf("error reading Typefor question %d: %v", k, err)
		}
		err = binary.Read(reader, binary.BigEndian, &questions[k].QClass)
		if err != nil {
			return fmt.Errorf("error reading Class for question %d: %v", k, err)
		}
	}
	m.Questions = questions
	return nil
}

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

func readDomainName(requestBuffer *bytes.Buffer) (string, error) {
	var domainName string

	b, err := requestBuffer.ReadByte()

	for ; b != 0 && err == nil; b, err = requestBuffer.ReadByte() {
		labelLength := int(b)
		labelBytes := requestBuffer.Next(labelLength)
		labelName := string(labelBytes)

		if len(domainName) == 0 {
			domainName = labelName
		} else {
			domainName += "." + labelName
		}
	}

	return domainName, err
}
