.PHONY: build build-wasm serve test test-race vet

GOROOT := $(shell go env GOROOT)

build:
	go build ./...

build-wasm:
	GOOS=js GOARCH=wasm go build -o main.wasm .
	cp "$(GOROOT)/misc/wasm/wasm_exec.js" .

serve: build-wasm
	go run tools/serve.go

test:
	go test ./internal/...

test-race:
	go test -race ./internal/...

cover:
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html

vet:
	go vet ./...
