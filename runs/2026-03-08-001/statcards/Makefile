VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOROOT ?= $(shell go env GOROOT)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
BIN := bin/statcard

.PHONY: build build-linux test lint clean

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN) ./cmd/statcard

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)-linux-amd64 ./cmd/statcard

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
