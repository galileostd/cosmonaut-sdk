# cosmonaut-sdk

**gRPC contract and Go helpers for building Cosmonaut plugins.**

This repository contains:

- `proto/plugin/v1/plugin.proto` — the canonical plugin contract
- `go/plugin/v1/` — generated Go code (protobuf + gRPC)
- `go/server/` — base server helpers for implementing a plugin in Go
- `go/client/` — gRPC client used by the control-plane to call plugins
- `docs/plugin-spec.md` — how to implement a plugin in any language

## Usage

### Implementing a plugin (Go)

```bash
go get github.com/galileostd/cosmonaut-sdk
```

```go
import (
    "github.com/galileostd/cosmonaut-sdk/go/server"
    pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
)

type MyPlugin struct {
    server.UnimplementedPlugin
}

// implement Describe, HealthCheck, Execute ...

func main() {
    srv := server.New(":50051", &MyPlugin{})
    srv.Serve(context.Background())
}
```

See [docs/plugin-spec.md](docs/plugin-spec.md) for the full specification.

### Implementing a plugin (other languages)

Use `proto/plugin/v1/plugin.proto` with the protoc compiler for your language.

## Regenerating Go code

```bash
make generate
```

Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc`.

```bash
# macOS
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Versioning

The proto contract follows semantic versioning. The current version is `v1`.
Breaking changes will only happen in a new major version (`plugin/v2`).

## License

Apache 2.0 — see [LICENSE](../cosmonaut/LICENSE).
Part of the [Cosmonaut](https://cosmonaut.galileostd.io) project by [galileostd.io](https://galileostd.io).
