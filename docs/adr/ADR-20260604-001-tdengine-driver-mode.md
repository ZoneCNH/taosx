# ADR-20260604-001: TDengine driver mode

## 状态

Accepted

## 背景

TDengine 可通过 websocket、native driver 和 REST SQL 子集接入。taosx 的 MVA 目标是先固定公共 adapter contract，而不是把标准库绑定到某一个第三方 driver。

## 决策

公共配置使用 `DriverMode`，默认值为 `websocket`。真实 driver 通过 `WithDriver` 注入。

## 后果

- 公共 API 可在没有 TDengine 环境时本地测试。
- health、error、metrics 和 redaction contract 先稳定。
- 真实 driver 支持矩阵必须在后续版本用 integration evidence 证明。
