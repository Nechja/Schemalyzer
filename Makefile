.PHONY: build test clean run-tests lint

# Build the binary
build:
	go build -o schemalyzer cmd/schemalyzer/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f schemalyzer
	rm -f coverage.out coverage.html
	rm -f *.json *.yaml

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Install the binary
install:
	go install ./cmd/schemalyzer

# Build for multiple platforms
build-all:
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/schemalyzer-linux-amd64 cmd/schemalyzer/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/schemalyzer-linux-arm64 cmd/schemalyzer/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/schemalyzer-darwin-amd64 cmd/schemalyzer/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/schemalyzer-darwin-arm64 cmd/schemalyzer/main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/schemalyzer-windows-amd64.exe cmd/schemalyzer/main.go
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/schemalyzer-windows-arm64.exe cmd/schemalyzer/main.go
	@echo "All builds complete. Binaries in dist/"