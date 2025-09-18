SHELL := /bin/sh

.PHONY: tidy build test cover vet fmt example

tidy:
	go mod tidy

build:
	go build ./...

test:
	go test ./... -race -cover -coverprofile=coverage.out

cover:
	go tool cover -html=coverage.out

vet:
	go vet ./...

fmt:
	@echo "Formatting..."
	@gofmt -s -w .
	@goimports -w . || true

example:
	go run ./examples/basic
