# DESIGN-{{MODULE_NAME}}-v1.0

## 架构

生成的库是独立 Go 模块。公共 API 位于 `pkg/{{PACKAGE_NAME}}`，内部辅助代码位于 `internal/`，契约位于 `contracts/`，运行证据位于 `release/manifest/`。

## 公共 API

模板暴露 `Config`、`SanitizedConfig`、`Client`、`New`、`Close`、`Option`、`HealthCheck`、`Error`、`Metrics`、`NoopMetrics`、`ModuleName` 和 `Version`。

## 配置

调用方必须显式传入配置。生成的库不得隐式读取 `x.go` 生产密钥路径。

## 错误模型

错误使用稳定的 `ErrorKind` 枚举，并通过 `Unwrap` 支持错误包装。

## 健康检查

持有资源的客户端暴露 `HealthCheck(context.Context)`，并返回 `healthy`、`degraded` 或 `unhealthy`。

## 指标

指标通过钩子注入，默认使用无操作实现。

## 测试

模板要求为配置校验、脱敏、客户端生命周期、健康检查和内部辅助代码提供单元测试与竞态测试。

## 发布

发布前必须通过 Harness Gates，并生成 `release/manifest/latest.json`。
