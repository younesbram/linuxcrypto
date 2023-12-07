#!/bin/bash

# Ensure we're in the test directory
cd "$(dirname "$0")"

# Build the test client
go build test_client.go

# Run the test client
./test_client