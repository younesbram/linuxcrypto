package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func TestValidSignature(t *testing.T) {
	// Read the base64-encoded signature and script content from files
	signature, err := ioutil.ReadFile("script.sig.b64")
	if err != nil {
		t.Fatalf("Error reading signature file: %v", err)
	}

	scriptContent, err := ioutil.ReadFile("script.sh")
	if err != nil {
		t.Fatalf("Error reading script file: %v", err)
	}

	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Construct the message with the signature and script content
	message := fmt.Sprintf("%s\n%s", signature, scriptContent)

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Error sending message: %v", err)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	// Verify the response
	expectedResponse := "Script executed successfully"
	if !strings.Contains(string(buf[:n]), expectedResponse) {
		t.Errorf("Unexpected response: %s", string(buf[:n]))
	}
}

func TestInvalidSignature(t *testing.T) {
	// Read the script content from file
	scriptContent, err := ioutil.ReadFile("script.sh")
	if err != nil {
		t.Fatalf("Error reading script file: %v", err)
	}

	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Send an invalid signature and the script content
	invalidSignature := "invalid_signature"
	message := fmt.Sprintf("%s\n%s", invalidSignature, scriptContent)

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Error sending message: %v", err)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	// Verify the response
	expectedResponse := "Invalid signature"
	if !strings.Contains(string(buf[:n]), expectedResponse) {
		t.Errorf("Unexpected response: %s", string(buf[:n]))
	}
}

func TestErrorExecutingScript(t *testing.T) {
	// Read the base64-encoded signature and script content from files
	signature, err := ioutil.ReadFile("script.sig.b64")
	if err != nil {
		t.Fatalf("Error reading signature file: %v", err)
	}

	// Modify the script content to contain an error-causing command
	scriptContent, err = ioutil.ReadFile("script_with_error.sh")
	if err != nil {
		t.Fatalf("Error reading script file: %v", err)
	}

	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Construct the message with the signature and script content
	message := fmt.Sprintf("%s\n%s", signature, scriptContent)

	// Send the message
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Error sending message: %v", err)
	}

	// Read the response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	// Verify the response
	expectedResponse := "Error executing script"
	if !strings.Contains(string(buf[:n]), expectedResponse) {
		t.Errorf("Unexpected response: %s", string(buf[:n]))
	}
}

func TestConcurrentRequests(t *testing.T) {
	// Define the number of concurrent requests
	numRequests := 5

	// Prepare the signature and script content
	signature, err := ioutil.ReadFile("script.sig.b64")
	if err != nil {
		t.Fatalf("Error reading signature file: %v", err)
	}

	scriptContent, err := ioutil.ReadFile("script.sh")
	if err != nil {
		t.Fatalf("Error reading script file: %v", err)
	}

	// Create a channel to collect responses
	responseChan := make(chan string, numRequests)

	// Create and run goroutines for each request
	for i := 0; i < numRequests; i++ {
		go func() {
			// Connect to the server
			conn, err := net.Dial("tcp", "localhost:8080")
			if err != nil {
				t.Errorf("Error connecting to server: %v", err)
				return
			}
			defer conn.Close()

			// Construct the message with the signature and script content
			message := fmt.Sprintf("%s\n%s", signature, scriptContent)

			// Send the message
			_, err = conn.Write([]byte(message))
			if err != nil {
				t.Errorf("Error sending message: %v", err)
				return
			}

			// Read the response
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				t.Errorf("Error reading response: %v", err)
				return
			}

			// Send the response to the channel
			responseChan <- string(buf[:n])
		}()
	}

	// Collect and verify responses
	for i := 0; i < numRequests; i++ {
		response := <-responseChan
		if !strings.Contains(response, "Script executed successfully") {
			t.Errorf("Unexpected response: %s", response)
		}
	}
}