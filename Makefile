# CommandPM — Makefile
# Usage: make <command>

.PHONY: fmt lint build run tidy check

## fmt: Format all Go files in place
fmt:
	gofmt -w .

## lint: Run the linter
lint:
	golangci-lint run ./...

## build: Compile the binary
build:
	go build -o bin/commandpm .

## run: Run the app directly (no binary)
run:
	go run main.go

## tidy: Clean up go.mod and go.sum
tidy:
	go mod tidy

## check: Format + lint + build in one shot (run this before committing)
check: fmt lint build
