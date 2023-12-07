package main

import (
    "bufio"
    "crypto"
    "crypto/rsa"
    "crypto/sha256"
    "crypto/x509"
    "encoding/base64"
    "encoding/pem"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
)

const (
    port              = ":8080"
    certsDirectory    = "./certs"
    certExtension     = ".crt"
    maxConcurrentReqs = 10 // Maximum concurrent requests allowed
)

var (
    certificates = make(map[string]*publicKeyInfo)
    requestChan  = make(chan *request, maxConcurrentReqs)
    logger       = log.New(os.Stdout, "[Server] ", log.LstdFlags)
)

type publicKeyInfo struct {
    publicKey *rsa.PublicKey
    keyUsage  x509.KeyUsage
}

type request struct {
    conn       net.Conn
    encodedSig string
    script     string
}

func main() {
    loadPublicKeysFromDir(certsDirectory)

    listener, err := net.Listen("tcp", port)
    if err != nil {
        logger.Fatalf("Error listening on port: %v", err)
    }
    defer listener.Close()

    logger.Println("Server listening on", port)

    for i := 0; i < maxConcurrentReqs; i++ {
        go requestHandler()
    }

    for {
        conn, err := listener.Accept()
        if err != nil {
            logger.Printf("Error accepting connection: %v", err)
            continue
        }

        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)

    encodedSig, err := reader.ReadString('\n')
    if err != nil {
        logger.Printf("Error reading signature: %v", err)
        return
    }

    script, err := ioutil.ReadAll(reader)
    if err != nil {
        logger.Printf("Error reading script: %v", err)
        return
    }

    requestChan <- &request{
        conn:       conn,
        encodedSig: strings.TrimSpace(encodedSig),
        script:     string(script),
    }
}

func loadPublicKeysFromDir(certDir string) {
    files, err := ioutil.ReadDir(certDir)
    if err != nil {
        logger.Fatalf("Error reading certificates directory: %v", err)
    }

    for _, f := range files {
        if filepath.Ext(f.Name()) == certExtension {
            certPath := filepath.Join(certDir, f.Name())
            certData, err := ioutil.ReadFile(certPath)
            if err != nil {
                logger.Printf("Error reading certificate file %s: %v", certPath, err)
                continue
            }

            block, _ := pem.Decode(certData)
            if block == nil {
                logger.Printf("Failed to decode PEM block containing the certificate %s", certPath)
                continue
            }

            cert, err := x509.ParseCertificate(block.Bytes)
            if err != nil {
                logger.Printf("Failed to parse certificate %s: %v", certPath, err)
                continue
            }

            rsaPublicKey, ok := cert.PublicKey.(*rsa.PublicKey)
            if !ok {
                logger.Printf("Certificate public key is not of type RSA in %s", certPath)
                continue
            }

            if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
                logger.Printf("Certificate %s does not have a digital signature key usage", certPath)
                continue
            }

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
    signature, err := base64.StdEncoding.DecodeString(req.encodedSig)
    if err != nil {
        logger.Printf("Error decoding signature: %v", err)
        fmt.Fprintln(req.conn, "Error decoding signature")
        return
    }

    if !validateScript(req.script) {
        logger.Println("Invalid script format")
        fmt.Fprintln(req.conn, "Invalid script format")
        return
    }

    valid := verifySignature(signature, []byte(req.script))
    if !valid {
        logger.Println("Invalid signature")
        fmt.Fprintln(req.conn, "Invalid signature")
        return
    }

    output, err := exec.Command("bash", "-c", req.script).CombinedOutput()
    if err != nil {
        logger.Printf("Error executing script: %v", err)
        fmt.Fprintln(req.conn, "Error executing script")
        return
    }

    fmt.Fprintln(req.conn, "Script executed successfully")
    fmt.Fprintln(req.conn, string(output))
}

func verifySignature(signature, script []byte) bool {
    hashed := sha256.Sum256(script)

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
    */// Add more validation checks as per your requirements
    return true
}
