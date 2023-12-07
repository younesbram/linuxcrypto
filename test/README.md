# X.509 Signature Verification Server - Testing Guide

## Manual Testing Steps

### Prerequisites

Before running the tests, ensure you have:

- OpenSSL installed for generating RSA key pairs and signatures.
- `go` compiler for building the test client.

### Generate RSA Key Pair

Run the following commands to generate an RSA key pair:

```
openssl genrsa -out private_key.pem 2048
openssl rsa -in private_key.pem -pubout -out public_key.pem
```

Place `public_key.pem` in the server's `certs` directory.

### Create Test Scripts

Create two bash scripts: `script.sh` and `script_with_error.sh`.

`script.sh`:
```
echo "This is a valid script."
# Change as you like for testing purposes
```

`script_with_error.sh`:
```
echo "I am an invalid script".
invalid_command or malicious_script
```

### Generate Signatures

Use the private key to generate signatures for these scripts:

```
openssl dgst -sha256 -sign private_key.pem -out script.sig script.sh
openssl base64 -in script.sig -out script.sig.b64
openssl dgst -sha256 -sign private_key.pem -out script_with_error.sig script_with_error.sh
openssl base64 -in script_with_error.sig -out script_with_error.sig.b64
```

### Run the Server and Test Client
Build and run the server and test client:
```
./build.sh
./test/test.sh

./server
./test_client
```

### Automated Testing (Future work)
Automated testing can streamline the process by generating keys and signatures, and running multiple test scenarios automatically. This requires implementing a script that automates the OpenSSL commands, generates test scripts, and runs the test client against a variety of inputs.

- Generates RSA keys and signatures automatically.
- Creates and manages test scripts dynamically.
- Runs a suite of test cases covering various scenarios.
- Provides detailed output for each test case.
- Integrates with CI/CD tools for ongoing testing.