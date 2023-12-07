#!/bin/bash

# Ensure we're in the project directory
cd "src"

# Build the server
go build -o ../server

echo "Build complete. Run the server binary with ./server"


#chmod +x build.sh