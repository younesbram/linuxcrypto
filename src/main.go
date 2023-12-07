// main.go
package main

import (
	"bufio"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	port            = ":8080"                
	certsDirectory  = "./certs" // dir containing the x509 certificates
	certExtension   = ".crt"
)

var (
	mu sync.Mutex // Mutex to handle concurrent requests
)

// publicKeyInfo holds the public key and the associated metadata
type publicKeyInfo struct {
	publicKey *rsa.PublicKey
	keyUsage  x509.KeyUsage
}

func main() {
	certificates, err := loadPublicKeysFromDir(certsDirectory)
	if err != nil {
		log.Fatalf("Error loading certificates: %v", err)
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Unable to listen on port %s: %v", port, err)
	}
	defer listener.Close()

	log.Printf("Server listening on %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Unable to accept connection: %v", err)
			continue
		}
		go handleConnection(conn, certificates)
	}
}

// handleConnection manages each client connection
func handleConnection(conn net.Conn, certificates []publicKeyInfo) {
	defer conn.Close()

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

	scriptContent, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("Unable to read script: %v", err)
		fmt.Fprintln(conn, "Error reading script")
		return
	}

	valid := false
	for _, cert := range certificates {
		if (cert.keyUsage & x509.KeyUsageDigitalSignature) != 0 {
			if err := verifySignature(cert.publicKey, signature, scriptContent); err == nil {
				valid = true
				break
			}
		}
	}

	if !valid {
		log.Printf("Invalid signature")
		fmt.Fprintln(conn, "Invalid signature")
		return
	}

	output, err := executeScript(scriptContent)
	if err != nil {
		log.Printf("Error executing script: %v", err)
		fmt.Fprintln(conn, "Error executing script")
		return
	}

	fmt.Fprintln(conn, "Script executed successfully")
	fmt.Fprintln(conn, string(output))
}

// loadPublicKeysFromDir loads all public keys and their usage from the certificate files in a directory
func loadPublicKeysFromDir(certDir string) ([]publicKeyInfo, error) {
	var keys []publicKeyInfo
	files, err := ioutil.ReadDir(certDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), certExtension) {
			certPath := filepath.Join(certDir, file.Name())
			certData, err := ioutil.ReadFile(certPath)
			if err != nil {
				log.Printf("Error reading certificate file %s: %v", certPath, err)
				continue
			}

			block, _ := pem.Decode(certData)
			if block == nil {
				log.Printf("Failed to decode PEM block containing the certificate %s", certPath)
				continue
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				log.Printf("Failed to parse certificate %s: %v", certPath, err)
				continue
			}

			rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
			if !ok {
				log.Printf("Certificate public key is not of type RSA in %s", certPath)
				continue
			}

			keys = append(keys, publicKeyInfo{
				publicKey: rsaPublicKey,
				keyUsage:  cert.KeyUsage,
			})
		}
	}

	return keys, nil
}

// verifySignature checks the digital signature of the script content using the provided public key
func verifySignature(publicKey *rsa.PublicKey, signature, scriptContent []byte) error {
	mu.Lock()
	defer mu.Unlock()

	hashed := sha256.Sum256(scriptContent)
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], signature)
}

// executeScript executes the script content in a bash shell and returns its combined output
func executeScript(scriptContent []byte) ([]byte, error) {
	cmd := exec.Command("bash")
	cmd.Stdin = strings.NewReader(string(scriptContent))
	return cmd.CombinedOutput()
}
