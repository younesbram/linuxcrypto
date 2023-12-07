package main

import (
	"encoding/base64"
	"fmt"
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

    // Construct the message with a fake signature and a simple script
    fakeSignature := base64.StdEncoding.EncodeToString([]byte("fake-signature"))
    scriptContent := `echo "Hello, world!"`
    message := fmt.Sprintf("%s\n%s", fakeSignature, scriptContent)

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
