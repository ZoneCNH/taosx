# Metrics Contract

标准指标用于描述 `pkg/taosx` 暴露给调用方的最小可观测面。实现可以接入任意 metrics 后端，但指标名、类型和标签语义必须保持兼容。

| 指标 | 类型 | 标签 | 说明 |
| --- | --- | --- | --- |
| `taosx_client_created_total` | counter | `name`, `driver_mode` | 成功创建 client 的次数。 |
| `taosx_client_closed_total` | counter | `name` | 成功关闭 client 的次数；重复关闭不重复计数。 |
| `taosx_client_errors_total` | counter | `op`, `kind` | client 生命周期错误次数，`kind` 必须来自 error contract。 |
| `taosx_client_health_status` | gauge | `name`, `status`, `driver_mode` | 健康状态数值，healthy 为 `1`，其他状态为 `0`。 |
| `taosx_client_health_latency_ms` | histogram | `name`, `status` | 单次健康检查耗时，单位为毫秒。 |
| `taosx_client_requests_total` | counter | `op`, `driver_mode` | TDengine 操作请求次数。 |
| `taosx_client_request_duration_seconds` | histogram | `op` | TDengine 操作耗时，单位为秒。 |
| `taosx_client_retries_total` | counter | `op` | retryable 操作重试计数。 |
| `taosx_client_inflight` | gauge | `op` | 当前进行中的 TDengine 操作数量。 |
| `taosx_client_batch_rows_total` | counter | `database` | 批量写入请求记录的行数事件。 |
| `taosx_client_schemaless_lines_total` | counter | `protocol` | schemaless 写入请求记录的行数事件。 |
