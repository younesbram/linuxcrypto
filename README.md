# X.509 Signature Verification Server

## Usage

0. Give permission to the server build and test build scripts: 
    ```
    chmod +x test/test.sh
    chmod +x build.sh
    ```

1. Build the server using the provided script:
    ```
    ./build.sh
    ```

2. Run the server:
    ```
    ./server
    ```

3. Build the test client using the provided script:
    ```
    ./test/test.sh
    ```

4. Run the test client to send a signed script to the server:
    ```
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

## Status Codes

- 200 - OK: Script executed successfully.
- 400 - Bad Request: Invalid request format.
- 401 - Unauthorized: Invalid signature.
- 500 - Internal Server Error: Script execution failure.

## Additional Information

## Additional Information

- The server currently supports RSA signatures with SHA-256 algorithm.
- Public keys for verifying signatures must be placed in the `certs` directory.
- The server is designed to handle a limited number of concurrent requests.
- Please ensure that the communication between the client and server is encrypted and secure. This could involve using TLS for the TCP connections.
-  Regularly review and update the cryptographic algorithms and key sizes to adhere to current best practices and standards.

## Features

- Verifies signatures using X.509 certificates
- Supports concurrent requests
- Identifies outputs for each request
- Checks certificate extension for code signing
- Verifies signatures against multiple certificates

## Testing

The `test_client.go` script demonstrates how to send a signed script and receive the response.


## Future Work

* Use TLS/SSL for secure communication
* Improve robustness of error handling.
* Harden the server against race conditions, malicious scripts etc rate limitting more logging with unique IDs etc.. by commiting PoCs and further threat modelling.
* Can customize the build scripts to incorporate and make use of access control for more robust production ready code. 
* Add support for additional signature algorithms.
* Improve performance by adding support for multiple threads.
* Implement better logging to capture detailed information about each request and its processing outcome. Integrate with open source monitoring tools to detect and alert on suspicious activities.
* Ensure that the server efficiently manages resources, especially when dealing with concurrent requests. Implement safeguards against resource exhaustion.
* Beyond just loading public keys, function a thorough validation of each certificate. This includes checking the validity period, verifying the certificate chain (if applicable), and ensuring that the certificate is not revoked.

## Contribution

Contributions are welcome! Please submit pull requests to the GitHub repository.

## License

