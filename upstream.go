package main

import (
	"crypto/tls"
	"os"
	"time"
)

func UpstreamConnection(tries int, err error) (*tls.Conn, error) {
	if tries > 5 {
		return nil, err
	}

	upstream, err := tls.Dial("tcp", os.Getenv("Upstream"), &tls.Config{
		ServerName:         os.Getenv("UpstreamName"),
		InsecureSkipVerify: os.Getenv("STAGING") != "production",
	})
	if err != nil {
		time.Sleep(2 * time.Second)
		return UpstreamConnection(tries+1, err)
	}
	return upstream, nil
}
