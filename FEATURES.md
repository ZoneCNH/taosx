# 功能总览

`taosx` 是 `github.com/ZoneCNH/taosx` 的 L2 TDengine adapter 基础库，当前发布版本为 `v1.0.5`。它负责把公共配置、driver 注入、错误分类、SQL/写入构造、健康检查和可观测性封装成稳定契约，不承担业务时序模型或应用编排。

## 核心能力

- `Config`：默认值补齐、字段校验、脱敏副本和 redacted DSN。
- `Client`：`Exec`、`Query`、`WriteBatch`、`SchemalessWrite`、`Health` 和 `Close`。
- `Statement`、`Query`、`Batch`、`SchemalessPayload`：覆盖 SQL、查询、批量写入和 schemaless 写入的输入结构。
- `ColumnSpec`、`StableSpec`、`RenderCreateStable`：提供 identifier allowlist 和稳定 STABLE SQL 生成。
- `Error`、`ErrorKind`、`IsKind`、`IsRetryable`：提供稳定错误 contract。

## 边界

- 默认 DriverMode 为 `websocket`，面向 TDengine WebSocket 连接。
- `native_legacy` 仅用于兼容旧运行时。
- `rest_sql_only` 仅用于 REST SQL fallback，不承诺批量或 schemaless 等价。
- 不隐式读取 `/home/k8s/secrets/env/*`，不记录原始密码、token 或完整生产 DSN。

## 权威入口

- [README](README.md)：项目简介、公共 API 和本地验证入口。
- [配置](docs/config.md)：DriverMode、字段、校验和脱敏规则。
- [API](docs/api.md)：Client、driver 注入和写入/查询 contract。
- [错误](docs/errors.md)：错误分类、retry 语义和 secret redaction。
- [可观测性](docs/observability.md)：metrics 注入与健康状态。
- [测试策略](docs/testing.md)：单测、集成测试和 gate 覆盖。
- [发布](docs/release.md)：release gate、preflight 和 auto patch 语义。
