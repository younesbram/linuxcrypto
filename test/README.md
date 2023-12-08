# X.509 Signature Verification Server - Testing Guide

## Manual Testing Steps

### Prerequisites

Before running the tests, ensure you have:

- OpenSSL installed for generating RSA key pairs and signatures.
- `go` compiler for building the test client.

### Create a self-sigined X.509 Certificate:

Please check and change `openssl.cnf` for a valid X.509 certificate before running this command.
```
openssl req -new -x509 -nodes -days 4 -config openssl.cnf -out certs/public_cert.pem
```
This command creates a self-signed X.509 certificate (certificate.pem) valid for 4 days.

Make sure `public_cert.pem` in the server's `certs` directory.
Please make sure your private key is kept secret.

### Create Test Scripts

Create two bash scripts: `good_script.sh` and `bad_script.sh`.
Place them in /test/.

`good_script.sh`:
```
echo "This is a valid script."
# Change as you like for testing purposes
```

`bad_script.sh`:
```
echo "I am an invalid script".
invalid_command or malicious_script
```


### Generate Signatures

Use the private key to generate signatures for these scripts:

```
openssl dgst -sha256 -sign private_key.pem -out test/good_script.sig test/good_script.sh
openssl base64 -in test/good_script.sig -out test/good_script.sig.b64
openssl dgst -sha256 -sign private_key.pem -out test/bad_script.sig test/bad_script.sh
openssl base64 -in test/bad_script.sig -out test/bad_script.sig.b64
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