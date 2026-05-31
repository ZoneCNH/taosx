# API 模板

## 占位符

- `{{MODULE_NAME}}`：生成的仓库名称。
- `{{MODULE_PATH}}`：生成的 Go 模块路径。
- `{{PACKAGE_NAME}}`：生成的包名。

## 公共 API

- `Config`：由用户显式提供的配置。
- `Validate`：拒绝无效配置。
- `Sanitize`：在日志或证据采集前屏蔽敏感值。
- `New`：基于显式配置创建客户端。
- `Close`：释放资源，并且必须幂等。
- `HealthCheck`：报告客户端健康状态。
- `Error`：稳定错误契约。
- `Metrics`：注入式指标钩子。
- `Version`：发布版本。

生成的基础库不得依赖 `x.go`。
