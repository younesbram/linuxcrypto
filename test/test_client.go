package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"strings"
)

func main() {
	// Run all tests
	tests := []func(*testing.T){
		TestValidSignature,
		TestInvalidSignature,
		TestErrorExecutingScript,
		TestConcurrentRequests,
	}

	for _, test := range tests {
		t := &testing.T{}
		test(t)
		if t.Failed() {
			fmt.Println("Test failed:", t.Name())
		} else {
			fmt.Println("Test passed:", t.Name())
		}
	}
}

// In the future instead of loading files, create files with robust automation techniques for each edge case unit case or integration tests etc. Depends on the context.
func loadFile(fileName string, t *testing.T) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error reading file %s: %v", fileName, err)
	}
	return content
}

func TestValidSignature(t *testing.T) {
    signature := loadFile("good_script.sig.b64", t)
    scriptContent := loadFile("good_script.sh", t)

    err := sendRequestToServer(signature, scriptContent, "Script executed successfully", t)
    if err != nil {
        t.Errorf("sendRequestToServer() failed: %v", err)
    }
}

func TestInvalidSignature(t *testing.T) {
	signature := []byte("invalid_signature")
	scriptContent := loadFile("bad_script.sh", t)
	sendRequestToServer(signature, scriptContent, "Invalid signature", t)
}

func TestErrorExecutingScript(t *testing.T) {
	signature := loadFile("./test/bad_script.sig.b64", t)
	scriptContent := loadFile("bad_script.sh", t) 
	sendRequestToServer(signature, scriptContent, "Error executing script", t)
}

func TestConcurrentRequests(t *testing.T) {
	numRequests := 5
	signature := loadFile("good_script.sig.b64", t)
	scriptContent := loadFile("good_script.sh", t)

	responseChan := make(chan string, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			sendRequestToServer(signature, scriptContent, "Script executed successfully", t)
			responseChan <- "Done"
		}()
	}

	for i := 0; i < numRequests; i++ {
		<-responseChan
	}
}

func sendRequestToServer(signature, scriptContent []byte, expectedResponse string, t *testing.T) error {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        return fmt.Errorf("error connecting to server: %v", err)
    }
    defer conn.Close()

    // Add a newline character to the end of the script content
    scriptContent = append(scriptContent, []byte("\n")...)

    // Send the signature to the server
    _, err = conn.Write(signature)
    if err != nil {
        return fmt.Errorf("error sending signature: %v", err)
    }

    // Send the script content to the server
    _, err = conn.Write(scriptContent)
    if err != nil {
        return fmt.Errorf("error sending script content: %v", err)
    }

    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        return fmt.Errorf("error reading response: %v", err)
    }

    response := string(buf[:n])
    fmt.Printf("Debug: Received response from server:\n%s\n", response) // Debugging statement

    if !strings.Contains(response, expectedResponse) {
        return fmt.Errorf("unexpected response: %s", response)
    }

    return nil
}
