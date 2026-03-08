.DEFAULT_GOAL := all

.PHONY: all build test lint clean install

all: lint test build

build:
	go build -o bin/gatekeeper ./cmd/gatekeeper

test:
	go test ./... -v -race -coverprofile=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out

install:
	go install ./cmd/gatekeeper
