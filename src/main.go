package main

import (
	"net"
	"os"
	"log"
)

const socketPath = "/tmp/server.sock"

func main() {
	// Remove the old socket file if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove old socket file: %s", err)
	}

	// Listen on the UNIX socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Listener error: %s", err)
	}
	defer listener.Close()

	log.Printf("Listening on %s", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %s", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Here you will handle incoming requests, verify signatures, and execute the scripts
	log.Println("Connection accepted")
}
