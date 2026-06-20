package main

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	MAX_LENGTH = 2 * 1024 // 2 KB
)

func main() {
	// Load Certs and Keys
	cert, err := tls.LoadX509KeyPair("certs/MyCert.crt", "certs/MyKey.key")
	if err != nil {
		log.Fatal("Unable to load TLS Certificates:", err)
	}

	// Load ENV file
	env := os.Getenv("STAGING")
	switch env {
	case "production":
		godotenv.Load(".env.prod")
	default:
		godotenv.Load(".env.dev")
	}

	// Create a TLS Listener
	listener, err := tls.Listen("tcp", ":"+os.Getenv("Port"), &tls.Config{
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

		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	upstream, err := UpstreamConnection(1, nil)
	if err != nil {
		log.Println("Upstream Unavilable:", err)
		return
	}
	defer upstream.Close()

	for {
		mssg, err := ResponseReciver(conn, true)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reciving dns request", err)
			}
			return
		}
		// Now we will send the data upstream, since we are not reading in phase 1
		_, err = upstream.Write(mssg)
		if err != nil {
			log.Println("Unable to write to upstream:", err)
		}

		// Now we recive the response from upstream
		resolved_mssg, err := ResponseReciver(upstream, false)
		if err != nil {
			if err != io.EOF {
				log.Println("Error reciving resolved dns request", err)
			}
			return
		}
		conn.Write(resolved_mssg)
	}
}

func ResponseReciver(conn net.Conn, limit bool) ([]byte, error) {
	length := make([]byte, 2)
	/*
		Used io.ReadFull instead of conn.Read because if the data is less than 2 bytes,
		conn.Read will read less. ReadFull ensures that we get 2 bytes regardless
	*/
	_, err := io.ReadFull(conn, length)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading length", err)
		}
		return nil, err
	}
	/*
		Reading from left to right is BigEndian byte order
		reading bytes from least significant is little
	*/
	len := binary.BigEndian.Uint16(length)
	/*
		2 byte length means 64KB. We cant allow that much length on each request.
		If data length exceeds this hardcoded length we return an error to user

		Only Upstream is allowed to send 64KB response
	*/
	if limit && len > MAX_LENGTH {
		return nil, errors.New("Max Lenght exceeded!!")
	}

	data := make([]byte, int(len))
	_, err = io.ReadFull(conn, data)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading data", err)
		}
		return nil, err
	}

	return append(length, data...), nil
}

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
