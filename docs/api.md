# taosx API

公共包路径为 `github.com/ZoneCNH/taosx/pkg/taosx`。API 目标是稳定表达 TDengine adapter contract，而不是绑定某一个第三方 driver。

## 创建客户端

```go
cfg := taosx.Config{
	Name:     "metrics",
	Endpoint: "taos.example.internal:6041",
	Database: "meters",
	Username: "root",
	Password: "secret",
}

client, err := taosx.New(ctx, cfg, taosx.WithDriver(driver), taosx.WithMetrics(metrics))
```

`New` 会检查 `nil`、canceled 和 expired context，并对配置执行 `Validate()`。未注入 `Driver` 时使用 unavailable driver：执行类方法返回 `ErrorKindUnavailable` retryable 错误，`Health` 返回 `degraded`，便于离线测试和文档示例不触达真实基础设施。

## Client contract

`Client` interface 暴露：

- `Exec(ctx, Statement) (ExecResult, error)`
- `Query(ctx, Query) (Rows, error)`
- `WriteBatch(ctx, Batch) (WriteResult, error)`
- `SchemalessWrite(ctx, SchemalessPayload) (WriteResult, error)`
- `Health(ctx) HealthStatus`
- `Close(ctx) error`

所有方法都必须尊重 context。`Close` 必须幂等；第一次成功关闭后，后续 `Close` 返回 nil。

## SQL 与查询

调用方应使用 `Statement`、`Query` 和 schema helper 表达 SQL。禁止把用户输入直接拼接进 SQL 字符串。

```go
sql, err := taosx.RenderCreateStable(taosx.StableSpec{
	Name: "meters",
	Columns: []taosx.ColumnSpec{
		{Name: "ts", Type: "TIMESTAMP"},
		{Name: "current", Type: "DOUBLE"},
	},
	Tags: []taosx.ColumnSpec{
		{Name: "location", Type: "BINARY(32)"},
	},
})
```

schema helper 对库表列名使用 allowlist：首字符为字母或 `_`，后续字符只能是字母、数字或 `_`，最大 64 字符。稳定 SQL helper 的输出顺序由输入 slice 决定，测试使用 exact string 锁定。

## 写入

`Batch` 表达结构化批量写入，`SchemalessPayload` 表达 TDengine schemaless 写入。driver adapter 必须决定当前 DriverMode 是否支持对应能力；不支持时返回 `ErrorKindUnavailable` 或 `ErrorKindWrite`，并保留 retry 语义。

## 可观测性

`WithMetrics(Metrics)` 是可选注入点。本包不依赖 Prometheus、OpenTelemetry 或任何具体可观测性 SDK。
