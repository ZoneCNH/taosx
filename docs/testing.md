# 测试模板

## 占位符

- `{{MODULE_NAME}}`
- `{{PACKAGE_NAME}}`

## 必需门禁

- `go test ./...`
- `go test -race ./...`
- `make boundary`
- `make security`
- `make contracts`
- `make evidence`

## 必需覆盖范围

- 配置校验。
- 配置脱敏。
- 客户端创建。
- 幂等关闭。
- 健康与非健康状态检查。

生成的基础库必须保持测试独立于 `x.go`。
