# X.509 Signature Verification Server

## Usage

1. Build the server:
    ```sh
    cd src
    go build -o ../server
    ```
2. Run the server:
    ```sh
    ./server
    ```
3. Send a signed script to the server:
    ```sh
    go build -o test_client test_client.go
    ./test_client
    ```

## Communication Protocol

* The server listens on TCP port 8080.
* Each client sends a message consisting of:
    1. Base64-encoded signature of the script content
    2. Script content
* The server responds with:
    1. Status code: This indicates whether the script was executed successfully or not.
    2. Script output (if signature is valid)
    3. Request identifier (for concurrent requests)



Status Codes
200 - OK: Script executed successfully.
400 - Bad Request: Invalid request format.
401 - Unauthorized: Invalid signature.
500 - Internal Server Error: Script execution failure.

## Additional Information

The server currently supports RSA signatures with SHA-256 algorithm.
Public keys for verifying signatures must be placed in the certs directory.
The server is single-threaded and can handle a limited number of concurrent requests.

## Features

* Verifies signatures using X.509 certificates
* Supports concurrent requests
* Identifies outputs for each request
* Checks certificate extension for code signing
* Verifies signatures against multiple certificates

## Testing

The `test_client.go` script demonstrates how to send a signed script and receive the response.


## Future Work

* Improve robustness of error handling
* Harden the server against race conditions, malicious scripts etc. Can use access control/rate limitting/ more logging.
* Add support for additional signature algorithms
* Improve performance by adding support for multiple threads.
* Add features like logging and monitoring.

## Contribution

Contributions are welcome! Please submit pull requests to the GitHub repository.

## License

