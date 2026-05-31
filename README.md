# {{MODULE_NAME}}

`{{MODULE_NAME}}` 是一个独立的 Go 基础库模板模块。

## 目标

本模块为 `foundationx`、`configx`、`observex`、`postgresx`、`kafkax`、`redisx`、`taosx`、`ossx` 和 `testkitx` 等独立基础库提供标准脚手架。

本仓库默认可编译模板包为 `templatex`。生成具体库时，应将 `{{MODULE_NAME}}`、`{{MODULE_PATH}}` 和 `{{PACKAGE_NAME}}` 替换为实际名称。

## 非目标

- 不依赖 `x.go`。
- 不包含 `x.go` 业务模型。
- 不隐式读取生产密钥。
- 不创建隐藏全局客户端。

## 标准结构

- `pkg/{{PACKAGE_NAME}}`：公共包 API。
- `internal/`：脱敏、校验和运行时说明等内部辅助代码。
- `testkit/`：可复用测试夹具和断言。
- `examples/`：最小使用示例。
- `contracts/`：JSON schema 和指标契约。
- `docs/`：规格、设计、API、配置、测试和发布模板。
- `scripts/`：Harness Gate 脚本。
- `.agent/`：Goal Runtime 工件、证据、评审、发布和复盘模板。
- `release/manifest/`：发布清单模板和生成的证据。

## 命令

```bash
make ci
make release-check
make evidence
```

## 证据

完成需要发布清单和 CI 证据。最终完成声明必须包含 `DONE with evidence:`。
