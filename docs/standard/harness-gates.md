# Harness Gates

Harness Gate 把 `xlib-standard` 的标准、模板、generator、Evidence 和 release 要求变成可执行检查。

Full Goal Runtime v3.1 以 `cmd/xlibgate` 作为 Go gate runtime。Makefile target 是推荐的人机入口，内部必须委托到 `GOWORK=off go run ./cmd/xlibgate ...`；`scripts/*.sh` 是兼容实现层，不再作为 CI/发布文档中的唯一权威入口。

## Required Gates

| Gate | 命令 | 目的 |
| --- | --- | --- |
| Format | `GOWORK=off make fmt` | 保持 Go 格式稳定 |
| Vet | `GOWORK=off make vet` | 基础静态检查 |
| Lint | `GOWORK=off make lint` | `golangci-lint` 强制检查，缺失时失败 |
| Unit | `GOWORK=off make test` | 单元和示例 smoke |
| Race | `GOWORK=off make race` | 并发安全基线 |
| Boundary | `GOWORK=off make boundary` | 模块边界和模板禁止项 |
| Security | `GOWORK=off make security` | `govulncheck` 和 secret scan |
| Contracts | `GOWORK=off make contracts` | schema、metrics 和 manifest contract |
| Docs Check | `GOWORK=off make docs-check` | 文档、链接、当前命名、下游同步策略、v3.1 runtime 和 release protocol |
| Integration | `GOWORK=off make integration` | generator 和 downstream smoke |
| Dependency Check | `GOWORK=off make dependency-check` | 校验 `renovate.json`、`.github/dependabot.yml` 和 Go dependency inventory |
| Standard Impact Check | `GOWORK=off make standard-impact-check` | 生成 `release/standard-impact/latest.md` 并判定 `downstream_sync_required` |
| Score | `GOWORK=off make score` / `GOWORK=off go run ./cmd/xlibgate score --min 9.8` | 校验 v3.1 gate runtime、CI 和文档契约一致性 |
| Evidence | `CHECK_STATUS=passed GOWORK=off make evidence` | 生成 release manifest |
| Release Evidence | `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check` | 校验 manifest 与仓库事实 |

## Extended Gate

- `GOWORK=off make property`
- `GOWORK=off make golden`
- `GOWORK=off make fuzz-smoke`
- `GOWORK=off make ci-extended`
- `GOWORK=off make release-check-extended`

## Generator Gate

Generator gate 必须证明模板能生成代表性下游，而不是只证明 `xlib-standard` 自身可用。

代表下游：

- `kernel`
- `corekit`

旧 `foundationx` 只作为迁移兼容扫描项，不再作为默认下游。

## Final Gates

- `GOWORK=off make release-final-check`
- `GOWORK=off make release-preflight VERSION=<version>`
- `GOWORK=off go run ./cmd/xlibgate score --min 9.8`
- `GOWORK=off make integration DOWNSTREAM=kernel`

## Secret Gate

Secret Gate 必须确认源码、README、测试日志、release manifest、PR 描述和 Evidence 不包含 `/home/k8s/secrets/env/*` 的真实内容。该路径只能在文档中作为调用方部署路径名出现。

Secret scan 会排除 `.git`、`.omc`、`.omx` 和 `vendor` 等本地或第三方目录，避免把 Agent runtime 或 vendored 依赖误判为源码凭据；这些目录一旦内容进入 git 历史、manifest、PR、Issue 或日志，仍按 secret leak 处理。

## Workflow Supply Chain Gate

CI、Release Check、Integration 和 Security workflow 引用的第三方 Action 必须固定为 40 位 commit SHA，并保留来源 tag 注释。`govulncheck` 安装必须使用固定版本；当前发布门禁基线是 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`。
