# taosx 错误 contract

`taosx.Error` 是稳定错误类型，调用方应通过 `errors.As`、`IsKind` 和 `IsRetryable` 做分支判断，不依赖错误字符串。

| `ErrorKind` | 字符串 | 典型场景 | Retryable |
| --- | --- | --- | --- |
| `ErrorKindConfig` | `config` | 配置来源或装配失败。 | 否 |
| `ErrorKindValidation` | `validation` | 字段缺失、identifier 非法、调用参数非法。 | 否 |
| `ErrorKindConnection` | `connection` | TDengine 连接建立失败。 | 通常是 |
| `ErrorKindUnavailable` | `unavailable` | driver 未注入、能力不支持、依赖暂不可用。 | 视场景 |
| `ErrorKindTimeout` | `timeout` | context deadline exceeded 或外部超时。 | 是 |
| `ErrorKindAuth` | `auth` | 认证或授权失败。 | 否 |
| `ErrorKindConflict` | `conflict` | 幂等冲突或资源状态冲突。 | 否 |
| `ErrorKindRateLimit` | `rate_limit` | 限流或配额耗尽。 | 是 |
| `ErrorKindSQL` | `sql` | SQL 构造或执行失败。 | 视场景 |
| `ErrorKindSchema` | `schema` | 表、列、identifier 或 schema helper 校验失败。 | 否 |
| `ErrorKindWrite` | `write` | 批量写入或 schemaless 写入失败。 | 视场景 |
| `ErrorKindInternal` | `internal` | 未分类内部错误。 | 否 |

## 约束

- `NewError` 和 `WrapError` 必须保留 `kind` 与 `retryable` 语义。
- `WrapError` 必须保留 cause，使调用方可以使用 `errors.Is` / `errors.As`。
- context canceled 映射为 `unavailable`，context deadline exceeded 映射为 `timeout`。
- 错误消息、metadata、health message 和 Evidence 不得包含原始密码、token 或生产 DSN。
- 对外日志应使用 `Config.Sanitized()` 或 `Config.RedactedDSN()`。
