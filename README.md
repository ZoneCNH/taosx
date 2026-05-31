# {{MODULE_NAME}}

`{{MODULE_NAME}}` 是一个独立的 Go 基础库模板模块，定位为生产级共享基础库基座。它提供可复用的包结构、稳定错误模型、可观测性 contract、release Evidence 和边界检查，供 `foundationx`、`configx`、`observex`、`postgresx`、`kafkax`、`redisx`、`taosx`、`ossx` 和 `testkitx` 等库生成时复用。

## 目标

本模块的目标不是只生成目录，而是让新基础库从第一天就具备可验证的最小生产语义：

- 显式 `Config`，带稳定 validation error。
- `New`、`Close` 和 `HealthCheck` 具备上下文处理、幂等关闭和生命周期 metrics。
- `ErrorKind`、health status 和 metrics 名称由 `contracts/` 锁定。
- `make ci` 和 `make release-check` 生成可追溯 Evidence。

本仓库默认可编译模板包为 `templatex`。生成具体库时，优先使用 `scripts/render_template.sh` 统一替换 `{{MODULE_NAME}}`、`{{MODULE_PATH}}`、`{{PACKAGE_NAME}}`、Go module path、包目录和 imports。

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
- `contracts/`：JSON schema 和 metrics contract。
- `docs/`：规格、设计、API、配置、测试和发布模板。
- `scripts/`：Harness Gate 脚本。
- `.agent/`：Goal Runtime 工件、Evidence、评审、发布和复盘模板。
- `release/manifest/`：release manifest 模板；`latest.json` 由 release gate 生成并作为 Evidence artifact 保存。

## 文档入口

- [规格](docs/spec.md)：模板能力、验收标准和可追踪性。
- [设计](docs/design.md)：模块边界、公共 API、错误、健康检查和指标设计。
- [API](docs/api.md)：`Config`、`Client`、typed error、health JSON 和 metrics contract。
- [配置](docs/config.md)：显式配置、validation 和脱敏规则。
- [生成](docs/generation.md)：从模板渲染 `foundationx` 等具体基础库。
- [错误模型](docs/errors.md)：`ErrorKind`、`NewError`、`WrapError` 和重试语义。
- [可观测性](docs/observability.md)：指标名、健康状态和 JSON 字段。
- [测试](docs/testing.md)：单元、race、contracts、boundary 和 release 验证要求。
- [发布](docs/release.md)：`release-check`、manifest 字段和 Evidence 规则。

## 命令

```bash
make ci
make release-check
make evidence
```

生成 `foundationx` 示例：

```bash
scripts/render_template.sh \
  --module-name foundationx \
  --module-path github.com/ZoneCNH/foundationx \
  --package-name foundationx \
  --out ../foundationx
```

如果当前目录被父级 `go.work` 包含，建议使用 `GOWORK=off` 验证本模板的独立性：

```bash
GOWORK=off make release-check
```

## Evidence

完成需要 release manifest 和 CI Evidence。`release/manifest/latest.json` 是生成产物，不提交到源码历史；它会记录 module、commit、Go 版本、生成时间、工作区状态和 gate 结果，并由 CI release workflow 上传为 artifact。最终完成声明必须包含 `DONE with evidence:`。
