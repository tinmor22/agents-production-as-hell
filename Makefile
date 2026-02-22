# Makefile — sets GOROOT to match the Homebrew Go 1.25.5 installation.
# Needed because the system has a GOROOT mismatch between /usr/local/go (1.24.7)
# and Homebrew's go binary (1.25.5).
GOROOT := /opt/homebrew/Cellar/go/1.25.5/libexec
export GOROOT

.PHONY: build test vet tidy clean

build:
	go build -o pipeline ./cmd/pipeline/

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f pipeline
	go clean -cache
