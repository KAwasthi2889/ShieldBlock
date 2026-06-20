package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

var (
	Initial  sync.Once
	Messages = map[string]string{
		"google.com.":      "142.251.43.206",
		"ads.google.com.":  "0.0.0.0",
		"doubleclick.com.": "0.0.0.0",
		"x.com.":           "172.66.0.227",
		"github.com.":      "20.207.73.82",
	}
)

func setupTestEnvoirment() {
	os.Setenv("STAGING", "testing")
	go StartMockServer()
	go main()
	time.Sleep(100 * time.Millisecond)
}

func TestFowarder(t *testing.T) {
	Initial.Do(setupTestEnvoirment)

	server, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal("Error connecting to test server", err)
	}
	defer server.Close()

	for site, ip := range Messages {
		response, err := dummyData(server, site)
		if err != nil {
			t.Fatal("Error in reading and writing to server", err)
		}

		res, err := getIP(response)
		if err != nil {
			t.Error("Error in response", err)
		}

		if res != ip {
			t.Error("Expected", ip, "\nRecived:", res)
		}
	}
}

// func TestOversizedInputs(t *testing.T) {
// 	initial.Do(setupTestEnvoirment)

// 	message := randomString(3 * 1024)

// 	conn, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
// 		InsecureSkipVerify: true,
// 	})
// 	if err != nil {
// 		t.Fatal("Error connecting to test server", err)
// 	}
// 	defer conn.Close()

// 	res, err := dummyData(conn, message)
// 	if err != nil {
// 		if err != io.EOF {
// 			t.Fatal("Error reading and writing message", err)
// 		}
// 		return
// 	}

// 	if string(res) != "Max Lenght exceeded!!" {
// 		t.Error("Max length error not encountered")
// 	}

// }

func BenchmarkFowarder(b *testing.B) {
	Initial.Do(setupTestEnvoirment)

	conn, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		b.Fatal("Error connecting to test server", err)
	}
	defer conn.Close()

	b.ResetTimer() // Start timer from here

	for b.Loop() {
		_, err := dummyData(conn, "google.com")
		if err != nil {
			b.Error("Error reading & writing data:", err)
		}
	}
}

func dummyData(conn *tls.Conn, site string) ([]byte, error) {
	length := make([]byte, 2)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(site), dns.TypeA)

	data, err := m.Pack()
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint16(length, uint16(len(data)))
	mssg := append(length, data...)

	_, err = conn.Write(mssg)
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(conn, length)
	if err != nil {
		return nil, err
	}

	msgLen := binary.BigEndian.Uint16(length)
	response := make([]byte, int(msgLen))
	_, err = io.ReadFull(conn, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func getIP(response []byte) (string, error) {
	m := new(dns.Msg)
	if err := m.Unpack(response); err != nil {
		return "", err
	}

	var res string
	for _, rr := range m.Answer {
		if a, ok := rr.(*dns.A); ok {
			res = a.A.String()
		}
	}
	return res, nil
}
