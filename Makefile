.PHONY: build test lint install clean

build:
	go build -o bin/livefpl-pp-cli ./cmd/livefpl-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/livefpl-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/livefpl-pp-mcp ./cmd/livefpl-pp-mcp

install-mcp:
	go install ./cmd/livefpl-pp-mcp

build-all: build build-mcp
