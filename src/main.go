package main

import (
    "crypto"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "encoding/base64"
    "encoding/pem"
    "bufio"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

// Constants for server config
const (
    port              = ":8080"
    certsDirectory    = "./certs"
    maxConcurrentReqs = 10 // Maximum concurrent requests allowed
)

// Global variables
var (
    certificates = make(map[string]*publicKeyInfo) // Map to store certificates
    requestChan  = make(chan *request, maxConcurrentReqs) // Channel for requests
    logger       = log.New(os.Stdout, "[Server] ", log.LstdFlags) // Logger
)

// Struct for storing public key information
type publicKeyInfo struct {
    publicKey *rsa.PublicKey
    keyUsage  x509.KeyUsage
}

// Struct for request data
type request struct {
    conn       net.Conn
    encodedSig string
    script     string
}

func main() {
    loadPublicKeysFromDir(certsDirectory)

    // Start listening on the specified TCP port
    listener, err := net.Listen("tcp", port)
    if err != nil {
        logger.Fatalf("Error listening on port: %v", err)
    }
    defer listener.Close()

    logger.Println("Server listening on", port)

    // Start the request handler goroutines
    for i := 0; i < maxConcurrentReqs; i++ {
        go requestHandler()
    }

    // Accept connections and handle them
    for {
        conn, err := listener.Accept()
        if err != nil {
            logger.Printf("Error accepting connection: %v", err)
            continue
        }

        // Handle each connection in a separate goroutine
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    // Ensure connection is closed after handling
    defer conn.Close()

    reader := bufio.NewReader(conn)

    // Read encoded signature from the connection
    encodedSig, err := reader.ReadString('\n')
    if err != nil {
        logger.Printf("Error reading signature: %v", err)
        return
    }

    // Debug: Print the received encoded signature
    fmt.Printf("Debug: Received encoded signature: %s\n", encodedSig)

    // Read the script content
    script, err := ioutil.ReadAll(reader)
    if err != nil {
        logger.Printf("Error reading script: %v", err)
        return
    }

    // Debug: Print the received script
    fmt.Printf("Debug: Received script: %s\n", string(script))


    // Send the request for processing
    requestChan <- &request{
        conn:       conn,
        encodedSig: strings.TrimSpace(encodedSig),
        script:     string(script),
    }
}

func loadPublicKeysFromDir(certDir string) {
    // Read all files in the directory
    files, err := ioutil.ReadDir(certDir)
    if err != nil {
        logger.Fatalf("Error reading certificates directory: %v", err)
    }

    // Iterate over the files and load the public keys
    for _, f := range files {
        ext := filepath.Ext(f.Name())
        if ext == ".pem" || ext == ".crt" {
            // Construct file path and read the certificate
            certPath := filepath.Join(certDir, f.Name())
            certData, err := ioutil.ReadFile(certPath)
            if err != nil {
                logger.Printf("Error reading certificate file %s: %v", certPath, err)
                continue
            }

            // Decode PEM block containing the certificate
            block, _ := pem.Decode(certData)
            if block == nil {
                logger.Printf("Failed to decode PEM block containing the certificate %s", certPath)
                continue
            }

            // Parse the X.509 certificate
            cert, err := x509.ParseCertificate(block.Bytes)
            if err != nil {
                logger.Printf("Failed to parse certificate %s: %v", certPath, err)
                continue
            }

            // Check if the cert is of type RSA
            rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
            if !ok {
                logger.Printf("Certificate public key is not of type RSA in %s", certPath)
                continue
            }

            // Check if the certificate has digital signature key usage
            if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
                logger.Printf("Certificate %s does not have a digital signature key usage", certPath)
                continue
            }

            // Store the public key information
            certificates[f.Name()] = &publicKeyInfo{
                publicKey: rsaPublicKey,
                keyUsage:  cert.KeyUsage,
            }
        }
    }
}


func requestHandler() {
    for req := range requestChan {
        processRequest(req)
    }
}

func processRequest(req *request) {
    // Decode the base64 encoded signature
    signature, err := base64.StdEncoding.DecodeString(req.encodedSig)
    if err != nil {
        logger.Printf("Error decoding signature: %v", err)
        fmt.Fprintln(req.conn, "Error decoding signature")
        return
    }

    // Debug: Print the decoded signature
    fmt.Printf("Debug: Decoded signature: %x\n", signature)


    // Validate the script format
    if !validateScript(req.script) {
        logger.Println("Invalid script format")
        fmt.Fprintln(req.conn, "Invalid script format")
        return
    }

    // Verify the signature
    valid := verifySignature(signature, []byte(req.script))
    if !valid {
        logger.Println("Invalid signature")
        fmt.Fprintln(req.conn, "Invalid signature")
        return
    }
    // Debug: Print the result of the signature verification
    fmt.Printf("Debug: Signature verification result: %v\n", valid)


    // Execute the script if the signature is valid
    output, err := exec.Command("bash", "-c", req.script).CombinedOutput()
    if err != nil {
        logger.Printf("Error executing script: %v", err)
        fmt.Fprintln(req.conn, "Error executing script")
        return
    }

    // Send the script execution output to the client
    fmt.Fprintln(req.conn, "Script executed successfully")
    fmt.Fprintln(req.conn, string(output))
}

// Verify the signature against the stored public keys
func verifySignature(signature, script []byte) bool {
    // Compute SHA-256 hash of the script
    hashed := sha256.Sum256(script)

    // Iterate over the stored public keys and verify the signature
    for _, info := range certificates {
        if info.keyUsage&x509.KeyUsageDigitalSignature != 0 {
            err := rsa.VerifyPKCS1v15(info.publicKey, crypto.SHA256, hashed[:], signature)
            if err == nil {
                return true
            }
        }
    }

    return false
}

// This function should reflect compliance with security standards such as CIS benchmarks.
func validateScript(script string) bool {
    // Example: Limit script length
    /*if len(script) > 1024 {
        logger.Println("Script length exceeds the allowed limit")
        return false
    }

    // Example: Disallow certain dangerous commands
    disallowedCommands := []string{"rm ", "dd ", "shutdown "}
    for _, cmd := range disallowedCommands {
        if strings.Contains(script, cmd) {
            logger.Printf("Script contains disallowed command: %s", cmd)
            return false
        }
    }

    // Example: Character whitelisting (allow only alphanumeric and specific symbols)
    if !regexp.MustCompile(`^[a-zA-Z0-9\s\.,_\/\(\)-]*$`).MatchString(script) {
        logger.Println("Script contains invalid characters")
        return false
    }
	// can use regular expressions too.
    */// Add more validation checks as per requirements
    // Shellcode injection, file system manipulation, etc.
    return true
}
