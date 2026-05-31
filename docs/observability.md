# 可观测性模板

## 占位符

- `{{MODULE_NAME}}`
- `{{PACKAGE_NAME}}`

## 指标

使用 `contracts/metrics.md` 中的指标契约。

## 健康检查

持有资源的客户端必须暴露 `HealthCheck(context.Context)`。

## 日志

只能记录脱敏配置。不得记录原始凭据或生产连接材料。

本模板不得依赖 `x.go`。
