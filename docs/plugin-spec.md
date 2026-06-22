# Cosmonaut Plugin Specification

This document describes how to implement a Cosmonaut plugin in any language.

## Overview

A Cosmonaut plugin is a gRPC service that implements the `PluginService` defined in
`proto/plugin/v1/plugin.proto`. The control-plane discovers plugins automatically via
Kubernetes Service labels and communicates with them over gRPC.

## Discovery

Deploy your plugin as a Kubernetes Service with the following label:

```yaml
metadata:
  labels:
    cosmonaut.galileostd.io/plugin: "true"
```

The control-plane watches all Services with this label across all namespaces
and registers them automatically. No manual configuration required.

## Required RPCs

Every plugin must implement:

| RPC | Description |
|---|---|
| `Describe` | Returns static metadata: name, version, type, capabilities |
| `HealthCheck` | Verifies the underlying tool is reachable |
| `Execute` | Submits work — always returns immediately with a `job_id` |

## Optional RPCs

Implement based on your declared capabilities:

| RPC | When to implement |
|---|---|
| `GetJob` | Any plugin that submits work |
| `ListJobs` | Any plugin that submits work |
| `CancelJob` | Any plugin that submits work |
| `GetLogs` | Workload plugins (Spark, Flink, User Pod) |
| `Savepoint` | `WORKLOAD_TYPE_STREAMING` only |
| `Restore` | `WORKLOAD_TYPE_STREAMING` only |
| `GetMetrics` | Any plugin with runtime metrics |

## ExecutionType

Declare the correct `ExecutionType` in your `Describe` response:

### `EXECUTION_TYPE_WORKLOAD`
Your plugin creates pods or CRDs in the cluster.
Examples: Spark (`SparkApplication`), Flink (`FlinkDeployment`), User Pod (`Job`/`Deployment`).

Must also declare `WorkloadType`:
- `WORKLOAD_TYPE_BATCH` — job runs, completes, terminates
- `WORKLOAD_TYPE_STREAMING` — job runs indefinitely; implement `Savepoint`/`Restore`

### `EXECUTION_TYPE_QUERY`
Your plugin sends work to a running service that processes it internally.
Examples: Trino, Dremio.
Does NOT create pods.

### `EXECUTION_TYPE_OBSERVABILITY`
Your plugin reads state from external systems.
Examples: Airflow (reads DAG runs).

## Implementing in Go

Use the SDK helpers:

```go
import (
    "github.com/galileostd/cosmonaut-sdk/go/server"
    pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
)

type MyPlugin struct {
    server.UnimplementedPlugin
}

func (p *MyPlugin) Describe(ctx context.Context, req *pluginv1.DescribeRequest) (*pluginv1.DescribeResponse, error) {
    return &pluginv1.DescribeResponse{
        PluginName:    "my-plugin",
        DisplayName:   "My Plugin",
        Version:       "v0.1.0",
        Description:   "Integrates My Tool with Cosmonaut",
        PluginType:    pluginv1.PluginType_PLUGIN_TYPE_PROCESSING,
        ExecutionType: pluginv1.ExecutionType_EXECUTION_TYPE_WORKLOAD,
        WorkloadType:  pluginv1.WorkloadType_WORKLOAD_TYPE_BATCH,
        Capabilities: []*pluginv1.Capability{
            {Type: "submit-job", Description: "Submit a job"},
            {Type: "list-jobs",  Description: "List all jobs"},
            {Type: "kill-job",   Description: "Cancel a running job"},
            {Type: "get-logs",   Description: "Fetch job logs"},
        },
    }, nil
}

func main() {
    srv := server.New(":50051", &MyPlugin{})
    srv.Serve(context.Background())
}
```

## Implementing in other languages

Generate a gRPC server from `proto/plugin/v1/plugin.proto` using the protoc
compiler for your language, then implement the `PluginService` interface.

The plugin must:
1. Listen on TCP port `50051` (configurable via `COSMONAUT_PLUGIN_PORT`)
2. Expose the standard gRPC health check protocol (`grpc.health.v1.Health`)
3. Deploy with the label `cosmonaut.galileostd.io/plugin: "true"` on its K8s Service

## Kubernetes deployment

Minimum required manifest:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cosmonaut-plugin-my-plugin
  namespace: cosmonaut-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cosmonaut-plugin-my-plugin
  template:
    metadata:
      labels:
        app: cosmonaut-plugin-my-plugin
    spec:
      containers:
      - name: plugin
        image: ghcr.io/your-org/cosmonaut-plugin-my-plugin:v0.1.0
        ports:
        - containerPort: 50051
---
apiVersion: v1
kind: Service
metadata:
  name: cosmonaut-plugin-my-plugin
  namespace: cosmonaut-system
  labels:
    cosmonaut.galileostd.io/plugin: "true"   # required for discovery
spec:
  selector:
    app: cosmonaut-plugin-my-plugin
  ports:
  - port: 50051
    targetPort: 50051
```

## Breaking changes

The proto contract follows semantic versioning. Breaking changes will only
happen in a new major version (`plugin/v2`). The control-plane will support
multiple plugin API versions simultaneously during transition periods.
