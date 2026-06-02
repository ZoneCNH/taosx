# 变更日志

## 未发布

### 修复

- Release manifest 测试改用临时 git fixture 构造 `.omc/state/agent-replay-fixture.jsonl`，避免依赖本地 Agent 运行态文件。
- GitHub Actions workflow 固定 `checkout`、`setup-go`、`cache` 和 `upload-artifact` 的 40 位 commit SHA，并将 `govulncheck` 固定为 `v1.3.0`。
- Secret Gate 同时排除 `.omc` 和 `.omx` 本地运行态目录，避免扫描 Agent 状态文件时产生误报。

### 测试

- 补齐 `internal/releasequality` 对 `Compute`、`Verify` 和 `Marshal` 的单元测试。

### 文档

- 对齐 `docs/independent-audit-20260602.md` 的修复状态，并补充 score 语义边界、workflow pinning 和固定工具版本要求。

## v0.3.5 - 2026-06-02

### 新增

- 新增 `cmd/xlibgate` gate 路由入口，统一封装 `release-final-check`、`release-evidence-check` 和 `score` 等发布前检查。
- 新增 release scorecard 文档与 `internal/releasequality` 评分实现，将 manifest、workflow artifact、安全门禁、复盘补丁和文档约束汇总为可执行分数。
- 新增 downstream compatibility matrix，记录 `kernel`、`foundationx` 和 `corekit` 渲染后的测试、contracts、boundary 与 Evidence 验证结果。

### 治理

- `release-final-check` 强制校验 scored release Evidence，避免仅凭局部 gate 结果推进发布。
- 发布文档、Agent 运行时文档和标准文档统一到 `xlib-standard` 命名与 release gate 口径。
- GitHub CI 与 release workflow 对齐 `GOWORK=off`、docs-check、security、contracts、boundary 和 release manifest 校验。

### 验证

- 发布前已运行 `GOWORK=off make release-final-check`。
- 发布前已运行 `go run ./cmd/xlibgate score --min 9.8`，当前质量分为 10。
- 发布前已运行 `go run ./cmd/xlibgate release-evidence-check`，确认 `release/manifest/latest.json` 通过校验。

## v0.3.0 - 2026-06-01

### 新增

- Release Evidence 现在同时生成并校验 `release/manifest/latest.json.sha256`，确保发布 manifest 和 checksum artifact 成对存在。
- `release/manifest/template.json` 将 checksum 纳入必需 artifacts，发布产物清单能完整描述可验证 Evidence。

### 治理

- `make release-check`、`make release-check-extended` 和 `make release-final-check` 强制要求 `GOWORK=off`，避免发布门禁受外部 workspace 污染。
- `docs-check` 新增标准源、checksum artifact、`GOWORK=off` 发布命令和 fuzz smoke 文档约束，防止文档与发布 Harness 漂移。
- 扩展 `golangci-lint` 规则集，提高模板基础库的静态质量门槛。

## v0.2.0 - 2026-06-01

### 新增

- 新增 `make release-preflight VERSION=vX.Y.Z`，在打 tag 前检查版本、`main` 同步状态、目标 tag、`CHANGELOG.md`、必需工具和最终 release gate。

### 修复

- Release Check workflow 在运行 `make release-check` 前安装 `golangci-lint` 和 `govulncheck`，并使用 `GOWORK=off`，与 CI 的强制 gate 环境保持一致。
- Release Evidence 校验新增目标版本比对，避免目标 tag 与 `manifest.version` 不一致。

## v0.1.0 - 2026-06-01

### 新增

- 初始化 `baselib-template` 结构。
- 添加标准 Go 基础库包骨架。
- 添加 Makefile 命令。
- 添加 Harness Gate 脚本。
- 添加 GitHub Actions 工作流。
- 添加 contracts 文件。
- 添加 Agent 运行时模板。
- 添加 release manifest 模板。
- 添加 typed error、错误包装和 `ErrorKind` contract。
- 添加 client 生命周期、健康检查和请求扩展 metrics contract。
- 添加 health JSON contract 与 contracts 回归测试。
- 添加 config schema 到 `Config` 字段映射的 contract 回归测试。
- 添加 `scripts/render_template.sh`，支持生成 `foundationx` 等具体基础库。
- 添加 `examples/basic`、`examples/config` 和 `examples/health` smoke 测试，锁定文档示例输出。
- 添加 `testkit` 夹具和断言回归测试。
- 添加配置属性测试、配置 fuzz smoke 测试、健康状态 golden 测试和 `testkit` golden 文件工具。

### 安全

- 添加 Secret Gate。
- `make security` 强制运行 `govulncheck ./...` 和密钥扫描；缺少 `govulncheck` 时必须失败。
- 配置脱敏规则覆盖 release Evidence 和日志可见内容。
- Boundary Gate 同时拦截 `github.com/bytechainx/x.go` 和 `github.com/ZoneCNH/x.go`。

### 治理

- 添加 Evidence 和复盘模板。
- CI 在 `make ci` 前安装 `golangci-lint` 和 `govulncheck`，与 Makefile 强制 gate 对齐。
- `make release-check` 统一执行 CI、integration 和 manifest 生成。
- `make release-final-check` 在发布前串联 `release-check`、release Evidence 校验和工作区洁净校验。
- `make integration` 通过临时 `foundationx` 和 `corekit` 渲染、测试、contracts、boundary 与 Evidence 生成验证模板链路。
- `release/manifest/latest.json` 作为生成产物保留在源码历史之外，避免 release Evidence 与源码提交互相污染。

### 验证

- 发布前已运行 `GOWORK=off make release-final-check`。
- `go fmt ./...`、`go vet ./...`、`golangci-lint run ./...`、`go test ./...`、`go test -race ./...`、Boundary、Security、contracts、integration 和 release Evidence 校验均通过。
- `v0.1.0` 为 annotated tag，指向提交 `b6dfe9b93e4417a3b7e077cec1b4c0fffdc37240`。
