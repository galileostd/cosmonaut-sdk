.PHONY: generate lint test clean

PROTO_DIR := proto
GO_OUT    := go

## generate: generate Go code from proto definitions
generate:
	@echo "==> Generating Go code from proto"
	@which protoc > /dev/null || (echo "ERROR: protoc not found. Install protobuf compiler." && exit 1)
	@which protoc-gen-go > /dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@which protoc-gen-go-grpc > /dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/plugin/v1/plugin.proto

## lint: lint proto files
lint:
	@which buf > /dev/null || (echo "Install buf: https://buf.build/docs/installation" && exit 1)
	buf lint

## test: run tests
test:
	go test ./...

## tidy: tidy go modules
tidy:
	go mod tidy

## clean: remove generated files
clean:
	find $(GO_OUT) -name "*.pb.go" -delete

## help: print this help
help:
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
