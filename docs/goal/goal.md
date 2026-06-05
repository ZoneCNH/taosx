# taosx L2 基础设施适配层标准工厂 Goal 可执行方案

> Goal ID: `GOAL-20260604-TAOSX-L2-FACTORY-001`  
> 目标仓库: `github.com/ZoneCNH/taosx`  
> 标准源: `github.com/ZoneCNH/xlib-standard`  
> L0 依赖: `github.com/ZoneCNH/kernel`  
> L1 运行时契约: `configx` / `observex`，测试契约可消费 `testkitx`  
> 适配对象: TDengine / taosAdapter / taosWS / SQL / Schemaless / Health / Evidence  
> 执行协议: Goal Runtime Prompt v3.1  
> 输出日期: 2026-06-04 Asia/Tokyo  
> 目标版本建议: `taosx v0.1.0` MVA，`v0.2.0` contracts+integration，`v1.0.0` production-ready adapter

---

## 0. 执行结论

`taosx` 不应该继续作为“零散 TDengine 工具封装库”。它应该升级为 `xlib-standard` 标准源控制下的 **L2 时序数据库基础设施适配层标准工厂产物**。

最终定位固定为：

```text
xlib-standard = 标准源 / 模板 / Generator / Harness / Evidence Runtime
kernel        = L0 primitive contract
configx       = L1 显式配置与脱敏 contract
observex      = L1 日志 / 指标 / tracing / health contract
testkitx      = L1 测试夹具 / golden / contract helper
taosx         = L2 TDengine adapter，独立仓库、独立发布、独立 Evidence
x.go          = L3/L4 消费方，只组合，不反向污染 taosx
```

`taosx` 的目标不是“把 TDengine Go driver 包一层”，而是建立一个可被 `x.go`、market-data、macro-data、regime-engine、其他时序写入服务统一消费的 **TDengine adapter contract**：连接、配置、错误、生命周期、SQL、批量写入、schemaless、健康检查、可观测性、集成测试、Release Evidence 与下游采纳证据全部标准化。

---

## 1. 当前事实基线

### 1.1 仓库事实

当前 `ZoneCNH/taosx` 已经是独立 public repository，但内容处于早期状态：

```text
repo: github.com/ZoneCNH/taosx
visibility: public
default_branch: main
current README:
  # taosx
  taosx 公共基础模块
go.mod: 当前未发现
```

这意味着本 Goal 不能假设已有完整 Go module、Makefile、`.agent`、contracts、CI、release manifest 或 package API。第一步必须是 **由 xlib-standard 渲染 / 迁移为标准基础库骨架**，再进入 TDengine adapter 实现。

### 1.2 xlib-standard 对 taosx 的标准定位

当前标准源已经把 `taosx` 登记为目标库：

```text
taosx = github.com/ZoneCNH/taosx
package = taosx
layer = L2
allowed runtime deps = kernel, configx, observex
forbidden = business metric model, application time-series strategy, x.go reverse dependency
```

因此 `taosx` 的实现必须遵循：

```text
runtime dependency:
  github.com/ZoneCNH/kernel
  github.com/ZoneCNH/configx
  github.com/ZoneCNH/observex

test-only dependency:
  github.com/ZoneCNH/testkitx
  docker / testcontainers / local fake driver, only inside tests or integration harness

forbidden:
  github.com/bytechainx/x.go
  x.go/internal/*
  business schema: kline, market_data, macro_data, regime, strategy, orderbook
  production secret content
  hidden global TDengine client
  implicit /home/k8s/secrets/env/* read
```

### 1.3 TDengine 技术事实

本方案按当前 TDengine 文档将 WebSocket 连接作为默认主路径：

```text
primary connector mode: WebSocket / taosWS
native Go mode: legacy / migration-only / disabled by default
REST mode: limited SQL-only fallback, not primary adapter
schemaless: supported as dedicated writer contract
```

重要约束：

1. Go 原生连接存在迁移到 WebSocket 的官方方向，`taosx` 不应把 native 作为默认生产路径。
2. REST API 能做 SQL 写入和查询，但不覆盖参数绑定、订阅等完整能力，不应作为主 contract。
3. Schemaless 支持自动创建超级表 / 子表 / 列 / 标签，但多行写入不提供原子性保证，必须在 contract 中明确 partial failure 语义。
4. TDengine 有 STABLE / subtable / tag / timestamp precision 等特定语义，`taosx` 只能提供基础设施 contract，不得内置业务 metric model。

---

## 2. 问题的底层本质

`taosx` 真正要解决的问题不是“如何连接 TDengine”，而是：

> 如何把 TDengine 这种具体基础设施，封装成可版本化、可审计、可测试、可迁移、可复用、可由 Evidence 证明的 L2 adapter contract。

没有标准化的 `taosx`，每个上层服务都会重复解决：

```text
TDengine DSN 拼接
连接池参数
超时与 context 传递
错误分类
健康检查
SQL 执行封装
批量写入
schemaless 写入
STABLE/subtable 约定
metrics 名称
trace span
secret 脱敏
integration test
release evidence
x.go 接入验证
```

这会导致：

```text
不同服务连接方式不一致
错误分类不可统一告警
健康检查语义漂移
批量写入失败不可审计
schemaless partial failure 被误认为事务失败
secret 可能进入日志 / manifest / PR
CI 只测本地 fake，不测真实 adapter contract
release 没有 source digest / contract fingerprint
下游无法证明采用 taosx 的哪个版本
```

`taosx` 的底层价值是把这些 TDengine 相关但非业务的基础设施行为沉淀为 L2 contract，让所有上层服务复用同一套基础设施语义。

---

## 3. 不可再拆解的基本真理

| ID | 基本真理 | 对 taosx 的执行含义 |
|---|---|---|
| T-001 | L2 adapter 是基础设施边界，不是业务层 | `taosx` 可理解 STABLE/subtable，但不能内置 Kline、Symbol、Regime、MarketData |
| T-002 | 独立仓库必须独立 Evidence | `taosx` 必须有自己的 CI、manifest、release tag、checks、contract digest |
| T-003 | 没有 Evidence 不得声明 DONE | 每个 Task / Issue / Release 都必须产出 `DONE with evidence:` |
| T-004 | 标准源唯一 | 标准变更先进入 `xlib-standard`，`taosx` 只消费标准，不反向定义标准 |
| T-005 | L0/L1 契约不可绕开 | 错误、生命周期、配置、可观测、测试必须复用 `kernel/configx/observex/testkitx` |
| T-006 | TDengine driver 不是 taosx 的公共 API | 公共 API 暴露 adapter contract，不暴露底层 driver 的可变细节 |
| T-007 | WebSocket 是默认主路径 | `taosWS` profile 为默认；native / REST 只能作为 explicit compatibility mode |
| T-008 | 所有操作必须 context-first | `Exec/Query/Write/Health/Close` 必须接受或继承 `context.Context` |
| T-009 | Secret 永远不进入 Evidence | password/token 只能脱敏显示，不能进入 README、日志、manifest、PR、issue |
| T-010 | Release 是可复现状态，不是口头版本 | tag、commit、tree SHA、contract fingerprint、gate 输出必须一致 |

---

## 4. 被误认为真理的常见假设

| 常见假设 | 为什么危险 | 正确裁决 |
|---|---|---|
| `taosx` 只是 `database/sql` 的薄封装 | 无法沉淀错误、health、metrics、contract、evidence | `taosx` 是 TDengine adapter contract |
| 先能写入数据就完成 | 缺失配置、生命周期、契约、集成测试、release evidence | 写入只是其中一个 requirement |
| 业务 K 线表结构应该放进 taosx | 业务模型污染 L2 | `taosx` 只提供 generic schema builder / writer contract |
| native 连接更传统，应默认支持 | Go native 有迁移风险，且 taosc 依赖复杂 | 默认 WebSocket，native 显式开启且标记 legacy |
| REST 简单，所以默认 REST | REST 能力有限，不适合完整 adapter 主路径 | REST 可作为 SQL-only fallback，不进入默认 profile |
| 直接读取 `/home/k8s/secrets/env/*` 很方便 | 基础库不应知道部署密钥路径 | 由调用方 / configx 显式注入 |
| 测试只要 fake driver | fake 不能证明 TDengine contract | fake + golden + optional docker integration + real profile gate |
| release manifest 可以提交 | generated artifact 不应进入源码历史 | 由 `make evidence` 生成并作为 CI artifact |
| 下游登记等于采纳 | registry 只是计划态 | 只有下游仓库命令输出和 Evidence 才算 proof-based adoption |

---

## 5. 可以被打破的限制

### 5.1 不需要一次实现所有 TDengine 能力

可以分阶段：

```text
v0.1.0 = adapter skeleton + config + client contract + health + fake tests + evidence
v0.2.0 = SQL exec/query + batch writer + Docker integration + metrics contract
v0.3.0 = schemaless writer + partial failure contract + golden tests
v0.4.0 = migration helpers + STABLE/subtable builder + x.go consumer smoke
v1.0.0 = production-ready release with downstream adoption proof
```

### 5.2 不需要把业务 schema 放入 taosx

可提供 schema primitive：

```go
StableSpec
ColumnSpec
TagSpec
DatabaseSpec
TableSpec
InsertBatch
LineProtocolBatch
```

但禁止提供：

```go
KlineTable
MarketDataWriter
MacroSeriesWriter
RegimeStateTable
OrderBookSnapshotTable
```

这些应留在 `x.go` 或业务服务中。

### 5.3 不需要把 TDengine client 作为全局单例

正确方式：

```text
Factory -> Client -> explicit Start/Close
Lifecycle Manager -> explicit ownership
No package-level mutable default client
No hidden goroutine without lifecycle registration
```

### 5.4 不需要把 x.go 作为测试前提

`taosx` 必须独立验证。`x.go` 只作为 consumer smoke：

```text
taosx release gate = independent
x.go consumer gate = optional downstream adoption proof
```

---

## 6. 从零设计的新方案

### 6.1 分层位置

```text
L0: kernel
    errx / timex / lifecycx / context / shutdown / validation primitive
        ↓
L1: cross-cutting libraries
    configx / observex / testkitx / resiliencx / schedulex
        ↓
L2: infrastructure adapters
    redisx / kafkax / postgresx / taosx / ossx / clickhousex / natsx
        ↓
L3: platform integration
    x.go internal/platform, market-data-server, macro-data-server
        ↓
L4: business workflows
    strategy / regime / trading / analytics
```

### 6.2 taosx 内部架构

```text
taosx/
  .agent/
    runtime/
    harness/
    evidence/
    traceability/
    release/
    review/
    retro/
  .github/workflows/
    ci.yml
    release-check.yml
    security.yml
  cmd/
    taosx-doctor/
  pkg/taosx/
    config.go
    client.go
    factory.go
    lifecycle.go
    health.go
    errors.go
    sql.go
    writer.go
    batch.go
    schemaless.go
    schema.go
    metrics.go
    options.go
  internal/
    driver/
      database_sql.go
      fake.go
    dsn/
      dsn.go
    redact/
      redact.go
    testenv/
      docker.go
  contracts/
    config.schema.json
    health.schema.json
    metrics.contract.yaml
    public_api.snapshot
    schemaless_failure_contract.md
  docs/
    README.md
    api.md
    config.md
    health.md
    errors.md
    metrics.md
    tdengine-profile.md
    schemaless.md
    testing.md
    release.md
    downstream-adoption.md
  examples/
    basic-connect/
    health-check/
    sql-exec/
    batch-write/
    schemaless-write/
  release/
    manifest/template.json
  scripts/
    check_boundary.sh
    check_contracts.sh
    check_docs.sh
    generate_manifest.sh
  Makefile
  go.mod
  README.md
```

### 6.3 Runtime dependency policy

```text
Allowed runtime imports:
  standard library
  github.com/ZoneCNH/kernel
  github.com/ZoneCNH/configx
  github.com/ZoneCNH/observex
  TDengine Go connector selected by AutoResearch and ADR

Allowed test-only imports:
  github.com/ZoneCNH/testkitx
  testcontainers-go or docker CLI wrapper, only if accepted by ADR
  golden test helpers

Forbidden imports:
  github.com/bytechainx/x.go
  github.com/ZoneCNH/x.go
  x.go/internal/*
  business service packages
  prometheus direct SDK in public API
  OpenTelemetry direct SDK in public API
  hardcoded secret files
```

### 6.4 Public API shape

`taosx` 的 API 应以 contract 为中心：

```go
type Config struct {
    Driver          DriverMode
    Endpoint        string
    Database        string
    Username        string
    Password        Secret
    Timeout         time.Duration
    ConnectTimeout  time.Duration
    QueryTimeout    time.Duration
    WriteTimeout    time.Duration
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    TLS             TLSConfig
    Tags            map[string]string
}

type Client interface {
    Exec(ctx context.Context, stmt Statement, args ...any) (Result, error)
    Query(ctx context.Context, query Query, args ...any) (Rows, error)
    WriteBatch(ctx context.Context, batch Batch) (WriteResult, error)
    SchemalessWrite(ctx context.Context, batch SchemalessBatch) (WriteResult, error)
    Health(ctx context.Context) HealthStatus
    Close(ctx context.Context) error
}

type Factory interface {
    New(ctx context.Context, cfg Config, opts ...Option) (Client, error)
}
```

原则：

1. Public API 返回 `kernel/errx` 可识别错误。
2. Public API 不返回底层 driver 私有类型，除非通过 `Unwrap` 或 `As` 显式取得。
3. Public API 不直接暴露业务表结构。
4. `Close` 必须可重复调用且幂等。
5. `Health` 不得泄露 password / token / DSN 原文。
6. 所有 SQL 字符串 helper 必须区分 identifier 与 value，避免误导调用方把 string concat 当安全 builder。

---

## 7. taosx 能力边界

### 7.1 必须实现

```text
配置 contract
  - WebSocket DSN profile
  - username/password secret redaction
  - database/profile validation
  - timeout / pool / TLS / tags

连接 contract
  - NewClient / Factory
  - explicit Start / Close
  - ping / health
  - connection error classification

SQL contract
  - Exec
  - Query
  - Result / Rows abstraction
  - context timeout
  - row scan helper, but no ORM

Schema contract
  - CreateDatabase
  - CreateStable
  - CreateSubTableUsingStable
  - Drop / Describe helpers, optional after MVA

Write contract
  - Batch insert
  - typed values
  - timestamp precision
  - retry classification
  - partial failure capture

Schemaless contract
  - Influx line protocol
  - OpenTSDB telnet protocol
  - OpenTSDB JSON protocol
  - precision enum
  - partial failure contract

Health / observability contract
  - health JSON schema
  - metrics names
  - trace fields
  - sanitized config fingerprint

Harness / Evidence
  - fmt/vet/lint/test/race
  - contracts
  - boundary
  - docs
  - security
  - integration
  - release manifest
  - release evidence check
```

### 7.2 禁止实现

```text
禁止业务模型:
  - Kline
  - Symbol
  - Exchange
  - MarketType
  - MacroSeries
  - RegimeState
  - StrategySignal
  - OrderBook

禁止应用编排:
  - x.go service wiring
  - market-data worker
  - macro-data pipeline
  - cron/scheduler business job
  - business migration plan

禁止隐式运行时:
  - package-level default client
  - init() 建连接
  - 自动读取生产 secrets
  - hidden goroutine
  - hidden retry loop without policy
```

---

## 8. Goal Runtime v3.1 对象模型

```yaml
goal:
  id: GOAL-20260604-TAOSX-L2-FACTORY-001
  name: Upgrade taosx to xlib-standard governed L2 TDengine adapter factory
  mode: Full
  owner: ZoneCNH
  repository: github.com/ZoneCNH/taosx
  standard_source: github.com/ZoneCNH/xlib-standard
  target_release: v0.1.0
  state_machine:
    current: INIT
    target: DONE

spec:
  id: SPEC-taosx-l2-adapter-v0.1
  scope:
    - independent repo bootstrap
    - TDengine WebSocket adapter contract
    - L0/L1 governance adoption
    - Harness and Evidence runtime
    - downstream consumer proof protocol

release:
  id: REL-20260604-taosx-v0.1.0
  evidence_required: true
  done_phrase: "DONE with evidence:"
```

### 8.1 State Machine

```text
INIT
  -> CONTEXT_READY
  -> GOAL_READY
  -> SPEC_READY
  -> DESIGN_READY
  -> PLAN_READY
  -> TASKS_READY
  -> EXECUTING
  -> VERIFYING
  -> REVIEWING
  -> RELEASING
  -> RETROSPECTING
  -> DONE
```

异常状态：

```text
BLOCKED
FAILED
NEEDS_RESEARCH
NEEDS_DECISION
NEEDS_REPLAN
NEEDS_ROLLBACK
NEEDS_HUMAN_APPROVAL
INCONSISTENT_STATE
```

### 8.2 Mode 划分

| Mode | 用途 | 允许省略 | 不允许省略 |
|---|---|---|---|
| Lite | 快速 bootstrap / MVA | real TDengine integration | go.mod、README、Makefile、boundary、unit test、fake driver |
| Standard | v0.1.0 release | downstream x.go adoption | release manifest、contracts、docs、CI、health、metrics |
| Full | v1.0.0 标准工厂 | 无 | real integration、x.go consumer smoke、adoption proof、retrospective patch |

---

## 9. Requirements

### REQ-001 独立仓库标准化

`taosx` 必须成为独立 Go module，并从 `xlib-standard` 渲染或对齐标准骨架。

Acceptance Criteria:

```text
AC-001-001 go.mod module = github.com/ZoneCNH/taosx
AC-001-002 README 明确 L2 TDengine adapter 定位
AC-001-003 Makefile 包含 xlib-standard required gates
AC-001-004 .agent runtime 存在并能关联 Goal v3.1 对象
AC-001-005 release/manifest/template.json 存在
```

### REQ-002 L0/L1 契约依赖

运行时代码必须复用 L0/L1 contract。

Acceptance Criteria:

```text
AC-002-001 errors 使用 kernel errx 分类或可映射到 errx
AC-002-002 lifecycle 使用 kernel lifecycle/context/shutdown primitive
AC-002-003 config 通过 configx contract 加载或映射
AC-002-004 observability 通过 observex interface 注入
AC-002-005 tests 可消费 testkitx，但不得把 testkitx 作为必要 runtime dependency
```

### REQ-003 TDengine WebSocket 主路径

`taosx` 默认以 WebSocket / taosWS profile 作为连接方式。

Acceptance Criteria:

```text
AC-003-001 DriverMode 包含 WebSocket / NativeLegacy / RESTSQLOnly
AC-003-002 default DriverMode = WebSocket
AC-003-003 NativeLegacy 必须显式开启并在 docs 标记迁移风险
AC-003-004 RESTSQLOnly 标记为 fallback，不参与完整 contract
AC-003-005 DSN redaction golden test 覆盖 password/token 脱敏
```

### REQ-004 Config Contract

`taosx` 必须提供配置 schema、校验与脱敏 contract。

Acceptance Criteria:

```text
AC-004-001 contracts/config.schema.json 覆盖 required fields
AC-004-002 Config.Validate() 返回 validation error kind
AC-004-003 Config.Sanitized() 不泄露 secret
AC-004-004 examples/config 输出 masked password
AC-004-005 docs/config.md 说明显式注入，不读取生产 secret 路径
```

### REQ-005 Client Contract

`taosx` 必须提供稳定 Client interface。

Acceptance Criteria:

```text
AC-005-001 Client interface 覆盖 Exec/Query/WriteBatch/SchemalessWrite/Health/Close
AC-005-002 所有 I/O 方法接受 context.Context
AC-005-003 Close 幂等
AC-005-004 public API snapshot 记录导出符号
AC-005-005 API diff gate 阻断未声明 breaking change
```

### REQ-006 Error Contract

TDengine 错误必须映射到统一错误分类。

Acceptance Criteria:

```text
AC-006-001 连接失败 -> ErrorKindConnection / Unavailable
AC-006-002 认证失败 -> ErrorKindAuth
AC-006-003 超时 / context deadline -> ErrorKindTimeout
AC-006-004 SQL / schema 校验错误 -> ErrorKindValidation 或 Conflict
AC-006-005 transient network / adapter unavailable -> retryable=true
AC-006-006 docs/errors.md 包含 mapping table
```

### REQ-007 SQL Contract

SQL 执行必须提供基础 contract，但不实现 ORM。

Acceptance Criteria:

```text
AC-007-001 Exec 支持 context timeout
AC-007-002 Query 支持 rows close 和 error propagation
AC-007-003 Result 包含 rows affected 或 driver unknown 状态
AC-007-004 Statement helper 不鼓励拼接 value
AC-007-005 docs/sql.md 明确 SQL injection 边界由调用方或 prepared/param helper 处理
```

### REQ-008 Schema Contract

`taosx` 提供 generic STABLE / subtable primitive。

Acceptance Criteria:

```text
AC-008-001 StableSpec 支持 timestamp column, fields, tags
AC-008-002 Identifier validation 阻断非法库表列名
AC-008-003 SQL rendering golden test 覆盖 escaping / quoting
AC-008-004 不出现 Kline / Market / Macro / Regime 业务词汇
```

### REQ-009 Batch Write Contract

批量写入必须有可测试语义。

Acceptance Criteria:

```text
AC-009-001 Batch 包含 database/table/columns/rows/time precision
AC-009-002 WriteResult 包含 attempted/succeeded/failed/driver_result
AC-009-003 部分失败必须返回 structured error 或 partial result
AC-009-004 metrics 记录 rows_total / duration / error_total
AC-009-005 golden test 覆盖 empty batch、invalid row、large batch boundary
```

### REQ-010 Schemaless Contract

Schemaless 写入必须明确自动建表与 partial failure 风险。

Acceptance Criteria:

```text
AC-010-001 Protocol enum: Line / Telnet / JSON
AC-010-002 Precision enum: hour/min/sec/ms/us/ns
AC-010-003 docs/schemaless.md 明确自动建表、自动加列、partial failure
AC-010-004 golden test 覆盖 line escaping / value suffix / timestamp precision
AC-010-005 contract 明确不保证多行原子性
```

### REQ-011 Health Contract

健康检查必须结构化并可脱敏。

Acceptance Criteria:

```text
AC-011-001 HealthStatus JSON schema 存在
AC-011-002 Health 包含 name/status/latency/message/metadata
AC-011-003 metadata 不包含 password/token/raw DSN
AC-011-004 health 可通过 fake driver 和 optional real TDengine integration 测试
```

### REQ-012 Observability Contract

metrics/log/trace 必须供应商无关。

Acceptance Criteria:

```text
AC-012-001 observex logger/metrics/tracer 注入
AC-012-002 不直接依赖 Prometheus/Otel/Zap 具体 SDK 作为 public API
AC-012-003 metrics.contract.yaml 包含名称、类型、labels、单位
AC-012-004 logs 默认不输出 SQL values 或 secrets
AC-012-005 tracing span 包含 operation、driver、database hash、status，不包含 secret
```

### REQ-013 Test Harness

必须有 fake、contract、golden、race、optional docker integration。

Acceptance Criteria:

```text
AC-013-001 go test ./... 通过
AC-013-002 race gate 通过
AC-013-003 golden tests 覆盖 DSN、SQL render、schemaless line
AC-013-004 fake driver 可验证 client contract
AC-013-005 docker integration 可通过环境变量显式启用
```

### REQ-014 Boundary Gate

必须阻断业务词汇和非法依赖。

Acceptance Criteria:

```text
AC-014-001 check_boundary.sh 阻断 x.go imports
AC-014-002 阻断 MarketData/Kline/OrderBook/Regime/Strategy 等业务词汇
AC-014-003 阻断隐式 secret path 读取
AC-014-004 boundary gate 进入 make ci 和 release-check
```

### REQ-015 Release Evidence

每次发布必须生成独立 Evidence。

Acceptance Criteria:

```text
AC-015-001 make evidence 生成 release/manifest/latest.json
AC-015-002 release-evidence-check 校验 manifest 与仓库事实一致
AC-015-003 release-final-check 要求 clean workspace
AC-015-004 CI 上传 manifest 和 sha256 artifact
AC-015-005 完成声明包含 DONE with evidence:
```

### REQ-016 Downstream Adoption

`taosx` 发布后必须形成下游采纳证据协议。

Acceptance Criteria:

```text
AC-016-001 docs/downstream-adoption.md 定义 x.go consumer smoke
AC-016-002 adoption proof 区分 registered / baseline_scanned / adopted
AC-016-003 没有 x.go 当前命令输出不得宣称 adopted
AC-016-004 downstream proof 包含 source commit、downstream commit、gate outputs、rollback
```

---

## 10. Traceability Matrix

| Requirement | Acceptance Criteria | Design Section | Tasks | Tests | Evidence |
|---|---|---|---|---|---|
| REQ-001 | AC-001-* | Repo Bootstrap | TASK-001 | doctor, go test | manifest, README |
| REQ-002 | AC-002-* | L0/L1 Contract | TASK-002 | boundary, unit | dependency list |
| REQ-003 | AC-003-* | TDengine Profile | TASK-003 | config golden | ADR, docs |
| REQ-004 | AC-004-* | Config | TASK-004 | schema, redaction | config.schema.json |
| REQ-005 | AC-005-* | Client API | TASK-005 | api snapshot | public_api.snapshot |
| REQ-006 | AC-006-* | Error Mapping | TASK-006 | error contract | docs/errors.md |
| REQ-007 | AC-007-* | SQL | TASK-007 | fake + golden | sql test output |
| REQ-008 | AC-008-* | Schema Builder | TASK-008 | golden SQL | contracts/schema |
| REQ-009 | AC-009-* | Batch Writer | TASK-009 | batch tests | metrics evidence |
| REQ-010 | AC-010-* | Schemaless | TASK-010 | line/json/telnet golden | schemaless contract |
| REQ-011 | AC-011-* | Health | TASK-011 | health schema | health.schema.json |
| REQ-012 | AC-012-* | Observability | TASK-012 | metrics contract | metrics.contract.yaml |
| REQ-013 | AC-013-* | Harness | TASK-013 | unit/race/docker | CI logs |
| REQ-014 | AC-014-* | Boundary | TASK-014 | boundary gate | boundary output |
| REQ-015 | AC-015-* | Release | TASK-015 | release-final-check | manifest + sha256 |
| REQ-016 | AC-016-* | Adoption | TASK-016 | consumer smoke | adoption proof |

---

## 11. Task Breakdown

### TASK-001 Repo Bootstrap from xlib-standard

```yaml
task_id: TASK-GOAL-20260604-TAOSX-001
title: Bootstrap taosx as xlib-standard generated repository
mode: Standard
commands:
  - git clone git@github.com:ZoneCNH/taosx.git
  - cd taosx
  - git switch main
  - git pull --ff-only
  - git worktree add .worktree/goal-taosx-l2-factory -b goal/taosx-l2-factory
  - cd .worktree/goal-taosx-l2-factory
  - render or copy xlib-standard template into empty repo
  - go mod init github.com/ZoneCNH/taosx
  - GOWORK=off go mod tidy
acceptance:
  - go.mod exists
  - README updated
  - Makefile exists
  - .agent exists
  - release manifest template exists
evidence:
  - git diff summary
  - GOWORK=off make test
  - GOWORK=off make docs-check
```

### TASK-002 L0/L1 Dependency Contract

```yaml
task_id: TASK-GOAL-20260604-TAOSX-002
title: Wire kernel/configx/observex/testkitx contracts
runtime_allowed:
  - github.com/ZoneCNH/kernel
  - github.com/ZoneCNH/configx
  - github.com/ZoneCNH/observex
test_allowed:
  - github.com/ZoneCNH/testkitx
acceptance:
  - dependency-check passes
  - boundary-check passes
  - no x.go import
```

### TASK-003 TDengine Connector AutoResearch ADR

```yaml
task_id: TASK-GOAL-20260604-TAOSX-003
title: Decide TDengine Go connector mode and driver policy
autoresearch_questions:
  - current official Go connector module path
  - taosWS driver registration name
  - DSN format
  - websocket vs native support matrix
  - schemaless API shape
  - error code mapping API
  - docker image and CI service strategy
outputs:
  - docs/adr/ADR-20260604-001-tdengine-connector-mode.md
  - docs/tdengine-profile.md
acceptance:
  - WebSocket default documented
  - native marked legacy / explicit
  - REST marked SQL-only fallback
```

### TASK-004 Config and Redaction Contract

```yaml
task_id: TASK-GOAL-20260604-TAOSX-004
title: Implement Config, Validate, Sanitized, schema
files:
  - pkg/taosx/config.go
  - contracts/config.schema.json
  - docs/config.md
  - examples/config/main.go
acceptance:
  - invalid endpoint rejected
  - empty username rejected if auth required
  - password redacted in String/JSON/log fields
  - config schema tests pass
```

### TASK-005 Client Interface and Factory

```yaml
task_id: TASK-GOAL-20260604-TAOSX-005
title: Implement public Client/Factory contract
files:
  - pkg/taosx/client.go
  - pkg/taosx/factory.go
  - pkg/taosx/options.go
  - contracts/public_api.snapshot
acceptance:
  - API snapshot generated
  - fake client implements interface
  - no driver-specific public leakage
```

### TASK-006 Error Mapping

```yaml
task_id: TASK-GOAL-20260604-TAOSX-006
title: Map TDengine and context errors to kernel errx
files:
  - pkg/taosx/errors.go
  - docs/errors.md
  - contracts/error-golden.json
acceptance:
  - timeout maps to Timeout
  - auth maps to Auth
  - connection refused maps to Connection/Unavailable
  - validation maps to Validation
  - retryable flag documented
```

### TASK-007 SQL Exec/Query Contract

```yaml
task_id: TASK-GOAL-20260604-TAOSX-007
title: Implement SQL Exec/Query primitives
files:
  - pkg/taosx/sql.go
  - internal/driver/database_sql.go
  - docs/sql.md
acceptance:
  - context cancellation propagated
  - rows close tested
  - result rows affected represented safely
  - no ORM semantics
```

### TASK-008 Schema Primitive

```yaml
task_id: TASK-GOAL-20260604-TAOSX-008
title: Implement STABLE/subtable schema builder
files:
  - pkg/taosx/schema.go
  - contracts/schema-golden/*.sql
  - docs/schema.md
acceptance:
  - CreateStable SQL golden pass
  - identifier validation pass
  - no business schema included
```

### TASK-009 Batch Writer

```yaml
task_id: TASK-GOAL-20260604-TAOSX-009
title: Implement batch write contract
files:
  - pkg/taosx/batch.go
  - pkg/taosx/writer.go
  - docs/batch-write.md
acceptance:
  - empty batch validation
  - row length mismatch validation
  - partial result structure
  - metrics emitted through observex
```

### TASK-010 Schemaless Writer

```yaml
task_id: TASK-GOAL-20260604-TAOSX-010
title: Implement schemaless write contract
files:
  - pkg/taosx/schemaless.go
  - contracts/schemaless_failure_contract.md
  - docs/schemaless.md
acceptance:
  - line/telnet/json protocol enum
  - timestamp precision enum
  - partial failure documented
  - golden examples pass
```

### TASK-011 Health and Lifecycle

```yaml
task_id: TASK-GOAL-20260604-TAOSX-011
title: Implement health check and lifecycle ownership
files:
  - pkg/taosx/health.go
  - pkg/taosx/lifecycle.go
  - contracts/health.schema.json
acceptance:
  - Health JSON schema pass
  - Close idempotent
  - Start/Stop order explicit
  - no hidden background runtime
```

### TASK-012 Observability

```yaml
task_id: TASK-GOAL-20260604-TAOSX-012
title: Implement vendor-neutral logs/metrics/tracing
files:
  - pkg/taosx/metrics.go
  - contracts/metrics.contract.yaml
  - docs/metrics.md
acceptance:
  - no direct prometheus/otel public API
  - metrics names documented
  - secret labels forbidden
```

### TASK-013 Harness and CI

```yaml
task_id: TASK-GOAL-20260604-TAOSX-013
title: Add required and extended gates
commands:
  - GOWORK=off make fmt
  - GOWORK=off make vet
  - GOWORK=off make lint
  - GOWORK=off make test
  - GOWORK=off make race
  - GOWORK=off make boundary
  - GOWORK=off make security
  - GOWORK=off make contracts
  - GOWORK=off make docs-check
  - GOWORK=off make integration
acceptance:
  - GitHub Actions run gates
  - local Makefile mirrors CI
```

### TASK-014 Boundary and Security

```yaml
task_id: TASK-GOAL-20260604-TAOSX-014
title: Enforce L2 boundary and secret policy
acceptance:
  - x.go import blocked
  - business terms blocked
  - raw password not logged
  - release manifest excludes generated secrets
  - /home/k8s/secrets/env/* only appears as documentation path, never content
```

### TASK-015 Release Evidence

```yaml
task_id: TASK-GOAL-20260604-TAOSX-015
title: Generate independent release manifest and evidence
commands:
  - CHECK_STATUS=passed GOWORK=off make evidence
  - RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check
  - XLIB_CONTEXT=release_verify GOWORK=off make release-final-check
acceptance:
  - release/manifest/latest.json generated but not committed
  - latest.json.sha256 generated
  - manifest includes module, commit, tree_sha, contracts, dependencies, tools, checks, workflow, score
```

### TASK-016 Downstream Consumer Proof

```yaml
task_id: TASK-GOAL-20260604-TAOSX-016
title: Define x.go and service consumer adoption proof
files:
  - docs/downstream-adoption.md
  - contracts/downstream-adoption-proof.schema.json
acceptance:
  - registered != adopted rule preserved
  - proof requires downstream commit and gate outputs
  - x.go smoke never becomes taosx release prerequisite unless explicitly scoped
```

### TASK-017 Retrospective and Self-improving Patch

```yaml
task_id: TASK-GOAL-20260604-TAOSX-017
title: Convert lessons into prompt/harness/rule patches
outputs:
  - docs/retro/RETRO-20260604-taosx-v0.1.0.md
  - docs/retro/PATCH-PROMPT-20260604-taosx.md
  - docs/retro/PATCH-HARNESS-20260604-taosx.md
  - docs/retro/PATCH-RULE-20260604-taosx.md
acceptance:
  - failed gates become future gates or docs
  - TDengine connector decisions update ADR
  - downstream adoption gaps logged
```

---

## 12. PR / Issue 执行波次

### Wave 0: Issue Registry and Goal Intake

```text
Issue: GOAL-20260604-TAOSX-L2-FACTORY-001
Labels: goal, taosx, l2-adapter, xlib-standard, evidence-required
Output: goal.md, spec.md, design.md, plan.md, traceability.md
Gate: semantic review
```

### Wave 1: Bootstrap PR

```text
PR-001: bootstrap taosx standard skeleton
Scope:
  - go.mod
  - README
  - Makefile
  - .agent
  - docs standard links
  - release manifest template
  - boundary/security scripts
Gate:
  - make test
  - make docs-check
  - make boundary
```

### Wave 2: Contract PR

```text
PR-002: config/client/error/health public contracts
Scope:
  - Config
  - Client interface
  - Error mapping
  - Health status
  - public API snapshot
Gate:
  - contracts
  - api-check
  - golden
```

### Wave 3: Runtime Adapter PR

```text
PR-003: TDengine WebSocket adapter runtime
Scope:
  - taosWS connector ADR
  - SQL Exec/Query
  - batch writer
  - fake driver
  - optional real driver behind build tag or integration profile
Gate:
  - unit
  - race
  - fake integration
  - optional docker integration
```

### Wave 4: Schemaless and Schema PR

```text
PR-004: schema builder and schemaless writer
Scope:
  - StableSpec
  - SubTableSpec
  - SchemalessBatch
  - partial failure contract
  - golden examples
Gate:
  - golden
  - docs-check
  - contract tests
```

### Wave 5: Release and Adoption PR

```text
PR-005: release evidence and downstream adoption protocol
Scope:
  - release manifest
  - CI artifact
  - docs/downstream-adoption.md
  - x.go smoke plan
  - retrospective templates
Gate:
  - release-final-check
  - score --min 9.8
```

---

## 13. Harness Gate 设计

### 13.1 Required Gate

```makefile
fmt:
	GOWORK=off go fmt ./...

vet:
	GOWORK=off go vet ./...

lint:
	golangci-lint run ./...

test:
	GOWORK=off go test ./...

race:
	GOWORK=off go test -race ./...

boundary:
	./scripts/check_boundary.sh

security:
	./scripts/check_secrets.sh

contracts:
	./scripts/check_contracts.sh

docs-check:
	./scripts/check_docs.sh

integration:
	./scripts/check_integration.sh

evidence:
	./scripts/generate_manifest.sh

release-check:
	$(MAKE) ci
	$(MAKE) evidence
	$(MAKE) release-evidence-check

release-final-check:
	$(MAKE) release-clean-check
	$(MAKE) release-check
	$(MAKE) release-clean-check
```

### 13.2 Extended Gate

```text
property: schema builder property tests
fuzz-smoke: DSN/parser/schemaless fuzz smoke
golden: SQL render / config redaction / metrics contract golden
integration-real: TDengine docker / taosAdapter optional real contract
consumer-smoke: x.go or sample app import smoke
score: goalcli score --min 9.8
```

### 13.3 Boundary Gate 规则

必须 fail-fast：

```text
forbidden imports:
  github.com/bytechainx/x.go
  github.com/ZoneCNH/x.go
  x.go/internal

forbidden business terms in runtime packages:
  BTCUSDT
  ETHUSDT
  Kline
  OrderBook
  MarketData
  MacroData
  Regime
  Strategy
  Position
  TradingSignal

forbidden secret behavior:
  os.ReadFile("/home/k8s/secrets/env/...") in runtime
  raw password in README examples
  raw DSN in manifest
```

---

## 14. Evidence Protocol

### 14.1 Task 完成声明

每个 Task 必须使用：

```text
DONE with evidence:
- scope: task
- task_id: TASK-GOAL-20260604-TAOSX-XXX
- branch: goal/taosx-l2-factory
- commit: <sha or not committed>
- gates:
  - GOWORK=off make test: passed <summary>
  - GOWORK=off make boundary: passed <summary>
- artifacts:
  - <path>: <purpose>
- known gaps:
  - <none or explicit blocker>
```

### 14.2 Issue 完成声明

```text
DONE with evidence:
- scope: issue
- issue_id: ISSUE-TAOSX-XXX
- pr: <url or number>
- commit: <sha>
- gates:
  - fmt/vet/lint/test/race: passed
  - contracts: passed
  - docs-check: passed
  - boundary: passed
- artifacts:
  - release/manifest/latest.json: local generated, not committed
  - release/manifest/latest.json.sha256: checksum
- review:
  - reviewer: <name>
  - result: approved/request-changes
- known gaps:
  - <none>
```

### 14.3 Release 完成声明

```text
DONE with evidence:
- scope: release
- release_id: REL-20260604-taosx-v0.1.0
- tag: v0.1.0
- commit: <sha>
- source_digest: <manifest.source_digest>
- contract_fingerprint: <manifest.contracts.sha256>
- dependency_list: <manifest.dependencies>
- tool_versions: <manifest.tools>
- workflow_artifact:
  - release/manifest/latest.json
  - release/manifest/latest.json.sha256
- gates:
  - GOWORK=off make release-final-check: passed
  - GOWORK=off go run ./cmd/goalcli score --min 9.8: passed
- downstream:
  - adoption_status: not_claimed | adopted
  - proof_based_adoption: false | true
  - x.go consumer smoke: passed | not_run | blocked
- known gaps:
  - <none or explicit blocker>
```

---

## 15. Release Manifest 必须字段

```json
{
  "module": "github.com/ZoneCNH/taosx",
  "version": "v0.1.0",
  "commit": "<HEAD>",
  "tree_sha": "<tree>",
  "source_digest": "sha256:<...>",
  "tracked_file_count": 0,
  "go_version": "<go version>",
  "generated_at": "<timestamp>",
  "generated_by": "taosx release manifest generator",
  "tree_state": "clean|dirty",
  "checks": {
    "fmt": "passed",
    "vet": "passed",
    "lint": "passed",
    "test": "passed",
    "race": "passed",
    "boundary": "passed",
    "security": "passed",
    "contracts": "passed",
    "docs": "passed",
    "integration": "passed|blocked"
  },
  "contracts": {
    "config_schema_sha256": "<sha>",
    "health_schema_sha256": "<sha>",
    "metrics_contract_sha256": "<sha>",
    "public_api_snapshot_sha256": "<sha>"
  },
  "dependencies": [],
  "tools": {},
  "standard_source": {
    "repo": "github.com/ZoneCNH/xlib-standard",
    "commit": "<source commit or recorded version>",
    "downstream_sync_required": false
  },
  "workflow": {
    "workflow_run_id": "local or CI id",
    "artifact_name": "taosx-release-manifest",
    "artifact_url": "local:<path> or CI artifact URL"
  },
  "score": {
    "threshold": 9.8,
    "actual": 0,
    "status": "passed|failed|blocked"
  }
}
```

---

## 16. API / Contract 设计细节

### 16.1 Config

```go
type DriverMode string

const (
    DriverWebSocket DriverMode = "websocket"
    DriverNativeLegacy DriverMode = "native_legacy"
    DriverRESTSQLOnly DriverMode = "rest_sql_only"
)

type Config struct {
    Driver          DriverMode
    Endpoint        string
    Database        string
    Username        string
    Password        Secret
    ConnectTimeout  time.Duration
    QueryTimeout    time.Duration
    WriteTimeout    time.Duration
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    TLS             TLSConfig
    Metadata        map[string]string
}
```

Validation rules:

```text
Driver required, default websocket
Endpoint required
Username required unless auth disabled explicitly for test fake
Password secret must sanitize
Database optional for admin operations, required for database-bound writer
Timeouts must be >0 or defaulted
MaxOpenConns >= 0
MaxIdleConns >= 0
MaxIdleConns <= MaxOpenConns when MaxOpenConns > 0
```

### 16.2 Error

```text
ErrInvalidConfig      -> Validation
ErrInvalidIdentifier  -> Validation
ErrConnect            -> Connection / Unavailable
ErrAuth               -> Auth
ErrTimeout            -> Timeout
ErrCanceled           -> Canceled
ErrSQL                -> Internal or Validation depending SQL state
ErrPartialWrite       -> Conflict or Internal with partial result
ErrUnsupportedDriver  -> Validation
```

### 16.3 Health

Health output：

```json
{
  "name": "taosx",
  "status": "healthy|degraded|unhealthy",
  "message": "ok",
  "checked_at": "2026-06-04T00:00:00Z",
  "latency_ms": 12,
  "metadata": {
    "driver": "websocket",
    "endpoint_hash": "sha256:...",
    "database": "market_ts"
  }
}
```

禁止：

```text
password
raw DSN
token
full endpoint with credentials
```

### 16.4 Metrics Contract

```yaml
metrics:
  - name: taosx_client_connect_total
    type: counter
    labels: [driver, status]
  - name: taosx_client_connection_errors_total
    type: counter
    labels: [driver, error_kind]
  - name: taosx_query_duration_seconds
    type: histogram
    labels: [operation, driver, status]
  - name: taosx_write_rows_total
    type: counter
    labels: [operation, driver, status]
  - name: taosx_write_batches_total
    type: counter
    labels: [operation, driver, status]
  - name: taosx_health_status
    type: gauge
    labels: [driver, status]
```

禁止 labels：

```text
password
raw_dsn
sql_text
raw_query
raw_table_name if business-sensitive
token
secret_path
```

### 16.5 Schemaless partial failure

Contract 必须明确：

```text
Input batch has N records.
TDengine schemaless may succeed partially.
WriteResult must record attempted=N.
If driver exposes affected rows or failed row index, taosx records it.
If driver only returns generic error, taosx returns ErrPartialWrite with unknown succeeded count.
Retry policy must not blindly retry non-idempotent unknown partial batch unless caller opts in.
```

---

## 17. AutoResearch Protocol

### 17.1 触发条件

以下问题必须进入 AutoResearch，不得凭记忆硬编码：

```text
TDengine Go connector 当前 module path / latest supported mode
WebSocket driver name and DSN syntax
Native Go connector deprecation status
REST capability limitations
Schemaless API shape
TDengine error code API
Docker image / test environment startup
License compatibility
Connection pool behavior under database/sql
Cloud vs self-hosted WebSocket endpoint differences
```

### 17.2 AutoResearch 输出

```text
docs/research/RESEARCH-20260604-tdengine-go-connector.md
  - source links
  - connector versions
  - driver names
  - API snippets
  - decisions
  - unresolved questions

docs/adr/ADR-20260604-001-tdengine-driver-mode.md
  - decision: WebSocket default
  - alternatives: native, REST
  - consequences
  - migration policy
```

### 17.3 决策门禁

未完成 AutoResearch 前，禁止：

```text
固定 Go connector import path
固定 driver name
固定 native/REST 支持等级
宣称 real integration passed
发布 v1.0.0
```

---

## 18. Risk Register

| Risk ID | 风险 | 严重级别 | 缓解 |
|---|---|---|---|
| RISK-001 | 当前 taosx 仓库几乎为空 | High | 先 bootstrap skeleton，再做 runtime |
| RISK-002 | TDengine driver API 变化 | High | AutoResearch + ADR + api compatibility guard |
| RISK-003 | native 连接废弃导致迁移风险 | High | WebSocket default，native legacy explicit |
| RISK-004 | secret 泄露到日志 / manifest | Critical | SecretString + redaction golden + security gate |
| RISK-005 | 业务 schema 下沉到 L2 | High | boundary terms + review gate |
| RISK-006 | fake tests 通过但真实 TDengine 不可用 | Medium | optional docker integration + separate evidence state |
| RISK-007 | partial write 被误重试 | High | WriteResult + ErrPartialWrite + retry policy docs |
| RISK-008 | public API 过早稳定 | Medium | maturity map: experimental -> candidate -> stable |
| RISK-009 | x.go 反向污染标准 | High | no x.go import + consumer-only docs |
| RISK-010 | release evidence 被提交 | Medium | .gitignore + release-evidence-check |

---

## 19. Decision Log

| Decision ID | 决策 | 状态 | 理由 |
|---|---|---|---|
| DEC-20260604-001 | `taosx` 定位为 L2 TDengine adapter | Accepted | xlib-standard downstream matrix 已定义 |
| DEC-20260604-002 | 默认 WebSocket / taosWS | Accepted | 当前官方方向更适合兼容性 |
| DEC-20260604-003 | Native 仅 legacy explicit | Accepted | 降低未来迁移风险 |
| DEC-20260604-004 | REST 仅 SQL-only fallback | Accepted | REST 能力边界较窄 |
| DEC-20260604-005 | 不内置业务 schema | Accepted | 保持 L2 边界 |
| DEC-20260604-006 | fake + optional real integration 双层测试 | Accepted | MVA 低成本启动，Full 模式增强可信度 |
| DEC-20260604-007 | 下游采纳不得由 registry 推断 | Accepted | 必须 proof-based adoption |

---

## 20. Rollback Protocol

### 20.1 Task Rollback

```text
如果单个 Task 引入错误：
1. revert task commit
2. 保留 failed evidence
3. 更新 risk register
4. 重新执行对应 gates
5. 不删除失败记录
```

### 20.2 Release Rollback

```text
如果 v0.1.0 release 后发现 contract 错误：
1. 标记 release notes known issue
2. 创建 patch issue
3. 发布 v0.1.1 修复
4. 若为 breaking public API，撤回 stable 级别，标记 candidate/experimental
5. 通知 downstream adoption owners
```

### 20.3 Downstream Rollback

```text
如果 x.go 或服务接入 taosx 失败：
1. 下游回滚 go.mod 到前一版本
2. taosx 保留 compatibility issue
3. adoption proof 标记 failed / rolled_back
4. 不把 failed adoption 改写为 not_run
```

---

## 21. DoD 分层

### Task DoD

```text
- Task scope 完成
- 对应测试通过
- 对应文档更新
- Evidence 输出存在
- known gaps 明确
```

### Issue DoD

```text
- 所有关联 Task 完成
- Traceability Matrix 更新
- PR review 完成
- required gates passed
- failed/blocked evidence 保留
```

### Goal DoD

```text
- taosx 独立 Go module 可编译
- L0/L1 contract 对齐
- TDengine WebSocket default contract 完成
- fake + contract + golden tests 完成
- optional real integration 有明确状态
- release manifest 可生成和校验
- downstream adoption protocol 完成
```

### Release DoD

```text
- tag 创建
- release-final-check passed
- manifest + sha256 artifact
- public API snapshot
- contract fingerprint
- source digest
- dependency list
- tool versions
- release notes
- DONE with evidence:
```

### Retrospective DoD

```text
- 记录哪些 gate 发现问题
- 记录 TDengine connector 决策是否仍有效
- 输出 Prompt Patch
- 输出 Harness Patch
- 输出 Rule Patch
- 输出下一轮 Issue candidates
```

---

## 22. 可复利增长的系统架构

`taosx` 的复利不是来自多写功能，而是来自以下可迁移资产：

```text
1. TDengine adapter contract
   -> 复用于 market-data / macro-data / metrics ingestion

2. Config schema + redaction
   -> 复用于所有 L2 infra adapters

3. Error mapping table
   -> 复用于 alert / retry / circuit breaker

4. Health schema
   -> 复用于 dashboard / deployment check

5. Metrics contract
   -> 复用于 observex provider adapters

6. Golden SQL/schema tests
   -> 复用于 future migration and compatibility

7. Release Evidence
   -> 复用于 downstream adoption and audit

8. AutoResearch ADR
   -> 复用于 driver upgrade decisions

9. Retrospective patches
   -> 反哺 xlib-standard / Harness / Rule registry
```

复利闭环：

```text
taosx implementation
  -> gates find drift
  -> evidence records drift
  -> retro converts drift to rule
  -> xlib-standard updates standard/harness
  -> future L2 adapters inherit stronger gate
  -> downstream adoption becomes cheaper and safer
```

---

## 23. AI / 自动化 / 研究增强介入点

### 23.1 gstack

```text
Goal Stack:
G0: taosx standard factory target
G1: repo bootstrap
G2: L0/L1 contract adoption
G3: TDengine adapter runtime
G4: Harness/Evidence
G5: downstream adoption
G6: self-improving
```

### 23.2 superpowers

```text
- Deep repository diff analysis
- Driver API research
- Contract generation
- Boundary rule generation
- Golden test generation
- Release manifest verification
- PR review checklist generation
```

### 23.3 Harness

```text
- Required gates as machine judge
- Extended gates for confidence
- Release gates for evidence
- Boundary gates for layer purity
- Secret gates for safety
```

### 23.4 Compound Engineering

```text
每次实现不只交付代码，还交付：
- contract
- test
- evidence
- docs
- rule patch
- future generator improvement
```

### 23.5 Self-improving

```text
失败模式 -> Retrospective -> Prompt Patch / Harness Patch / Rule Patch -> xlib-standard standard update
```

### 23.6 AutoResearch

```text
TDengine version / driver / DSN / schemaless / error mapping / Docker integration 一律 research-first
```

### 23.7 Goal-Oriented Thinking

```text
所有 PR 都反查：
- 是否推进 Goal?
- 是否有 Evidence?
- 是否可 rollback?
- 是否防止未来漂移?
```

---

## 24. 最小可行行动 MVA

### MVA 目标

在最短路径内把 `taosx` 从早期仓库变为 **可执行、可验证、可发布的 L2 adapter skeleton**。

### MVA 范围

必须做：

```text
1. go.mod: github.com/ZoneCNH/taosx
2. README: L2 TDengine adapter 定位
3. Makefile: fmt/vet/test/boundary/contracts/docs/evidence/release-check
4. .agent: Goal v3.1 最小工件
5. pkg/taosx:
   - Config
   - Client interface
   - FakeClient
   - Health
   - Error mapping skeleton
   - Secret redaction
6. contracts:
   - config.schema.json
   - health.schema.json
   - public_api.snapshot
7. docs:
   - api.md
   - config.md
   - health.md
   - errors.md
   - release.md
8. scripts:
   - boundary
   - contracts
   - docs-check
   - generate_manifest
9. tests:
   - config validation
   - redaction golden
   - fake health
   - public API snapshot
10. release evidence:
   - make evidence
   - release-evidence-check
```

暂不做：

```text
- 真实 TDengine Docker integration as required gate
- x.go adoption claim
- full schemaless writer
- full migration helper
- v1.0 stable compatibility guarantee
```

MVA 完成声明：

```text
DONE with evidence:
- scope: release-mva
- version: v0.1.0
- gates: fmt/vet/test/boundary/contracts/docs/evidence/release-evidence-check
- known gaps:
  - real TDengine integration optional/not_run
  - x.go adoption not_claimed
```

---

## 25. 1 天行动计划

### Day 1 目标

完成标准骨架和 MVA contract。

```text
Hour 1-2:
  - 创建 worktree
  - bootstrap xlib-standard skeleton
  - go.mod / README / Makefile

Hour 3-4:
  - Config / Secret / Validate
  - Client interface / FakeClient
  - HealthStatus

Hour 5-6:
  - contracts/config.schema.json
  - contracts/health.schema.json
  - public API snapshot
  - docs/api.md docs/config.md docs/health.md

Hour 7-8:
  - boundary/security/docs scripts
  - tests + golden
  - make test / boundary / contracts / docs-check

End of Day:
  - generate local evidence
  - open PR-001 bootstrap
  - known gaps: real TDengine integration not_run
```

Day 1 成功标准：

```text
GOWORK=off make test: passed
GOWORK=off make boundary: passed
GOWORK=off make contracts: passed
GOWORK=off make docs-check: passed
CHECK_STATUS=passed GOWORK=off make evidence: passed
```

---

## 26. 7 天行动计划

### Day 1: Bootstrap

```text
完成 MVA skeleton + fake client + evidence
```

### Day 2: TDengine AutoResearch + ADR

```text
确认 Go connector module path
确认 taosWS DSN
确认 schemaless API
确认 docker integration approach
输出 ADR
```

### Day 3: Real SQL Adapter

```text
实现 database/sql adapter
Exec / Query / Rows / Result
context timeout
error mapping first version
```

### Day 4: Schema + Batch Writer

```text
StableSpec
SubTableSpec
Batch
WriteResult
SQL render golden
```

### Day 5: Schemaless Writer

```text
Line/Telnet/JSON enum
Precision enum
Partial failure contract
Golden tests
```

### Day 6: Integration + Observability

```text
optional docker TDengine gate
metrics contract
observex integration
health check against real/fake
```

### Day 7: Release Candidate

```text
release-final-check
score --min 9.8
release notes
v0.1.0 or v0.2.0 tag
retrospective patches
```

7 天成功标准：

```text
- taosx independent repo standards complete
- WebSocket adapter candidate implemented
- fake + golden + optional real integration state explicit
- release manifest generated and verified
- no x.go adoption overclaim
```

---

## 27. 30 天行动计划

### Week 1: v0.1.0 MVA

```text
- standard skeleton
- Config/Client/Health/Error
- fake tests
- release evidence
```

### Week 2: v0.2.0 Adapter Candidate

```text
- real WebSocket SQL adapter
- batch writer
- schema builder
- metrics contract
- docker integration optional gate
```

### Week 3: v0.3.0 Schemaless + Compatibility

```text
- schemaless writer
- partial failure contract
- connector upgrade ADR
- API diff gate
- fuzz/golden expansion
```

### Week 4: v1.0.0 Readiness

```text
- x.go consumer smoke branch
- downstream adoption proof schema
- performance baseline
- release-final-check clean
- score >= 9.8
- retrospective patches submitted to xlib-standard if reusable
```

30 天成功标准：

```text
taosx 可以作为 x.go / market-data / macro-data 的 TDengine adapter foundation，且每个 release 都能用独立 Evidence 证明。
```

---

## 28. 衡量指标

### Engineering Metrics

```text
required_gates_pass_rate >= 100%
release_final_check_pass = true
boundary_violations = 0
secret_findings = 0
public_api_undocumented_changes = 0
contract_coverage >= 95%
```

### Adapter Metrics

```text
connect_success_rate
query_latency_p50/p95/p99
write_batch_latency_p50/p95/p99
rows_written_per_second
partial_write_error_rate
health_check_latency
connection_error_rate
retryable_error_rate
```

### Governance Metrics

```text
evidence_completeness_score >= 9.8
traceability_coverage = 100%
requirements_without_tests = 0
requirements_without_evidence = 0
adoption_claims_without_proof = 0
retro_patch_created_per_release >= 1
```

### Downstream Metrics

```text
x.go consumer smoke status: not_run | blocked | passed
number_of_consumers_using_tagged_release
downstream rollback count
downstream integration time
breaking change count
```

---

## 29. 迭代优化机制

### 29.1 每次 PR 后

```text
- 检查 failed gate
- 分类是 code bug / contract gap / harness gap / doc gap / research gap
- 更新 risk register
- 必要时生成 Patch-Harness 或 Patch-Rule
```

### 29.2 每次 Release 后

```text
- 生成 retrospective
- 记录 TDengine connector API 是否变化
- 记录 release evidence 是否完整
- 记录 downstream adoption 是否真实发生
- 输出下一轮 issues
```

### 29.3 每次 xlib-standard 更新后

```text
- 运行 downstream sync impact check
- 判断 taosx 是否需要同步
- 如需同步，开独立 issue/PR
- 不直接把 xlib-standard 变化混入业务功能 PR
```

### 29.4 每次 TDengine 版本更新后

```text
- AutoResearch connector release note
- 更新 compatibility matrix
- 跑 docker integration
- 更新 ADR if driver mode changes
- 若 breaking，发布 taosx patch/minor
```

---

## 30. x.go / 下游采纳协议

`taosx` 可以被 `x.go` 消费，但不能被 `x.go` 反向污染。

### 30.1 Consumer Smoke 范围

```text
x.go branch:
  goal/adopt-taosx-v0.1.0

allowed changes:
  go.mod require github.com/ZoneCNH/taosx v0.1.0
  internal/platform/taos adapter wiring
  config injection from application layer
  health smoke
  no business schema moved into taosx
```

### 30.2 Adoption Proof 必需字段

```yaml
source_repo: github.com/ZoneCNH/taosx
source_version: v0.1.0
source_commit: <sha>
downstream_repo: github.com/bytechainx/x.go
downstream_commit: <sha>
mode: consumer_smoke
gate_outputs:
  - command: GOWORK=off go test ./...
    status: passed|failed|blocked
    artifact_path: <path>
    sha256: <sha>
rollback:
  command: go get github.com/ZoneCNH/taosx@<previous>
  owner: <owner>
adoption_status: adopted|not_adopted|blocked
```

禁止声明：

```text
registered == adopted
PR opened == adopted
go.mod changed == adopted
dry-run passed == production adopted
```

---

## 31. 安全与密钥策略

`taosx` 必须把密钥视为调用方责任，只接受显式注入。

允许：

```text
Config.Password = SecretString from caller
Config.Sanitized() output masked
Docs mention /home/k8s/secrets/env/* as deployment convention only
Examples use placeholder: ${TAOS_PASSWORD}
```

禁止：

```text
runtime 自动读取 /home/k8s/secrets/env/*
README 写真实密码
tests 输出真实 DSN
manifest 写 password/token
PR 描述粘贴生产配置
logs 输出 raw query containing secrets
```

Gate：

```text
make security
check_secrets.sh
redaction golden tests
manifest secret scan
PR template secret checklist
```

---

## 32. 完整执行命令骨架

```bash
# 0. 准备
cd ~/code
git clone git@github.com:ZoneCNH/taosx.git
cd taosx
git switch main
git pull --ff-only

# 1. worktree-only 开发
git worktree add .worktree/goal-taosx-l2-factory -b goal/taosx-l2-factory
cd .worktree/goal-taosx-l2-factory

# 2. 从 xlib-standard 渲染或迁移骨架
# 方式 A：如果 taosx 为空，直接 render 到临时目录后同步
# 方式 B：保留 README 历史，迁移模板文件

# 3. 初始化 module
GOWORK=off go mod init github.com/ZoneCNH/taosx || true
GOWORK=off go mod tidy

# 4. 本地验证
GOWORK=off make fmt
GOWORK=off make vet
GOWORK=off make test
GOWORK=off make boundary
GOWORK=off make contracts
GOWORK=off make docs-check

# 5. 生成 Evidence
CHECK_STATUS=passed GOWORK=off make evidence
RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check

# 6. Release final
XLIB_CONTEXT=release_verify GOWORK=off make release-final-check

# 7. 完成声明
# DONE with evidence: ...
```

---

## 33. 验收清单

### Repo Checklist

```text
[ ] go.mod module path correct
[ ] README role correct
[ ] Makefile gates present
[ ] .agent runtime present
[ ] docs standard links present
[ ] release manifest template present
[ ] .gitignore excludes generated latest.json
```

### Runtime Checklist

```text
[ ] Config validate
[ ] Secret redaction
[ ] Client interface
[ ] Fake client
[ ] SQL exec/query
[ ] Batch writer
[ ] Schemaless writer
[ ] Health check
[ ] Error mapping
[ ] Observability injection
[ ] Close idempotency
[ ] Context propagation
```

### Contract Checklist

```text
[ ] config schema
[ ] health schema
[ ] metrics contract
[ ] public API snapshot
[ ] error mapping docs
[ ] schemaless partial failure contract
[ ] SQL render golden
[ ] redaction golden
```

### Harness Checklist

```text
[ ] fmt
[ ] vet
[ ] lint
[ ] test
[ ] race
[ ] boundary
[ ] security
[ ] contracts
[ ] docs-check
[ ] integration
[ ] evidence
[ ] release-evidence-check
[ ] release-final-check
[ ] score >= 9.8
```

### Evidence Checklist

```text
[ ] manifest generated
[ ] manifest not committed
[ ] sha256 generated
[ ] CI artifact uploaded
[ ] source digest exists
[ ] contract fingerprint exists
[ ] dependency list exists
[ ] tool versions exist
[ ] workflow metadata exists
[ ] known gaps explicit
```

### Downstream Checklist

```text
[ ] x.go not imported by taosx
[ ] downstream adoption not overclaimed
[ ] consumer smoke defined
[ ] rollback defined
[ ] adoption proof schema exists
```

---

## 34. 最终推荐路径

最佳路径不是“先手写 TDengine 封装”，而是：

```text
Step 1: 先把 taosx 变成 xlib-standard 标准骨架仓库
Step 2: 建立 L0/L1 contract dependency 和 boundary gate
Step 3: 实现 Config / Client / Health / Error / FakeClient MVA
Step 4: 通过 release Evidence 发布 v0.1.0
Step 5: AutoResearch TDengine Go connector，冻结 WebSocket ADR
Step 6: 实现 SQL / batch / schema / schemaless candidate
Step 7: 加入 optional real TDengine integration gate
Step 8: 发布 v0.2.0 / v0.3.0
Step 9: x.go consumer smoke 形成 proof-based adoption
Step 10: 每次失败和迁移经验反哺 xlib-standard Harness / Rule / Prompt
```

最终结果：

```text
taosx 不再是一个孤立 TDengine 封装库，
而是一个可被标准源生成、被 Harness 裁判、被 Evidence 证明、被下游安全采纳、能持续自我强化的 L2 TDengine adapter 工厂产物。
```

---

## 35. 附录：最小文件交付清单

```text
README.md
go.mod
Makefile
.agent/goal.md
.agent/runtime/goal-runtime.md
.agent/traceability/traceability-matrix.md
.agent/harness/harness.yaml
.agent/evidence/evidence-protocol.md
pkg/taosx/config.go
pkg/taosx/client.go
pkg/taosx/factory.go
pkg/taosx/errors.go
pkg/taosx/health.go
pkg/taosx/sql.go
pkg/taosx/schema.go
pkg/taosx/batch.go
pkg/taosx/schemaless.go
pkg/taosx/metrics.go
internal/driver/fake.go
internal/driver/database_sql.go
internal/dsn/dsn.go
contracts/config.schema.json
contracts/health.schema.json
contracts/metrics.contract.yaml
contracts/public_api.snapshot
contracts/schemaless_failure_contract.md
docs/api.md
docs/config.md
docs/errors.md
docs/health.md
docs/metrics.md
docs/schema.md
docs/schemaless.md
docs/testing.md
docs/release.md
docs/downstream-adoption.md
docs/adr/ADR-20260604-001-tdengine-driver-mode.md
examples/basic-connect/main.go
examples/health-check/main.go
examples/sql-exec/main.go
examples/batch-write/main.go
examples/schemaless-write/main.go
scripts/check_boundary.sh
scripts/check_contracts.sh
scripts/check_docs.sh
scripts/check_secrets.sh
scripts/generate_manifest.sh
release/manifest/template.json
.github/workflows/ci.yml
.github/workflows/release-check.yml
.github/workflows/security.yml
```

---

## 36. 附录：完成声明模板

```text
DONE with evidence:
- scope: goal
- goal_id: GOAL-20260604-TAOSX-L2-FACTORY-001
- repo: github.com/ZoneCNH/taosx
- branch: goal/taosx-l2-factory
- commit: <sha>
- tag: v0.1.0
- gates:
  - GOWORK=off make fmt: passed
  - GOWORK=off make vet: passed
  - GOWORK=off make lint: passed
  - GOWORK=off make test: passed
  - GOWORK=off make race: passed
  - GOWORK=off make boundary: passed
  - GOWORK=off make security: passed
  - GOWORK=off make contracts: passed
  - GOWORK=off make docs-check: passed
  - GOWORK=off make integration: passed|blocked
  - CHECK_STATUS=passed GOWORK=off make evidence: passed
  - RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check: passed
  - XLIB_CONTEXT=release_verify GOWORK=off make release-final-check: passed
- artifacts:
  - release/manifest/latest.json: generated evidence, not committed
  - release/manifest/latest.json.sha256: checksum
  - contracts/public_api.snapshot: API contract
  - contracts/config.schema.json: config contract
  - contracts/health.schema.json: health contract
  - contracts/metrics.contract.yaml: metrics contract
- downstream:
  - x.go consumer smoke: not_run|blocked|passed
  - proof_based_adoption: false|true
  - adoption_claim: not_claimed|adopted
- known gaps:
  - <none or explicit blocker>
```
