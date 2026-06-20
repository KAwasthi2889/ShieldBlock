package main

import (
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
		msg, err := ResponseReciver(conn, true)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reciving dns request", err)
			}
			return
		}

		m := new(dns.Msg)
		if err := m.Unpack(msg[2:]); err != nil {
			log.Println("Bad Request:", err)
			return
		}

		// Check if domain name is empty
		if len(m.Question) == 0 {
			log.Println("No Questions in this query")
			continue
		}
		domain := m.Question[0]

		// Create a response
		reply := new(dns.Msg).SetReply(m)

		// Cheak if domain name is in blocklist with trailing .
		if _, ok := blocklist[dns.Fqdn(domain.Name)]; ok &&
			(domain.Qtype == dns.TypeA || domain.Qtype == dns.TypeAAAA) {
			msg, err := SinkholeReply(reply, domain.Name, domain.Qtype)
			if err != nil {
				log.Println("Error while building sinkhole", err)
			}

			conn.Write(msg)
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
		conn.Write(res_msg)
	}
}
