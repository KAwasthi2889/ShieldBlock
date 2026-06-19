package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

var inital sync.Once

func setupTestEnvoirment() {
	os.Setenv("STAGING", "testing")
	go StartMockServer()
	go main()
}

func TestFowarder(t *testing.T) {
	inital.Do(setupTestEnvoirment)
	time.Sleep(100 * time.Millisecond)
	conn, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal("Error connecting to test server", err)
	}

	MESSAGES := [5]string{
		"This is a test DNS message",
		"This is another test DNS message",
		"This is the 3rd message",
		"And this one too",
		"At Last, The final message",
	}

	EXPECTED := [5]string{
		"Here you go: This is a test DNS message",
		"Here you go: This is another test DNS message",
		"Here you go: This is the 3rd message",
		"Here you go: And this one too",
		"Here you go: At Last, The final message",
	}

	for i := 0; i < 5; i++ {
		data := []byte(MESSAGES[i])
		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(len(data)))
		mssg := append(length, data...)

		_, err = conn.Write(mssg)
		if err != nil {
			t.Fatal("Error Writing test message", err)
		}

		_, err = io.ReadFull(conn, length)
		if err != nil {
			t.Fatal("Error reciving length", err)
		}

		len := binary.BigEndian.Uint16(length)
		response := make([]byte, int(len))
		_, err = io.ReadFull(conn, response)
		if err != nil {
			t.Fatal("Error reading responsee", err)
		}

		if string(response) != EXPECTED[i] {
			t.Error("Expected", EXPECTED[i], "\nRecived:", string(response))
		}
	}
}
