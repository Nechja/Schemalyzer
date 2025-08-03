#!/bin/bash

# Build script for creating standalone binaries

VERSION="1.0.0"
BUILD_DIR="dist"

# Clean previous builds
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

echo "Building SchemaR v$VERSION standalone binaries..."

# Build for Linux AMD64
echo "Building Linux AMD64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o $BUILD_DIR/schemar-linux-amd64 cmd/schemar/main.go

# Build for Linux ARM64
echo "Building Linux ARM64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$VERSION" -o $BUILD_DIR/schemar-linux-arm64 cmd/schemar/main.go

# Build for macOS AMD64
echo "Building macOS AMD64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o $BUILD_DIR/schemar-darwin-amd64 cmd/schemar/main.go

# Build for macOS ARM64 (Apple Silicon)
echo "Building macOS ARM64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$VERSION" -o $BUILD_DIR/schemar-darwin-arm64 cmd/schemar/main.go

# Build for Windows AMD64
echo "Building Windows AMD64..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$VERSION" -o $BUILD_DIR/schemar-windows-amd64.exe cmd/schemar/main.go

# Create archives
echo "Creating archives..."
cd $BUILD_DIR

# Linux
tar -czf schemar-${VERSION}-linux-amd64.tar.gz schemar-linux-amd64
tar -czf schemar-${VERSION}-linux-arm64.tar.gz schemar-linux-arm64

# macOS
tar -czf schemar-${VERSION}-darwin-amd64.tar.gz schemar-darwin-amd64
tar -czf schemar-${VERSION}-darwin-arm64.tar.gz schemar-darwin-arm64

# Windows
zip schemar-${VERSION}-windows-amd64.zip schemar-windows-amd64.exe

cd ..

# Create checksums
echo "Creating checksums..."
cd $BUILD_DIR
sha256sum *.tar.gz *.zip > checksums.txt
cd ..

echo "Build complete! Binaries are in the $BUILD_DIR directory."
echo ""
echo "File sizes:"
ls -lh $BUILD_DIR/schemar-*