# ADR-20260604-001: L0/L1/L2/L3 分层治理与 L3 私有边界

Status: Accepted
Date: 2026-06-04
Supersedes: 无

## Context

`xlib-standard` 是公开的标准源、模板、generator、Harness 和 Evidence runtime。`kernel`、L1 横切治理库和 L2 基础设施适配库可以公开复用，但 L3 业务系统承载业务模型、业务策略、生产配置注入和部署 wiring，不能开源。

本次标准更新还把 `natsx` 纳入 L2 适配库目标。由于 `.agent/rules/registry.yaml` 是生成结果，分层治理规则需要先进入标准文档、ADR、下游矩阵和 docs-check 锚点，再由既有规则生成与治理流程同步。

## Decision

1. `docs/standard/layer-governance-rules.md` 作为 L0/L1/L2/L3 分层治理的人类可读规则入口。
2. L3 私有业务系统包括 `x.go`、`market-data`、`market-engine`、`macro-data`、`macro-engine`、`regime-engine`，只消费已发布的 L0/L1/L2 基础库版本。
3. 公开基础库不得包含业务模型、业务 topic/subject、业务消息 schema、业务策略、真实连接串、生产密钥或 `/home/k8s/secrets/env/*` 读取逻辑。
4. `natsx` 作为 L2 目标库，必须进入 downstream matrix、downstream adoption status、debt scan 和标准影响检查范围。
5. `docs-check` 负责检查分层规则、ADR、L3 私有边界、`natsx`、`GOPRIVATE`、例外字段和规则补丁记录的文档锚点；更深层的执行约束仍由 `boundary`、`downstream-adoption`、`standard-impact-check`、release Evidence 和私有 CI 承接。

## Consequences

- 新增 L2 适配库时，必须同步更新 `docs/downstream-matrix.md`、`.agent/registries/downstream-adoption-status.yaml`、相关 debt/downstream scan 和 docs-check 锚点。
- L3 验证 Evidence 保留在私有仓库或私有 CI 中；公开 release Evidence 只能记录下游同步结论、blocked 状态、owner 和 known gap，不写入业务细节。
- `.agent/rules/registry.yaml` 继续作为生成产物，不手工编辑；规则变更先改标准文档、Harness、`goalcli` 或生成器输入。

## Rejected

- 拒绝开源 L3 业务系统：会暴露业务模型、策略、生产拓扑和客户数据语义。
- 拒绝把业务 repository、业务 schema 或业务策略下沉到 L2：会破坏基础设施适配库的复用边界。
- 拒绝手工编辑 `.agent/rules/registry.yaml`：会绕过规则生成流程并制造漂移。

## Traceability

- REQ：`REQ-001`、`REQ-009`、`REQ-010`、`REQ-011`
- 规则族：`RULE-DOWNSTREAM-CONTRACT-*`、`RULE-XSTACK-*`
- Gate：`docs-check`、`boundary`、`downstream-adoption`、`standard-impact-check`、release Evidence

## Evidence

本 ADR 的完成证据由本变更的 gate 输出收敛：

- `GOWORK=off make docs-check`
- `GOWORK=off go test ./internal/debtcheck`
- `git diff --check`
