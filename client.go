package main

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/miekg/dns"
)

const (
	MAX_LENGTH = 2 * 1024 // 2 KB
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	upstream, err := UpstreamConnection(1, nil)
	if err != nil {
		log.Println("Upstream Unavilable:", err)
		return
	}
	defer upstream.Close()

	for {
		// Collect dns query form a client
		msg, err := ResponseReciver(conn, true)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reciving dns request", err)
			}
			return
		}

		// Check validity of dns query against blocklist
		if err := AuthenticateQuery(conn, msg); err != nil {
			if err != io.EOF {
				log.Println("Error Authenticating Query", err)
			}
			continue
		}

		// Now send the dns query to upstream
		upstream.Write(msg)

		// Now we recive the response from upstream
		res_msg, err := ResponseReciver(upstream, false)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reciving resolved dns request", err)
			}
			return
		}

		// Write response to client
		conn.Write(res_msg)
	}
}

func AuthenticateQuery(conn net.Conn, msg []byte) error {
	m := new(dns.Msg)
	if err := m.Unpack(msg[2:]); err != nil {
		return err
	}

	// Check if domain name is empty
	if len(m.Question) == 0 {
		return errors.New("No Questions in this query")
	}
	domain := m.Question[0]

	// Create a response
	reply := new(dns.Msg).SetReply(m)

	// Check if domain name is in blocklist with trailing .
	if _, ok := blocklist[dns.Fqdn(domain.Name)]; ok &&
		(domain.Qtype == dns.TypeA || domain.Qtype == dns.TypeAAAA) {
		msg, err := SinkholeReply(reply, domain.Name, domain.Qtype)
		if err != nil {
			log.Println("Error while building sinkhole", err)
		}

		conn.Write(msg)
		return io.EOF
	}
	return nil
}
