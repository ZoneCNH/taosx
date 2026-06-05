# taosx 配置

`taosx.Config` 由调用方显式构造或由上层配置库装配后传入。本包不读取环境变量、配置文件或 `/home/k8s/secrets/env/*`。

## 字段

| 字段 | 说明 | 默认值 | 校验 |
| --- | --- | --- | --- |
| `Name` | 客户端名称，用于 health 和 metrics label。 | `taosx` | 非空 |
| `DriverMode` | TDengine driver 模式。 | `websocket` | 必须是允许值 |
| `Endpoint` | TDengine endpoint，允许 `host:port` 或带 scheme 的地址。 | 无 | 非空 |
| `Database` | 默认数据库。 | 无 | 非空 |
| `Username` | 连接用户名。 | 空 | 可空 |
| `Password` | 连接密码。 | 空 | 可空，输出必须脱敏 |
| `Timeout` | 单次操作默认超时。 | `5s` | 不得为负 |
| `MaxRetries` | driver 可使用的最大重试次数。 | `0` | 不得为负 |
| `TLS` | 是否使用 TLS scheme。 | `false` | 布尔值 |

## DriverMode

- `websocket`：默认模式，生成 `taosws` / `taoswss` DSN。
- `native_legacy`：保留给旧 native driver adapter，DSN 仍按 TDengine WebSocket 兼容格式脱敏表达。
- `rest_sql_only`：REST SQL fallback，生成 `http` / `https` DSN；只承诺 SQL 执行语义，不承诺批量写入和 schemaless 能力等价。

## 校验与脱敏

`Validate()` 会在归一化默认值后检查必填字段和数值边界。配置错误返回 `ErrorKindValidation`，错误消息不得包含原始密码。

`Sanitized()` / `Sanitize()` 返回 `SanitizedConfig`：

- `Password` 固定为 `***`，空密码保持空。
- `RedactedDSN` 中 userinfo 密码固定为 `***`。
- 可安全写入日志、Evidence、health metadata 和发布说明。

`contracts/config.schema.json` 与 `Config` 字段保持同步，schema required 字段为 `endpoint` 和 `database`。
