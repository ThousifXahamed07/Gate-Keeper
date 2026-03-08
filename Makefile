.DEFAULT_GOAL := all

.PHONY: all build test lint clean install check-deps

all: lint test build

check-deps:
	bash scripts/check-deps.sh

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
