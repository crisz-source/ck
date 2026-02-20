#!/bin/bash

VERSION="0.2.0"
OUTPUT_DIR="dist"

mkdir -p $OUTPUT_DIR

echo "Building ck $VERSION..."

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/ck-linux-amd64 .
echo "✓ Linux AMD64"

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o $OUTPUT_DIR/ck-linux-arm64 .
echo "✓ Linux ARM64"

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/ck-darwin-amd64 .
echo "✓ macOS AMD64"

# macOS ARM64 (M1/M2)
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/ck-darwin-arm64 .
echo "✓ macOS ARM64"

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/ck-windows-amd64.exe .
echo "✓ Windows AMD64"

echo ""
echo "Binários gerados em $OUTPUT_DIR/"
ls -lh $OUTPUT_DIR/
