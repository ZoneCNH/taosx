package taosx

// Metric name constants use the "taosx_" prefix to match existing dashboards.
// They intentionally differ from the observex.MetricClient* generic names
// (which use no prefix), so cannot be aliased to observex.
// See: github.com/ZoneCNH/observex/pkg/observex/metrics.go
const (
MetricClientCreatedTotal           = "taosx_client_created_total"
MetricClientClosedTotal            = "taosx_client_closed_total"
MetricClientErrorsTotal            = "taosx_client_errors_total"
MetricClientHealthStatus           = "taosx_client_health_status"
MetricClientHealthLatencyMS        = "taosx_client_health_latency_ms"
MetricClientRequestsTotal          = "taosx_client_requests_total"
MetricClientRequestDurationSeconds = "taosx_client_request_duration_seconds"
MetricClientRetriesTotal           = "taosx_client_retries_total"
MetricClientInflight               = "taosx_client_inflight"

// taosx-specific metric names.
MetricClientBatchRowsTotal         = "taosx_client_batch_rows_total"
MetricClientSchemalessLinesTotal   = "taosx_client_schemaless_lines_total"
)

// Metrics is the observability hook interface for taosx clients.
// It is a 3-method subset of observex.Metrics; any observex.Metrics
// implementation satisfies this interface.
type Metrics interface {
IncCounter(name string, labels map[string]string)
ObserveHistogram(name string, value float64, labels map[string]string)
SetGauge(name string, value float64, labels map[string]string)
}

// NoopMetrics discards all observations. observex.NoopMetrics also satisfies
// this interface and may be used directly if observex is already a dependency.
type NoopMetrics struct{}

func (NoopMetrics) IncCounter(name string, labels map[string]string)                {}
func (NoopMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {}
func (NoopMetrics) SetGauge(name string, value float64, labels map[string]string)         {}
