package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"

	"github.com/miekg/dns"
)

// This function will recive the dns query []byte from a connection
func ResponseReciver(conn net.Conn, limit bool) ([]byte, error) {
	// First read the length of the dns payload
	length := make([]byte, 2)

	/* Used io.ReadFull instead of conn.Read because if the data is less than 2 bytes,
	conn.Read will read less. ReadFull ensures that we get 2 bytes regardless */
	_, err := io.ReadFull(conn, length)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading length", err)
		}
		return nil, err
	}

	/* Reading from left to right is BigEndian byte order
	reading bytes from least significant is little */
	msgLen := binary.BigEndian.Uint16(length)
	if msgLen == 0 {
		return nil, errors.New("Zero Length Message")
	}

	/* 2 byte length means 64KB. We cant allow that much length on each request.
	If data length exceeds this hardcoded length we return an error to user

	Only Upstream is allowed to send 64KB response */
	if limit && msgLen > MAX_LENGTH {
		return nil, errors.New("Max Length exceeded!!")
	}

	data := make([]byte, int(msgLen))
	_, err = io.ReadFull(conn, data)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading data", err)
		}
		return nil, err
	}

	return append(length, data...), nil
}

func SinkholeReply(reply *dns.Msg, dName string, qType uint16) ([]byte, error) {
	switch qType {
	// return sinkhole 0.0.0.0 for TypeA and :: for TypeAAAA
	case dns.TypeA:
		reply.Answer = append(reply.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   dName,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: net.ParseIP("0.0.0.0").To4(),
		})
	case dns.TypeAAAA:
		reply.Answer = append(reply.Answer, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   dName,
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			AAAA: net.ParseIP("::"),
		})
	}

	length := make([]byte, 2)
	res, err := reply.Pack()
	binary.BigEndian.PutUint16(length, uint16(len(res)))

	return append(length, res...), err
}
