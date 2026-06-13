# taosx 身份

## 一句话定义

`taosx` 是 TDengine 的 L2 Go 适配器契约模块，为上层系统提供可注入驱动的写入、查询、健康检查和 metrics 边界。

## 是什么

- 一个 Go module：`github.com/ZoneCNH/taosx`。
- 一个公开包：`pkg/taosx`。
- 一组 TDengine adapter contracts：`Client`、`Driver`、`Config`、`Rows`、写入 payload、错误和 metrics。
- 一个默认不可用的 client 实现，用来强制调用方显式注入真实 driver。

## 不是什么

- 不是完整 TDengine 客户端平台。
- 不内置真实 TDengine 网络连接。
- 不提供连接池、STMT、自动重试、自动建表或 schema migration。
- 不直接承载配置中心、观测系统或 resilience 策略。

## 能力边界

| 能力 | 当前状态 |
| --- | --- |
| 配置归一化、校验、脱敏 | 已实现 |
| SQL 执行/查询契约 | 已实现 |
| 批量写入契约 | 已实现 |
| Schemaless 写入契约 | 已实现 |
| 健康检查模型 | 已实现 |
| 可选 metrics 接口 | 已实现 |
| 默认真实 TDengine 连接 | 非目标 |
| 连接池 / STMT / 自动重试 | 非目标 |

## 依赖边界

上游 foundation 矩阵允许 `taosx` 直接依赖 `kernel`。当前核心包应保持最小依赖面；其他横切能力通过调用方注入或外部适配器组合。
