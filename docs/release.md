# 发布模板

## 占位符

- `{{MODULE_NAME}}`
- `{{MODULE_PATH}}`
- `{{PACKAGE_NAME}}`

## 发布门禁

- `go test ./...`
- `go test -race ./...`
- `make boundary`
- `make security`
- `make contracts`
- `make evidence`

## 证据

发布证据生成到 `release/manifest/latest.json`。

## 规则

- 没有证据不得发布。
- 不得在清单、PR、Issue 或变更日志条目中包含原始凭据。
- 不得依赖 `x.go`。
