package main

import (
	"crypto/tls"
	"log"
	"os"

	"github.com/joho/godotenv"
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
