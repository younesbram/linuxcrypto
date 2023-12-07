#!/bin/bash

# Ensure we're in the project directory
cd "src"

# Build the server
go build -o ../server

echo "Build complete. The server binary is located at ../server"


#chmod +x build.sh