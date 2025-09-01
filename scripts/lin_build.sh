#!/bin/bash

echo "Building CrunchyCleaner for Windows amd64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ../crunchycleaner.exe ../src/.
if [ $? -ne 0 ]; then
    echo "Windows build failed"
    exit 1
fi

echo "Building CrunchyCleaner for Linux amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../crunchycleaner ../src/.
if [ $? -ne 0 ]; then
    echo "Linux build failed"
    exit 1
fi

echo "Build finished successfully."
exit 0
