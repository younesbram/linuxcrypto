package main

import (
	"bufio"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"encoding/base64"
)

const (
	port            = ":8080"              // Port where the server will listen for TCP connections
	certificatePath = "./certs/server.crt" // Path to the x509 certificate
)

func main() {
	// Listen on the specified port for incoming TCP connections
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Unable to listen on port %s: %v", port, err)
	}
	defer listener.Close()

	log.Printf("Server listening on %s", port)

	// Main loop to accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Unable to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the first line for the signature
	reader := bufio.NewReader(conn)
	encodedSig, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Unable to read signature: %v", err)
		fmt.Fprintln(conn, "Error reading signature")
		return
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedSig))
	if err != nil {
		log.Printf("Unable to decode signature: %v", err)
		fmt.Fprintln(conn, "Error decoding signature")
		return
	}

	// Read the rest of the input for the script
	scriptContent, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Unable to read script: %v", err)
		fmt.Fprintln(conn, "Error reading script")
		return
	}

	// Verify the signature
	publicKey, err := extractPublicKey(certificatePath)
	if err != nil {
		log.Printf("Unable to extract public key: %v", err)
		fmt.Fprintln(conn, "Error extracting public key")
		return
	}
	if err := verifySignature(publicKey, signature, scriptContent); err != nil {
		log.Printf("Invalid signature: %v", err)
		fmt.Fprintln(conn, "Invalid signature")
		return
	}

	// Execute the script if the signature is valid
	output, err := executeScript(string(scriptContent))
	if err != nil {
		log.Printf("Error executing script: %v", err)
		fmt.Fprintln(conn, "Error executing script")
		return
	}

	// Send back the script output
	fmt.Fprintln(conn, "Script executed successfully")
	fmt.Fprintln(conn, string(output))
}

func extractPublicKey(certPath string) (*rsa.PublicKey, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("read certificate file: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing the certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %v", err)
	}

	rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("certificate public key is not of type RSA")
	}

	return rsaPublicKey, nil
}

func verifySignature(publicKey *rsa.PublicKey, signature, scriptContent []byte) error {
	hashed := sha256.Sum256(scriptContent)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
}

func executeScript(scriptContent string) ([]byte, error) {
	cmd := exec.Command("bash", "-c", scriptContent)
	return cmd.CombinedOutput()
}
