package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"

	"github.com/miekg/dns"
)

func StartMockServer() {
	// Load Certs and Keys
	cert, err := tls.LoadX509KeyPair("certs/MyCert.crt", "certs/MyKey.key")
	if err != nil {
		log.Fatal("Unable to load TLS Certificates:", err)
	}

	listener, err := tls.Listen("tcp", ":8531", &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	if err != nil {
		log.Fatal("Unable to Listen:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Difficulty listening:", err)
			break
		}

		go mockHandleConnection(conn)
	}
}

func mockHandleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		mssg, err := ResponseReciver(conn, false)
		if err != nil {
			if err != io.EOF {
				log.Println("Mock: Error reciving dns request", err)
			}
			return
		}

		m := new(dns.Msg)
		if err := m.Unpack(mssg[2:]); err != nil {
			log.Println("Mock: Bad Request", err)
		}

		r := new(dns.Msg).SetReply(m)
		domain := m.Question[0]

		ip, exists := Messages[domain.Name]
		if !exists {
			ip = "1.1.1.1"
		}

		r.Answer = append(r.Answer, &dns.A{
			Hdr: dns.RR_Header{
				Name:   domain.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: net.ParseIP(ip).To4(),
		})

		res, err := r.Pack()
		if err != nil {
			log.Println("Mock: Error packing dns response", err)
			continue
		}

		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(len(res)))
		conn.Write(append(length, res...))
	}
}
