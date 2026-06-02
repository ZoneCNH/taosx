# xlibgate CLI Contract

`xlibgate` 是 xlib-standard v2.9.3 的 machine-verifiable gate surface。除明确委托给既有脚本的命令外，所有命令必须输出包含 `command`、`status`、可选 `details` 和可选 `gaps` 的 JSON report，并符合 `contracts/xlibgate-report.schema.json`。

该 CLI 只执行本地非破坏性 contract checks：不读取真实 secrets，不连接外部生产系统，也不修改 downstream 仓库。涉及 GitHub、release、runtime 或 downstream 状态的命令，在当前版本中以本地 contract、manifest 和 dry-run 证据表达。

## 退出码语义

- `status=passed` 必须返回 `0`。
- `status=failed`、`status=planned` 和 `status=gap` 必须返回 `1`。
- 未知命令、非法参数或非法 context 必须返回 `2`。
- downstream 仓库不存在时必须报告 `status=gap` 且返回非 0；该结果不能作为 release gate 成功证据。
- 带 `--verify` 或 `--strict` 的命令遇到 `planned`/`gap` 时必须阻断 gate。

## P0 commands

- `version`
- `doctor`
- `minimal-kernel`
- `main-guard --context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify`
- `worktree-guard --context local_write|local_readonly|ci_pull_request|ci_main_verify|release_verify`
- `evidence-check`
- `done-assertion`
- `cli-contract`
- `issue-registry`
- `command-registry`
- `makefile-baseline`
- `context-profile`
- `context-profile-check`
- `context-schema-check`
- `context-fast-check`
- `context-standard-check`
- `context-full-check`
- `boundary`
- `contracts`
- `dependency-check`
- `docs-check`
- `evidence`
- `manifest`
- `integration`
- `release-evidence-check`
- `release-evidence-checksum-check`
- `release-evidence-hash`
- `release-final-check`
- `render-check`
- `score`
- `secrets`
- `security`
- `standard-impact-check`

## P1 commands

- `agent-team-contract`
- `scope-lock`
- `pr-template`
- `acceptance-matrix`
- `runtime-health`
- `goal-runtime`
- `naming`
- `upgrade-standard --dry-run --repo <path>`
- `conformance-profile`
- `downstream-registry`
- `self-healing-skeleton`
- `policy-schema`
- `github-settings --verify`
- `github-governance`
- `governance-fixture-test`
- `toolchain`
- `evidence-artifacts`
- `supply-chain`
- `autoresearch`
- `changelog`

## P2 commands

- `install-runtime --dry-run`
- `upgrade-runtime --dry-run`
- `release-ready`
- `evidence-replay`
- `attest-conformance --profile standard-source|l0-kernel`
- `pack-standard`
- `pack-gate`
- `pack-evidence`
- `runtime-file-ownership`
- `downstream-baseline --repo kernel/configx --mode patch-only`
- `downstream-adoption --repo kernel/configx --mode patch-only`
- `execution-context`

## 执行约束

- `--dry-run` 只能执行本地 contract、manifest 或 patch planning 检查。
- `--verify` 或 `--strict` 不能把 `planned`/`gap` 当作成功证据。
- `--context` 仅允许 `local_write`、`local_readonly`、`ci_pull_request`、`ci_main_verify` 和 `release_verify`。
- `--repo` 指向的 downstream 仓库不存在时，命令必须返回 `gap`，且不得自动创建目录。
- 新增命令时必须同步 `run` dispatch、`plannedCommandFiles`、Makefile gate、CLI contract 和测试表。

## Context Runtime v4 commands

Context Runtime v4.0 增加增量 profile baseline，不替换现有 P0/P1/P2 command registry。以下命令受 registry governance 约束，必须同时保留在 `run` dispatch、`.agent/command-registry.yaml`、`.agent/issue-registry.yaml`、Makefile targets 和本 contract 中：

- `context-profile --profile lite|standard|full|release`
- `context-profile-check`
- `context-schema-check`
- `context-lite`
- `context-standard`
- `context-full`
- `context-release`
- `context-fast-check`
- `context-standard-check`
- `context-full-check`

`context-release` 不得调用 `release-check` 或 `release-final-check`；允许的方向是 `release-final-check` 在 strict release evidence checks 前委托给 `context-release`。Legacy aliases（`context-fast-check`、`context-standard-check`、`context-full-check`）必须保持可用。除非 `.agent/context` 文件实际存在于仓库中，否则该 baseline 不得宣称这些文件已落地。
