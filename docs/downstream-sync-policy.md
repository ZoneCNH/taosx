# 下游同步策略

本文定义 `xlib-standard` 变更后如何同步到下游基础库。`xlib-standard` 是 Standard Source、Go Reference Template、Generator、Harness 和 Evidence Runtime 的唯一标准源；下游只消费这些标准和模板结果，不反向决定标准内容。

旧 `baselib-template` 和 `foundationx` 名称只保留在迁移 ADR、迁移指南、历史记录和兼容性说明中。当前同步目标使用 `kernel`、`corekit` 和 `xlib-standard` 命名。

## 角色

| 角色 | 当前名称 | 同步关系 |
| --- | --- | --- |
| Standard Source | `xlib-standard` | 定义标准、边界、gate、Evidence 和下游同步规则 |
| Go Reference Template | `xlib-standard` | 提供可渲染模板和参考实现 |
| Generator | `scripts/render_template.sh` / `cmd/xlibgate integration` | 把模板渲染为具体基础库 |
| L0 代表下游 | `kernel` | 第一优先级同步目标，验证最小基础库形态 |
| L1 基础库 | `configx`、`observex`、`testkitx` | 继承 L0 标准并提供基础能力 |
| L2 基础库 | `postgresx`、`redisx`、`kafkax`、`taosx`、`ossx`、`clickhousex` | 在 L1 能力上提供具体基础设施适配 |
| 组合层 | `x.go` | 仅作为消费方组合基础库，不反向影响 `xlib-standard` |

## 变更到同步动作映射

| `xlib-standard` 变更类型 | 必需同步动作 | Evidence |
| --- | --- | --- |
| `docs/standard/**` 标准文本、仓库角色、分层或模块边界变更 | 更新下游 README、标准引用、DoD 和边界说明；检查旧名是否只在迁移上下文出现 | `GOWORK=off make docs-check`，必要时附 `release/standard-impact/latest.md` |
| `contracts/**`、metrics、health JSON 或 config schema 变更 | 通知所有受影响基础库更新 contract、测试和示例；breaking change 必须进入 release notes | contracts gate、下游 contract 测试、`downstream-sync-required` 结论 |
| `scripts/render_template.sh`、generator 占位符或包目录规则变更 | 重新渲染 `kernel` 和 `corekit`；确认 module path、package name、README、docs 和 contracts 无旧模板残留 | `GOWORK=off make integration` |
| Harness gate、Makefile、CI 或 `.agent/harness.yaml` 变更 | 同步下游 gate 文档和 CI 入口；强制 gate 不得在下游降级为可选 | `GOWORK=off make release-check` 或对应下游 gate 输出 |
| Evidence protocol、release manifest 字段或 artifact 规则变更 | 更新下游 Evidence 生成、校验和发布模板；manifest 字段变化必须标记同步需求 | manifest 校验、checksum、CI artifact |
| 依赖或安全策略变更 | 判断是否影响所有基础库；安全变更默认触发下游同步 | `govulncheck`、secret scan、依赖清单 |
| 命名、仓库角色或默认下游变更 | 当前主叙事必须使用 `xlib-standard`、`kernel`、`corekit`；旧名只能保留在迁移上下文 | `docs-check` 命名残留断言 |

## `kernel` 同步规则

`kernel` 是 L0 代表下游，也是默认生成验证目标。任何触达模板、generator、公共 API、contracts、Harness gate、Evidence protocol、release manifest 或仓库角色命名的变更，都必须判断是否需要同步到 `kernel`。

必须同步到 `kernel` 的情形：

- 模板渲染输出、package 目录、module path 或 README/docs 占位符发生变化。
- Standard、DoD、module boundary、Harness gate 或 Evidence 要求发生变化。
- public API、config、error、health、metrics、contracts 或 release manifest schema 发生变化。
- `release/standard-impact/latest.md` 或同类报告标记 `downstream-sync-required=true`，且 release manifest 字段 `downstream_sync_required` 为 `true`。

同步完成前，变更说明必须记录 `kernel` 的状态：已同步、无需同步并说明原因，或 blocked 并列出阻塞条件。

## L1 基础库同步规则

L1 基础库包括 `configx`、`observex` 和 `testkitx`。它们继承 L0 标准，并为其他基础库提供基础配置、观测和测试能力。

必须同步到 L1 的情形：

- Standard、Harness gate、Evidence protocol 或 security policy 发生跨库规则变化。
- `config`、metrics、health、testkit fixture 或错误分类规则发生变化。
- `kernel` 同步后发现 L1 文档、contracts 或 gate 与新标准不一致。

L1 同步不得引入 `x.go` 业务模型、生产密钥路径读取或应用 wiring。L1 可以记录 `x.go` 作为消费方，但不能依赖 `x.go`。

## L2 基础库同步规则

L2 基础库包括 `postgresx`、`redisx`、`kafkax`、`taosx`、`ossx` 和 `clickhousex`。L2 只在变更影响具体适配、contracts、release Evidence、security policy 或共享 gate 时同步。

L2 同步优先级低于 `kernel` 和 L1；若 `release/standard-impact/latest.md` 未标记 L2 影响，可以只在 release note 中记录无需同步的原因。

## `x.go` 消费方规则

x.go 仅作为基础库消费方和应用组合层。它可以根据自身需要拉取 `kernel`、L1 或 L2 基础库的新版本，但不得反向要求 `xlib-standard` 引入业务模型、业务 repository、业务消息 schema、生产密钥读取或应用 wiring。

当 `x.go` 暴露出基础库标准缺口时，应先在 `xlib-standard` 形成标准变更、ADR 或 Issue，再由标准源决定是否触发下游同步。不得把 `x.go` 的临时实现直接复制回 `xlib-standard`。

## PR 与发布要求

- 触达标准、模板、generator、Harness、Evidence、contracts、命名或下游矩阵的 PR，必须说明是否触发 `downstream-sync-required`，并记录 release manifest 的 `downstream_sync_required` 结论。
- 触发同步时，PR 或 release Evidence 必须列出 `kernel`、L1、L2 和 `x.go` 的影响结论。
- 未完成同步时，不得在完成声明中写 “无需下游动作”；必须写明 blocked 原因和后续 owner。
- `GOWORK=off make docs-check` 必须校验本文件存在、关键角色命名和旧名限制。
