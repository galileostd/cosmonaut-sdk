// Package client provides a gRPC client for calling Cosmonaut plugins.
// Used by the control-plane to communicate with registered plugins.
package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	pluginv1 "github.com/galileostd/cosmonaut-sdk/go/plugin/v1"
)

// Client is a gRPC client for a single plugin instance.
type Client struct {
	endpoint string
	conn     *grpc.ClientConn
	plugin   pluginv1.PluginServiceClient
	health   grpc_health_v1.HealthClient
}

// New creates a new plugin client connected to the given endpoint.
// The endpoint is the gRPC address of the plugin (e.g. "cosmonaut-plugin-trino.cosmonaut:50051").
// The dns:/// prefix is added automatically to ensure the gRPC DNS resolver
// is used, which is required for Kubernetes service discovery.
func New(endpoint string) (*Client, error) {
	// ensure the dns:/// prefix is present so gRPC uses the DNS resolver
	// this is required for K8s service DNS to work correctly
	if !strings.HasPrefix(endpoint, "dns:///") &&
		!strings.HasPrefix(endpoint, "passthrough:///") {
		endpoint = "passthrough:///" + endpoint
	}

	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to plugin at %s: %w", endpoint, err)
	}

	return &Client{
		endpoint: endpoint,
		conn:     conn,
		plugin:   pluginv1.NewPluginServiceClient(conn),
		health:   grpc_health_v1.NewHealthClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Endpoint returns the plugin endpoint this client is connected to.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// IsAlive checks if the plugin gRPC server is reachable via the standard
// gRPC health protocol. This is a lightweight check — it does NOT verify
// the underlying tool (Trino, Spark, etc). Use HealthCheck for that.
func (c *Client) IsAlive(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.health.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return false
	}
	return resp.Status == grpc_health_v1.HealthCheckResponse_SERVING
}

// Describe returns static metadata about the plugin.
func (c *Client) Describe(ctx context.Context) (*pluginv1.DescribeResponse, error) {
	return c.plugin.Describe(ctx, &pluginv1.DescribeRequest{})
}

// HealthCheck verifies the component is reachable and operational.
func (c *Client) HealthCheck(ctx context.Context, component *pluginv1.Component) (*pluginv1.HealthCheckResponse, error) {
	return c.plugin.HealthCheck(ctx, &pluginv1.HealthCheckRequest{
		Component: component,
	})
}

// Execute submits work to the component via the plugin.
func (c *Client) Execute(ctx context.Context, component *pluginv1.Component, action string, payload map[string]string) (*pluginv1.ExecuteResponse, error) {
	return c.plugin.Execute(ctx, &pluginv1.ExecuteRequest{
		Component: component,
		Action:    action,
		Payload:   payload,
	})
}

// GetJob returns the current state of a submitted job.
func (c *Client) GetJob(ctx context.Context, component *pluginv1.Component, jobID string) (*pluginv1.GetJobResponse, error) {
	return c.plugin.GetJob(ctx, &pluginv1.GetJobRequest{
		Component: component,
		JobId:     jobID,
	})
}

// ListJobs returns all jobs known to this component.

func (c *Client) ListJobs(ctx context.Context, component *pluginv1.Component, stateFilter pluginv1.JobState, jobName, jobGroup string, limit, offset int32) (*pluginv1.ListJobsResponse, error) {
	return c.plugin.ListJobs(ctx, &pluginv1.ListJobsRequest{
		Component:   component,
		StateFilter: stateFilter,
		JobName:     jobName,
		JobGroup:    jobGroup,
		Limit:       limit,
		Offset:      offset,
	})
}

// CancelJob attempts to cancel a running job.
func (c *Client) CancelJob(ctx context.Context, component *pluginv1.Component, jobID string) (*pluginv1.CancelJobResponse, error) {
	return c.plugin.CancelJob(ctx, &pluginv1.CancelJobRequest{
		Component: component,
		JobId:     jobID,
	})
}

// GetLogs returns logs for a job.
func (c *Client) GetLogs(ctx context.Context, component *pluginv1.Component, jobID, podName string, tailLines int32) (*pluginv1.GetLogsResponse, error) {
	return c.plugin.GetLogs(ctx, &pluginv1.GetLogsRequest{
		Component: component,
		JobId:     jobID,
		PodName:   podName,
		TailLines: tailLines,
	})
}

// Savepoint triggers a savepoint on a streaming job.
func (c *Client) Savepoint(ctx context.Context, component *pluginv1.Component, jobID, targetPath string, cancelJob bool) (*pluginv1.SavepointResponse, error) {
	return c.plugin.Savepoint(ctx, &pluginv1.SavepointRequest{
		Component:  component,
		JobId:      jobID,
		TargetPath: targetPath,
		CancelJob:  cancelJob,
	})
}

// Restore submits a streaming job from a savepoint.
func (c *Client) Restore(ctx context.Context, component *pluginv1.Component, savepointPath string, payload map[string]string) (*pluginv1.RestoreResponse, error) {
	return c.plugin.Restore(ctx, &pluginv1.RestoreRequest{
		Component:     component,
		SavepointPath: savepointPath,
		Payload:       payload,
	})
}

// GetMetrics returns runtime metrics for a running job.
func (c *Client) GetMetrics(ctx context.Context, component *pluginv1.Component, jobID string) (*pluginv1.GetMetricsResponse, error) {
	return c.plugin.GetMetrics(ctx, &pluginv1.GetMetricsRequest{
		Component: component,
		JobId:     jobID,
	})
}

// GetManifest returns the plugin installation manifest.
// Used by the UI wizard and CLI installer to guide the operator
// through installing the plugin and its dependencies.
func (c *Client) GetManifest(ctx context.Context) (*pluginv1.GetManifestResponse, error) {
	return c.plugin.GetManifest(ctx, &pluginv1.GetManifestRequest{})
}
