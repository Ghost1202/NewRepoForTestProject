SHELL := /bin/bash

BUF ?= buf
GO ?= go

PROTOC_GEN_GO_VERSION ?= v1.36.11
PROTOC_GEN_GO_GRPC_VERSION ?= v1.5.1

.PHONY: tools check-tools fmt lint clean generate contracts

tools:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

check-tools:
	@command -v $(BUF) >/dev/null 2>&1 || (echo "buf is not installed" && exit 1)
	@command -v protoc-gen-go >/dev/null 2>&1 || (echo "protoc-gen-go is not installed, run: make tools" && exit 1)
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || (echo "protoc-gen-go-grpc is not installed, run: make tools" && exit 1)

fmt: check-tools
	$(BUF) format -w

lint: check-tools
	$(BUF) lint

clean:
	find gen/go -type f \( -name '*.pb.go' -o -name '*_grpc.pb.go' \) -delete || true

generate: check-tools
	mkdir -p gen/go
	$(BUF) generate
	$(GO) mod tidy

contracts: fmt lint clean generate
