# 发布模板

## 占位符

- `{{MODULE_NAME}}`
- `{{MODULE_PATH}}`
- `{{PACKAGE_NAME}}`

## Release Gate

- `make ci`
- `make integration`
- `make evidence`
- `make release-evidence-check`

推荐入口是：

```bash
GOWORK=off make release-check
```

`GOWORK=off` 用于证明模板不依赖父级 workspace。

发布前的最终入口是：

```bash
GOWORK=off make release-final-check
```

`release-final-check` 会在完整 gate 之后要求 `release/manifest/latest.json` 与当前 HEAD、源码摘要、contract 指纹和依赖清单一致，并要求 git 工作区为 `clean`。它适合在打 tag 或发布前运行；开发中的 `release-check` 允许工作区因为未提交改动显示为 `dirty`，但仍会校验 manifest 与当前内容一致。

## Required Release Check

`make release-check` 是默认发布门禁，必须通过：

```text
ci
integration
evidence
release-evidence-check
```

## Extended Release Check

`make release-check-extended` 是发布前强验证，推荐在重要版本、公共 API 变更、contract 变更、schema 变更、metrics 变更时执行：

```text
ci-extended
integration
evidence
release-evidence-check
```

`make ci-extended` 会在默认 `ci` 外追加：

```text
property
golden
fuzz-smoke
```

## Gate 工具契约

`make ci` 中的 `make lint` 和 `make security` 是强制 gate。运行前必须可用：

- `golangci-lint`
- `govulncheck`

缺少任一工具时，本地 Makefile 必须硬失败。GitHub Actions CI 会在运行 `make ci` 前安装 `golangci-lint` 和 `govulncheck`，以保证本地与 CI 对同一组强制 gate 负责。

`make security` 必须同时运行 `govulncheck ./...` 和 `scripts/check_secrets.sh`；不得把漏洞扫描降级为可选检查。

## Evidence

发布 Evidence 生成到 `release/manifest/latest.json`，该文件是生成产物，不提交到源码历史。提交到仓库的是 `release/manifest/template.json`；CI release workflow 会上传 `latest.json` 作为 artifact。

`latest.json` 至少包含：

- `module`
- `version`
- `commit`
- `tree_sha`
- `source_digest`
- `tracked_file_count`
- `go_version`
- `generated_at`
- `generated_by`
- `tree_state`
- `checks`
- `contracts`
- `dependencies`
- `tools`
- `artifacts`
- `notes`

`make release-check` 成功后会以 `CHECK_STATUS=passed` 生成 manifest，并立即运行 `make release-evidence-check`。若单独运行 `make evidence`，未显式传入的检查状态默认为 `unknown`，后续校验会拒绝把这些状态当作已通过的 release gate。因为 `latest.json` 不再提交，manifest 中的 `commit` 可以指向实际执行 release gate 的 HEAD，避免自引用提交哈希导致的永久漂移。

Extended Evidence 推荐额外记录：

- `make ci-extended` 结果。
- `make property` 结果。
- `make fuzz-smoke` 结果。
- `make golden` 结果。
- compatibility 和 observability contract 结果。

`source_digest` 基于 `git ls-files` 中的受跟踪文件内容计算；`contracts` 固定记录核心 contract 文件的 SHA256；`dependencies` 来自 `go list -m -json all`；`tools` 记录 Go、`golangci-lint` 和 `govulncheck` 的版本或可用状态。这些字段由 `internal/tools/releasemanifest` 生成并校验，不再由 shell 拼接 JSON。

`make integration` 会调用 `scripts/render_template.sh` 生成临时 `foundationx` 和 `corekit` 两个下游库，并对每个生成目录执行：

- 模块路径、包目录和旧模板标识扫描。
- `GOWORK=off go test ./...`
- `GOWORK=off make contracts`
- `GOWORK=off make boundary`
- `CHECK_STATUS=passed GOWORK=off make evidence`
- `RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check`

这一步用于证明模板替换、包目录迁移、imports、contracts、边界检查和生成后 release Evidence 都能在下游库中独立工作。

## 规则

- 没有 Evidence artifact 不得发布。
- `tree_state` 为 `dirty` 时可以在开发中生成 Evidence，但正式发布前必须通过 `make release-final-check`。
- 不得在 release manifest、PR、Issue 或变更日志条目中包含原始凭据。
- 不得依赖 `github.com/bytechainx/x.go` 或 `github.com/ZoneCNH/x.go`。
- public API、config schema、error kind、health JSON、metrics name 或 release manifest schema 变更必须在 release notes 或 release manifest 中显式标记 breaking change。
