package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Read the base64-encoded signature from file
	signature, err := ioutil.ReadFile("script.sig.b64")
	if err != nil {
		fmt.Println("Error reading signature file:", err)
		os.Exit(1)
	}

	// Read the script content from file
	scriptContent, err := ioutil.ReadFile("script.sh")
	if err != nil {
		fmt.Println("Error reading script file:", err)
		os.Exit(1)
	}

	// Construct the message with the signature and script content
	message := fmt.Sprintf("%s\n%s", signature, scriptContent)

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Error sending message:", err)
		os.Exit(1)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	fmt.Println("Server response:", string(buf[:n]))
}
