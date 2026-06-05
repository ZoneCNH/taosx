# taosx Metrics

taosx 只定义 metrics contract，不绑定 Prometheus、OpenTelemetry 或其它 SDK。调用方通过 `WithMetrics` 注入实现。

## 指标

- `taosx_client_created_total`
- `taosx_client_closed_total`
- `taosx_client_errors_total`
- `taosx_client_health_status`
- `taosx_client_health_latency_ms`
- `taosx_client_requests_total`
- `taosx_client_request_duration_seconds`
- `taosx_client_retries_total`
- `taosx_client_inflight`
- `taosx_client_batch_rows_total`
- `taosx_client_schemaless_lines_total`

结构化 contract 位于 `contracts/metrics.contract.yaml`。禁止把 `endpoint`、`username`、`password` 或完整 DSN 作为 label。
