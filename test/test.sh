#!/bin/bash

# Build the test client
go build -o test/test_client test/test_client.go

echo "Build complete. Run the test client binary with ./test/test_client"


#chmod +x test/test.sh