# taosx 健康检查

`Client.Health(ctx)` 返回 `HealthStatus`，用于本地 smoke、服务启动探针和运行时诊断。健康检查不连接真实 TDengine，除非调用方通过 `WithDriver` 注入真实 driver adapter。

## 状态

- `healthy`：driver 健康检查通过。
- `degraded`：driver 不可用或依赖异常，但客户端仍可返回结构化状态。
- `unhealthy`：context 无效、已取消或 client 已关闭。

默认 unavailable driver 返回 `degraded`，用于证明没有 driver 配置时不会误报 `healthy`。

## 输出边界

`HealthStatus` 允许暴露 `Name`、`DriverMode`、`Database` 和脱敏后的诊断信息。`Message` 必须经过脱敏；不得包含 password、完整 DSN 或 token。

`contracts/health.schema.json` 是 JSON contract；`contracts/contracts_test.go` 会绑定 `healthy`、`degraded` 和 `unhealthy` 三个枚举值。
