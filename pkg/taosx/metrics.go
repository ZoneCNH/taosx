package taosx

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
	MetricClientBatchRowsTotal         = "taosx_client_batch_rows_total"
	MetricClientSchemalessLinesTotal   = "taosx_client_schemaless_lines_total"
)

type Metrics interface {
	IncCounter(name string, labels map[string]string)
	ObserveHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
}

type NoopMetrics struct{}

func (NoopMetrics) IncCounter(name string, labels map[string]string) {}

func (NoopMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {}

func (NoopMetrics) SetGauge(name string, value float64, labels map[string]string) {}
