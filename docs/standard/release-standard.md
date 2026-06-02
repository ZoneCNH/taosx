# 发布标准

发布流程必须证明源码、contracts、依赖和 gate 状态一致。`xlib-standard` 的 release 标准同时约束生成基础库；旧 `baselib-template` 仅作为迁移兼容名记录。

## 发布路径

1. 运行 required gate。
2. 运行 integration 和 generator smoke。
3. 生成 Evidence manifest。
4. 校验 Evidence manifest。
5. 校验 release score 和 workflow artifact Evidence。
6. 在 clean workspace 运行 final check。
7. 使用明确版本运行 preflight。
8. 在 PR 或 release notes 中附上 Evidence 摘要。

## 命令

```bash
GOWORK=off make release-check
GOWORK=off go run ./cmd/xlibgate score --min 9.8
GOWORK=off make release-check-extended
GOWORK=off make release-final-check
GOWORK=off make release-preflight VERSION=v1.0.0
```

`release-check` 内置 `score-check`，默认要求 `score >= 9.8`。`release-final-check` 会在 clean workspace 约束之外再次校验 manifest 内的 score threshold；release score 只能作为发布治理完整性信号，不能替代 `make ci`、`make security`、integration、race 或人工语义审查。

## Manifest

`release/manifest/latest.json` 是生成产物：

- 可以作为 CI artifact 上传。
- 可以作为本地 Evidence 检查输入。
- 不提交到源码历史。
- `release/manifest/latest.json.sha256` 是对应 checksum 产物，随 CI artifact 上传，并保持在 `.gitignore` 中。
- manifest 必须记录 `score` 和 `workflow`；`workflow_run_id`、`artifact_name`、`artifact_url` 用于对齐 CI 上传的 release manifest artifact，本地运行时可使用 `local:*` Evidence URL。

Release manifest 相关测试必须在临时 fixture 仓库构造所需 `.omc` state，不得依赖当前工作区的 Agent 运行态文件。

## 供应链约束

- GitHub Actions workflow 引用的第三方 Action 必须固定为 40 位 commit SHA，并在同一行保留来源 tag 注释。
- CI、Release Check 和 Security workflow 安装 `govulncheck` 时必须使用固定版本；当前基线是 `golang.org/x/vuln/cmd/govulncheck@v1.3.0`。
- 本地缺少 `golangci-lint` 或 `govulncheck` 时，`make lint` / `make security` 必须失败，不得把必需 gate 记录为跳过。

## 版本

- `VERSION` 必须显式传入 release-preflight。
- 版本应与 release notes、tag 和 manifest 一致。
- 未创建 tag 或工作区 dirty 时，不得宣称最终发布完成。

## 变更说明

PR 或 release notes 必须说明：

- 对模板行为的影响。
- 对生成库的影响。
- 已运行命令。
- Evidence artifact。
- known gaps 或 blocked gate。

## Context Runtime v4 release profile

Context Runtime v4.0 将 `context-release` 定义为 release profile baseline。`release-final-check` 委托给 `context-release`；`context-release` 不得调用 `release-check` 或 `release-final-check`，以避免递归 release governance。release manifest 必须包含 `governance_runtime` evidence，记录启用的 profile 集合、`context-profile-check`、`context-release` 和 legacy profile aliases。
