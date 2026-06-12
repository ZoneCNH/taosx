# taosx 设计

## 设计目标

`taosx` 把 TDengine 访问收敛为一个小而稳定的 Go 契约。核心包只负责边界、类型和委托，不负责选择真实 TDengine driver、连接池实现或重试策略。

目标是让上层模块可以用同一组 `Client` 方法表达执行 SQL、查询、批量写入、schemaless 写入、健康检查和关闭，同时把真实 I/O 放在可替换的 `Driver` 适配器后面。

## 架构边界

```text
caller
  |
  | taosx.Config / Statement / Query / Batch / SchemalessPayload
  v
pkg/taosx.Client
  |
  | validation, redaction, metrics, error classification
  v
taosx.Driver
  |
  | implemented outside core package
  v
TDengine transport / SDK / test fake
```

核心包可以独立测试。真实 TDengine 连接由调用方或后续适配器实现，并通过 `WithDriver` 注入。

## 主要组件

| 组件 | 职责 |
| --- | --- |
| `Config` | 表示连接配置，提供默认值、校验、脱敏快照和 redacted DSN。 |
| `Client` | 对外稳定端口，处理校验、状态、metrics 和驱动委托。 |
| `Driver` | 真实 I/O 的适配器端口。 |
| `Statement` / `Query` | 原始 SQL 边界类型，只拒绝空 SQL。 |
| `Batch` | 有 database、table、points 的批量写入契约。 |
| `SchemalessPayload` | TDengine schemaless 行协议契约。 |
| `Rows` | 查询结果抽象。 |
| `Metrics` | 可选指标接口，默认 no-op。 |

## 错误与脱敏

错误使用 `taosx.<Operation>` 操作名，保留 validation、unavailable、closed 等分类。配置和 DSN 展示必须 redacted，不能泄漏密码。

## 指标

核心包只调用 `Metrics` 接口。默认实现是 no-op；调用方可以注入 Prometheus、OpenTelemetry 或其他适配器。当前指标名前缀为 `taosx_client_*`。

## 非目标

- 内置真实 TDengine driver。
- 连接池、STMT 写入、自动重试、退避或熔断。
- 自动建表、schema migration、订阅或流式处理。
- 配置中心、环境变量读取或密钥管理。
