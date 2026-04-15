BINARY_NAME := kubectl-example
BUILD_DIR := bin
GO_MODULE := github.com/ogormans-deptstack/kubectl-example
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test test-unit test-e2e lint clean install

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/kubectl-example

test: test-unit

test-unit:
	go test -race -count=1 ./pkg/... ./cmd/...

test-e2e:
	go test -race -count=1 -tags=e2e -timeout=10m ./test/e2e/...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR) dist

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/$(BINARY_NAME)
