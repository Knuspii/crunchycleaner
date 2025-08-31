#!/bin/bash

echo "Building CrunchyCleaner for Windows amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ../crunchycleaner.exe CrunchyCleaner.go
if [ $? -ne 0 ]; then
    echo "Windows build failed"
    exit 1
fi

echo "Building CrunchyCleaner for Linux amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ../crunchycleaner CrunchyCleaner.go
if [ $? -ne 0 ]; then
    echo "Linux build failed"
    exit 1
fi

echo "Build finished successfully."
exit 0
