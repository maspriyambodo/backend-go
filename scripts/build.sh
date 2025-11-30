#!/bin/bash

# Build script for adminbe

echo "Building adminbe..."

# Build the binary
go build -o bin/adminbe ./cmd/server

# Check if build succeeded
if [ $? -eq 0 ]; then
    echo "Build successful! Binary created at bin/adminbe"
else
    echo "Build failed!"
    exit 1
fi
