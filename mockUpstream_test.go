package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"
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
		mssg, err := dummyResponseReciver(conn)
		if err != nil {
			if err != io.EOF {
				log.Println("Mock: Error reciving dns request", err)
			}
			return
		}

		resolved_mssg := messageParser(mssg)

		length := make([]byte, 2)
		binary.BigEndian.PutUint16(length, uint16(len(resolved_mssg)))
		conn.Write(append(length, resolved_mssg...))
	}
}

func dummyResponseReciver(conn net.Conn) ([]byte, error) {
	length := make([]byte, 2)

	_, err := io.ReadFull(conn, length)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading length", err)
		}
		return nil, err
	}

	len := binary.BigEndian.Uint16(length)
	data := make([]byte, int(len))

	_, err = io.ReadFull(conn, data)
	if err != nil {
		if err != io.EOF {
			log.Println("Error reading data", err)
		}
		return nil, err
	}

	return data, nil
}

func messageParser(data []byte) []byte {
	s := "Here you go: "
	return append([]byte(s), data...)
}
