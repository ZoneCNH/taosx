# taosx 实现规格

## 范围

本规格描述当前 `pkg/taosx` 的可验证行为。模块提供 TDengine L2 adapter contract，不承诺默认连接真实 TDengine。

## 配置

`Config` 字段：

- `Name`
- `DriverMode`
- `Endpoint`
- `Database`
- `Username`
- `Password`
- `Timeout`
- `MaxRetries`
- `TLS`

`Normalize` 补齐默认名称、驱动模式和超时时间。`Validate` 拒绝缺失 endpoint/database、非法驱动模式、负超时和负重试次数。`Sanitized` 与 `RedactedDSN` 不得暴露密码。

## Client 契约

```go
type Client interface {
	Exec(context.Context, Statement) (ExecResult, error)
	Query(context.Context, Query) (Rows, error)
	WriteBatch(context.Context, Batch) (WriteResult, error)
	SchemalessWrite(context.Context, SchemalessPayload) (WriteResult, error)
	Health(context.Context) HealthStatus
	Close(context.Context) error
}
```

`New(ctx, cfg, opts...)` 必须校验 context、配置和 options。未注入 driver 时构造可以成功，但运行操作返回 unavailable 错误。

## Driver 契约

`Driver` 与 `Client` 方法集一致，表示真实 TDengine I/O 的注入端口。测试 fake、HTTP/WebSocket driver 或其他 SDK wrapper 都应实现此接口。

## 操作行为

- `Exec`：拒绝空 SQL statement，委托 driver。
- `Query`：拒绝空 SQL query，返回 driver 的 `Rows`。
- `WriteBatch`：拒绝空 database、table 或 points；driver 错误时保留 partial result。
- `SchemalessWrite`：拒绝空 lines 或非法协议。
- `Health`：返回 driver 健康状态；默认 unavailable driver 返回 degraded。
- `Close`：幂等；关闭后操作返回 closed 错误。

## 指标

metrics 通过接口注入，默认 no-op。当前核心指标包括：

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

## 非目标

- 真实 TDengine driver。
- 连接池、STMT 和自动重试。
- 自动建表和 schema migration。
- 配置中心读取或观测系统硬依赖。
