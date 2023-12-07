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

func loadFile(fileName string, t *testing.T) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error reading file %s: %v", fileName, err)
	}
	return content
}

func TestValidSignature(t *testing.T) {
	signature := loadFile("script.sig.b64", t)
	scriptContent := loadFile("script.sh", t)
	sendRequestToServer(signature, scriptContent, "Script executed successfully", t)
}

func TestInvalidSignature(t *testing.T) {
	signature := []byte("invalid_signature")
	scriptContent := loadFile("script.sh", t)
	sendRequestToServer(signature, scriptContent, "Invalid signature", t)
}

func TestErrorExecutingScript(t *testing.T) {
	signature := loadFile("script.sig.b64", t)
	scriptContent := loadFile("script_with_error.sh", t)
	sendRequestToServer(signature, scriptContent, "Error executing script", t)
}

func TestConcurrentRequests(t *testing.T) {
	numRequests := 5
	signature := loadFile("script.sig.b64", t)
	scriptContent := loadFile("script.sh", t)

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

func sendRequestToServer(signature, scriptContent []byte, expectedResponse string, t *testing.T) {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		t.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	message := fmt.Sprintf("%s\n%s", signature, scriptContent)
	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Error sending message: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Error reading response: %v", err)
	}

	if !strings.Contains(string(buf[:n]), expectedResponse) {
		t.Errorf("Unexpected response: %s", string(buf[:n]))
	}
}
