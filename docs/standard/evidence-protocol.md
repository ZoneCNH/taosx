# Evidence 协议

Evidence 是完成声明的一部分，不是附加说明。没有 Evidence 不得宣称完成。

## 必需格式

```text
DONE with evidence:
- scope: <task|issue|goal|release>
- gates:
  - <command>: <passed|failed|blocked> <short evidence>
- artifacts:
  - <path>: <purpose>
- known gaps:
  - <none or explicit blocker>
```

## 必需 Artifact

- `release/manifest/latest.json`：由 `make evidence` 生成。`latest.json is a generated Evidence artifact`，`MUST NOT be committed`。
- `release/standard-impact/latest.md`：由 `GOWORK=off make standard-impact-check` 生成，记录标准影响面、`downstream_sync_required`、`downstream_release_decision` 和 `repository_rules_release_decision` 结论。
- release score：来自 `GOWORK=off go run ./cmd/xlibgate score --min 9.8`，并写入 manifest 的 `score` 字段。
- workflow artifact 元数据：manifest 的 `workflow_run_id`、`artifact_name`、`artifact_url` 必须能对齐 CI 上传的 release manifest artifact；本地运行时可记录 `local:*` Evidence URL。
- gate 输出：来自本地命令或 CI job。
- review/retrospective：当变更触达标准、release 或 generator 时必须更新。

## `latest.json` 生命周期

`release/manifest/latest.json` 的唯一职责是记录某次 gate 执行的 Evidence。它必须由 `make evidence`、`make release-check`、`make release-check-extended` 或 `make release-final-check` 在当前工作区生成，并由 `.gitignore` 排除在源码历史之外。

CI 必须把 `latest.json` 和 `latest.json.sha256` 上传为 workflow artifact，并输出该文件的 `sha256`，便于 final Evidence 引用。远端 Evidence 建议记录 `artifact_url`、`workflow_run_id` 和 `sha256`；本地 Evidence 至少记录文件路径和 `sha256`。

Release manifest 测试必须在临时 fixture 仓库中构造所需 `.omc` state，不得读取当前工作区的 Agent 运行态文件。该约束用于证明 Evidence 生成和校验在 clean checkout、CI 和本地开发目录中行为一致。

生命周期链路：

```text
release/manifest/template.json
  -> 提交到源码历史，作为 manifest 字段和结构契约
release/manifest/latest.json
  -> 由 make evidence / release gate 生成
  -> 被 .gitignore 排除，不得提交
  -> 由 CI 上传为 artifact，并在完成声明中记录 artifact_url、sha256 和 workflow_run_id
make release-check
  -> ci + integration + dependency-check + standard-impact-check + docs-check
  -> CHECK_STATUS=passed make evidence
  -> make release-evidence-hash
  -> make release-evidence-check
  -> make release-evidence-checksum-check
make release-final-check
  -> release-check
  -> 要求工作区 clean 后再次校验 release Evidence
```

Context Runtime v4.0 的目标链路必须保持单向：`context-release` 不得依赖 `release-check` 或 `release-final-check`；迁移完成后 `release-final-check` 可以调用 `context-release`。在 wrapper、Makefile target、`cmd/xlibgate` 命令和 registry bridge 物理落地前，完成声明只能把该链路记录为目标态或 known gap，不能写成已执行。release manifest 和 release evidence proof artifacts 必须用实际 gate 输出支撑，不能用目标态文字替代。

本地 release gate 必须运行 `GOWORK=off make dependency-check`、`GOWORK=off make standard-impact-check` 和 `GOWORK=off make docs-check`。

## Manifest 要求

manifest 必须记录：

- `module`：当前 Go module。
- `commit`：执行 Evidence gate 的 HEAD。
- `tree_sha`：当前源码树 SHA。
- `source_digest`：源码摘要。
- `tracked_file_count`：参与摘要的追踪文件数量。
- `go_version`：执行 gate 的 Go 版本。
- `generated_at`：Evidence 生成时间。
- `generated_by`：生成工具或脚本。
- `tree_state`：工作区 clean/dirty 状态。
- `checks`：gate status，必须包含 `dependency_check` 和 `standard_impact`。
- `contracts`：contract digest。
- `dependencies`：dependency list。
- `tools`：tool versions。
- `standard_impact`：标准影响报告摘要。
- `downstream_sync_required`：是否需要同步到 `kernel`、L1/L2 基础库或记录 `x.go` 消费方影响。
- `generator_evidence`：`kernel` 和 `corekit` 的生成验证摘要。
- `workflow`：CI 或本地 Evidence artifact 元数据，至少包含 `workflow_run_id`、`artifact_name`、`artifact_url`。
- `score`：release governance 评分结果、阈值、状态和维度明细。
- `governance_runtime`：Context Runtime v4.0 目标证据字段；落地后必须记录 runtime/schema version、profile 状态、`context-standard`/`context-full`/`context-release` gate 结果、registry source 和 profile wrapper 命令。当前 `internal/tools/releasemanifest/main.go` 尚未发出该字段时，release Evidence 必须把它列入 known gaps，而不是写成 passed。
- `artifacts`：必须包含 `release/manifest/latest.json` 和 `release/manifest/latest.json.sha256`。

外部发布记录或 CI job summary 必须补充 manifest 外部字段：

- `artifact_url`：CI artifact 或 release asset URL；本地运行没有远端 URL 时必须明确写为本地 artifact。
- `sha256`：`release/manifest/latest.json` 的 SHA256。
- `workflow_run_id`：CI workflow run ID；本地运行没有该字段时必须明确说明。

## 完整完成声明字段

Goal 或 Release 级完成声明必须覆盖以下字段，缺失项要写入 `known gaps`：

- `commit`：当前提交或明确说明未提交状态。
- `branch`：当前分支。
- `tag`：发布 tag；非发布目标可写明未创建。
- `release manifest`：`release/manifest/latest.json` 的生成和校验状态。
- `source digest`：manifest 中的源码摘要。
- `contract fingerprint`：manifest 中的 contract 指纹或 digest。
- `dependency list`：manifest 中的依赖清单状态。
- `tool versions`：manifest 中的 Go、工具链和 gate 工具版本状态。
- `release score`：`xlibgate score` 的阈值和 manifest `score` 校验状态。
- `workflow artifact`：`workflow_run_id`、`artifact_name`、`artifact_url` 或明确的本地 artifact 说明。
- `gates`：`fmt`、`vet`、`test`、`race`、`lint`、`security`、`contracts`、`boundary`、`integration`、`dependency-check`、`standard-impact-check`、`evidence`、`release-evidence-check`、`release-final-check`。
- `rendered downstream`：`kernel` 和 `corekit` 的 generator 验证状态；旧 `foundationx` 仅作为迁移扫描项记录。
- `workspace`：clean、dirty 或 blocked，并说明 dirty 原因。

## 禁止声明

- 禁止使用没有命令输出支撑的 “tests pass”。
- 禁止把 skipped required gate 记录为 passed。
- 禁止在 dirty workspace 下宣称 release-final ready。
- 禁止删除失败 Evidence。

## 失败 Evidence

失败 Evidence 仍然有价值。失败时记录：

- 命令。
- 返回码或关键错误。
- 已确认不受影响的范围。
- 下一步修复条件。

## Context Runtime v4 evidence

Release manifest 必须记录 Context Runtime v4.0 的 `governance_runtime` Evidence。必需 Evidence 包括 runtime identifier、profile list、`context-profile-check`、`context-release` 和 legacy alias 保留情况。Standard Impact report 必须把 governance registry、`repository_rules`、`context_runtime` 和 `downstream_context` 变更分类清楚，使下游同步决策显式可审计。
