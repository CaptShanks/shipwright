BINARY := ship
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
BUILD_DIR := dist

.PHONY: build clean install test lint

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/ship

install:
	go install $(LDFLAGS) ./cmd/ship

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

# Cross-compile for releases
.PHONY: release
release: clean
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 ./cmd/ship
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 ./cmd/ship
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 ./cmd/ship
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 ./cmd/ship
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe ./cmd/ship
