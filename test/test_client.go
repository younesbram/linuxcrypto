package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Utility function to load a file
func loadFile(fileName string, t *testing.T) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Error reading file %s: %v", fileName, err)
	}
	return content
}

// Utility function to simulate server interaction
func sendRequestToServer(signature, scriptContent []byte, expectedResponse string, t *testing.T) error {
	conn, err := net.DialTimeout("tcp", "localhost:8080", 10*time.Second)
	if err != nil {
		return fmt.Errorf("error connecting to server: %v", err)
	}
	defer conn.Close()

	// Ensure signature and scriptContent end with newline
	signature = append(signature, '\n')
	scriptContent = append(scriptContent, '\n')

	// Send the signature and script content to the server
	if _, err := conn.Write(signature); err != nil {
		return fmt.Errorf("error sending signature: %v", err)
	}
	if _, err := conn.Write(scriptContent); err != nil {
		return fmt.Errorf("error sending script content: %v", err)
	}

	// Read the server's response
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

// Test case for valid signature
func TestValidSignature(t *testing.T) {
	signature := loadFile("good_script.sig.b64", t)
	scriptContent := loadFile("good_script.sh", t)

	err := sendRequestToServer(signature, scriptContent, "Script executed successfully", t)
	if err != nil {
		t.Errorf("sendRequestToServer() failed: %v", err)
	}
}

// Test case for invalid signature
func TestInvalidSignature(t *testing.T) {
	signature := []byte("invalid_signature")
	scriptContent := loadFile("bad_script.sh", t)

	err := sendRequestToServer(signature, scriptContent, "Invalid signature", t)
	if err != nil {
		t.Errorf("sendRequestToServer() failed: %v", err)
	}
}

// Test case for error executing script
func TestErrorExecutingScript(t *testing.T) {
	signature := loadFile("bad_script.sig.b64", t)
	scriptContent := loadFile("bad_script.sh", t)

	err := sendRequestToServer(signature, scriptContent, "Error executing script", t)
	if err != nil {
		t.Errorf("sendRequestToServer() failed: %v", err)
	}
}

// Test case for concurrent requests
func TestConcurrentRequests(t *testing.T) {
	numRequests := 5
	signature := loadFile("good_script.sig.b64", t)
	scriptContent := loadFile("good_script.sh", t)

	responseChan := make(chan error, numRequests)
	var wg sync.WaitGroup
	wg.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			err := sendRequestToServer(signature, scriptContent, "Script executed successfully", t)
			responseChan <- err
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(responseChan)

	// Check for errors
	for err := range responseChan {
		if err != nil {
			t.Errorf("sendRequestToServer() failed: %v", err)
		}
	}
}

func main() {
	// Initialize logging
	logFile, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v", err)
		return
	}
	defer logFile.Close()

	log := log.New(logFile, "LOG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Run all tests
	tests := []func(*testing.T){
		TestValidSignature,
		TestInvalidSignature,
		TestErrorExecutingScript,
		TestConcurrentRequests,
	}

	for _, test := range tests {
		t := &testing.T{}
		log.Printf("Running test: %s", getFunctionName(test))
		test(t)
		if t.Failed() {
			log.Printf("Test failed: %s", getFunctionName(test))
		} else {
			log.Printf("Test passed: %s", getFunctionName(test))
		}
	}
}

func getFunctionName(i interface{}) string {
	return strings.TrimSuffix(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), "-fm")
}

// Helper function to create a test signature
func createTestSignature(privateKey *rsa.PrivateKey, message []byte) (string, error) {
	hashed := sha256.Sum256(message)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// Helper function to generate RSA keys for testing
func generateRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// Test case for verifying the creation of test signature
func TestCreateTestSignature(t *testing.T) {
	privateKey, _, err := generateRSAKeys()
	if err != nil {
		t.Fatalf("Failed to generate RSA keys: %v", err)
	}

	message := []byte("test message")
	signature, err := createTestSignature(privateKey, message)
	if err != nil {
		t.Fatalf("Failed to create test signature: %v", err)
	}

	fmt.Printf("Test signature: %s\n", signature)
}
