package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

// For generating random string
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var initial sync.Once

func setupTestEnvoirment() {
	os.Setenv("STAGING", "testing")
	go StartMockServer()
	go main()
	time.Sleep(100 * time.Millisecond)
}

func TestFowarder(t *testing.T) {
	initial.Do(setupTestEnvoirment)

	// Random Number of testcases, max 100
	num := rand.Int() % 100
	Messages := make([]string, 0, num)

	for range num {
		// Random length input, max 2KB.
		length := rand.Int() % (MAX_LENGTH)
		Messages = append(Messages, randomString(length))
	}

	conn, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal("Error connecting to test server", err)
	}
	defer conn.Close()

	for i := 0; i < len(Messages); i++ {
		message := Messages[i]
		response, err := dummyData(conn, message)
		if err != nil {
			t.Fatal("Error in reading and writing to server", err)
		}

		message = "Here you go: " + message
		if res := string(response); res != message {
			t.Error("Expected", message, "\nRecived:", res)
		}
	}
}

func TestOversizedInputs(t *testing.T) {
	initial.Do(setupTestEnvoirment)

	message := randomString(3 * 1024)

	conn, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatal("Error connecting to test server", err)
	}
	defer conn.Close()

	res, err := dummyData(conn, message)
	if err != nil {
		if err != io.EOF {
			t.Fatal("Error reading and writing message", err)
		}
		return
	}

	if string(res) != "Max Lenght exceeded!!" {
		t.Error("Max length error not encountered")
	}

}

func BenchmarkFowarder(b *testing.B) {
	initial.Do(setupTestEnvoirment)
	conn, err := tls.Dial("tcp", "127.0.0.1:8530", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		b.Fatal("Error connecting to test server", err)
	}
	defer conn.Close()

	message := "Here is a DNS message to benchmark"
	b.ResetTimer() // Start timer from here

	for b.Loop() {
		_, err := dummyData(conn, message)
		if err != nil {
			b.Error("Error reading & writing data:", err)
		}
	}
}

func dummyData(conn *tls.Conn, message string) ([]byte, error) {
	data := []byte(message)
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(data)))
	mssg := append(length, data...)

	_, err := conn.Write(mssg)
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(conn, length)
	if err != nil {
		return nil, err
	}

	len := binary.BigEndian.Uint16(length)
	response := make([]byte, int(len))
	_, err = io.ReadFull(conn, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
