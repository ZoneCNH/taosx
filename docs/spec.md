# SPEC-{{MODULE_NAME}}-v1.0

## 需求

- 为可复用基础库提供独立 Go 模块。
- 提供 `Config`、`Validate`、`Sanitize`、`Client`、`New`、`Option`、`HealthCheck`、错误模型、指标钩子和版本元数据。
- 提供 Harness Gate 脚本、CI 工作流、契约、示例、证据、发布和复盘模板。

## 验收标准

- `go test ./...` 和 `go test -race ./...` 通过。
- `make boundary`、`make security`、`make contracts` 和 `make evidence` 通过。
- 模块不得依赖 `x.go`。
- 模块不得隐式读取生产密钥。

## 非目标

- 不包含业务模型、生产连接默认值和隐藏全局客户端。

## 可追踪性

- 目标：`GOAL-20260601-001`
- 模板占位符：`{{MODULE_NAME}}`、`{{MODULE_PATH}}`、`{{PACKAGE_NAME}}`
