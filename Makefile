BINARY_NAME=nerifect
BUILD_DIR=bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X github.com/nerifect/nerifect-cli/internal/cli.Version=$(VERSION)"

.PHONY: build clean test install lint docs docs-build docs-gen

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/nerifect

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

lint:
	golangci-lint run ./...

docs-gen:
	go run ./tools/docgen

docs: docs-gen
	mkdocs serve

docs-build: docs-gen
	mkdocs build --strict
