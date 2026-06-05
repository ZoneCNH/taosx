# Schemaless failure contract

taosx 的 schemaless 写入必须把 driver 返回的错误归一为 `ErrorKindWrite`，并在部分写入场景保留 `WriteResult`。

## 规则

- driver 返回 `(WriteResult, error)` 时，`Client.SchemalessWrite` 不得丢弃 `WriteResult`。
- error 不得包含明文 password、DSN userinfo password 或 query secret。
- `WriteResult.Partial=true` 表示调用方必须按上层幂等键或业务时间窗决定补偿策略。
- `RowsAttempted` 记录本次 payload line 数；driver 已填充该字段时以 driver 值为准。
- adapter 不自动盲重试 partial write，避免重复写入。
