# taosx 可观测性

`taosx` 只定义可选观测注入 contract，不绑定具体 SDK。生产项目可以在 `Metrics` 中桥接 Prometheus、OpenTelemetry 或内部 metrics 设施。

## Metrics

`WithMetrics(Metrics)` 注入指标记录器。默认 no-op，未注入时不会产生后台 goroutine 或全局状态。

推荐指标名见 [contracts/metrics.md](../contracts/metrics.md)：

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

## 健康检查

`Health(ctx)` 返回 `HealthStatus`：

- `healthy`：driver health 成功。
- `degraded`：未注入真实 driver、能力不可用或依赖暂时降级。
- `unhealthy`：配置无效、client 已关闭、context 失败或 driver 明确失败。

健康结果可以进入 Evidence，但 metadata 只能包含脱敏配置和非敏感运行信息。

## 日志

日志和 tracing attribute 只能记录 `SanitizedConfig`、脱敏 DSN、DriverMode、status、latency 和 retry 分类。不得记录原始密码、token、生产连接串或业务时序数据。
