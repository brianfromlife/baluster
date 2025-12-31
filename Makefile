.PHONY: help proto build build-cli test clean web-build web infra release

# Variables
BUF_VERSION := 1.36.0
PROTO_DIR := proto
GEN_GO_DIR := internal/gen
BUF_BIN := $(shell which buf || echo "")
GO_BIN := $(shell which go || echo "")

help:
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

install-buf:
	@if [ -z "$(BUF_BIN)" ]; then \
		echo "Installing buf..."; \
		curl -sSL "https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(shell uname -s)-$(shell uname -m)" -o /tmp/buf && \
		chmod +x /tmp/buf && \
		sudo mv /tmp/buf /usr/local/bin/buf; \
	else \
		echo "buf already installed"; \
	fi

proto: install-buf
	@echo "Generating protocol buffers..."
	@rm -rf $(GEN_GO_DIR)
	@mkdir -p $(GEN_GO_DIR)
	@cd $(PROTO_DIR) && buf generate
	@echo "Protocol buffers generated successfully"

build-rest:
	@echo "Building REST server..."
	@go build -o bin/rest ./cmd/rest
	@echo "REST server built: bin/rest"

build-grpc:
	@echo "Building gRPC server..."
	@go build -o bin/grpc ./cmd/grpc
	@echo "gRPC server built: bin/grpc"

build-cli:
	@echo "Building CLI tool..."
	@go build -o bin/cli ./cmd/cli
	@echo "CLI tool built: bin/cli"

build: clean proto build-rest build-grpc build-cli

test:
	@echo "Running tests..."
	@go test ./...

web-build:
	@echo "Building web application..."
	@cd web && pnpm build

rest:
	@echo "Running REST server..."
	@go run ./cmd/rest

grpc:
	@echo "Running gRPC server..."
	@./bin/grpc

web:
	@cd web && pnpm dev

infra:
	@echo "Deploying infrastructure..."
	@./bin/cli deploy infra
	@echo "Infrastructure deployed successfully"

release:
	@echo "Deploying all applications..."
	@./bin/cli deploy release
	@echo "Deployment completed successfully"

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf $(GEN_GO_DIR)
	@rm -rf web/dist/
	@echo "Clean complete"


