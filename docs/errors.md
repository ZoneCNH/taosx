# 错误模板

## 占位符

- `{{MODULE_NAME}}`
- `{{PACKAGE_NAME}}`

## 错误类型

- `config`
- `validation`
- `connection`
- `unavailable`
- `timeout`
- `auth`
- `conflict`
- `rate_limit`
- `internal`

错误应保持稳定，可在有价值时包装，并且可以安全纳入证据。生成的库不得使用 `x.go` 业务模型。
