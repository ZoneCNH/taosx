# baselib-template 完整可执行 Goal Prompt v1.0

> 文件用途：把本文件完整交给 Agent Teams / Codex / 自动化执行器，用于创建 `baselib-template` 独立基础库模板仓库。
> 适用项目：x.go 基础库体系。
> 执行标准：Goal Runtime Prompt v3.1。
> 当前日期：2026-06-01。
> 默认执行模式：Full。
> 完成声明必须使用：`DONE with evidence:`。

---

# 0. 执行总提示

你是负责 x.go 基础设施资产化的工程 Agent。你的任务不是写一份说明文档，而是实际创建一个可复用、可测试、可发布、可治理的独立基础库模板仓库：`baselib-template`。

该模板将作为后续所有基础库的统一生成源，包括但不限于：

- `foundationx`
- `configx`
- `observex`
- `postgresx`
- `kafkax`
- `redisx`
- `taosx`
- `ossx`
- `testkitx`

你必须按照 Goal Runtime Prompt v3.1 执行，完整走完：

```text
目标
 → 上下文恢复
 → 规格
 → Design
 → 计划
 → 任务
 → 执行
 → 验证
 → 证据
 → 评审
 → 发布
 → 复盘
 → 自我改进
```

你不得只输出计划。你必须创建文件、脚本、目录、CI、文档模板、Harness 门禁、证据模板和发布清单模板。

---

# 1. Goal

```text
GOAL-20260601-001
建立 baselib-template 独立基础库模板仓库，为 x.go 及其周边基础设施库提供统一模块脚手架、目录规范、公共文档、CI/Harness 门禁、发布证据、复盘和自我改进机制，使后续 foundationx/postgresx/kafkax/redisx/taosx 等基础库可以低成本、标准化、可验证地生成。
```

---

# 2. 当前对象

当前要创建的是：

```text
仓库：baselib-template
模块类型：独立 Go 基础库模板
目标使用方：
  - foundationx
  - configx
  - observex
  - postgresx
  - kafkax
  - redisx
  - taosx
  - ossx
  - testkitx
```

模板仓库本身不应该绑定某个具体基础设施实现。它应该提供通用基础库的最小标准结构。

---

# 3. 执行模式

采用完整模式（Full）。

完整模式要求：

1. 必须创建完整目录结构。
2. 必须创建可执行 Makefile。
3. 必须创建 CI 工作流。
4. 必须创建 Harness 门禁脚本。
5. 必须创建文档模板。
6. 必须创建 `.agent/` Goal Runtime 文件。
7. 必须创建 contracts 模板。
8. 必须创建 release manifest 模板。
9. 必须创建 examples 模板。
10. 必须运行验证命令。
11. 必须输出证据。
12. 必须输出复盘。
13. 不允许没有证据就声称完成。

---

# 4. 固定事实与硬约束

## 4.1 x.go 基础库体系约束

所有从 `baselib-template` 生成的基础库都必须满足：

```text
1. 独立 Go module。
2. 不依赖 github.com/bytechainx/x.go。
3. 不依赖 x.go/internal/*。
4. 不包含 BTCUSDT、Kline、MacroRegime、TradingSignal、OrderBook、Position、RiskGate 等业务语义。
5. 不持有隐式全局 client。
6. 不默认读取生产密钥。
7. 不把密钥写入源码、README、测试日志、Release Manifest、PR 描述或 Issue。
8. 公共 API 必须稳定、可测试、可文档化。
9. Config 必须 Validate + Sanitize。
10. 有资源的 client 必须支持 Close，并且 Close 幂等。
11. L2 基础设施库必须支持 HealthCheck。
12. 必须支持 Harness 门禁。
13. 必须支持证据。
```

## 4.2 x.go 密钥路径约束

x.go 的 Redis / Kafka / PostgreSQL / TDengine / OSS 等配置和密钥位于：

```text
/home/k8s/secrets/env/*
```

模板只能提供显式加载接口或文档说明，不得默认读取该路径。

正确模式：

```text
调用方显式传入路径 → configx 加载 → 基础库接收 Config
```

错误模式：

```text
基础库自动读取 /home/k8s/secrets/env/*
```

## 4.3 禁止行为

严禁：

```text
- 在模板中写入任何真实密钥。
- 在模板中硬编码生产连接地址。
- 在模板中 import github.com/bytechainx/x.go。
- 在模板中定义 x.go 业务模型。
- 使用 todo!/panic/未实现占位作为完成状态。
- 只创建 README，不创建可执行脚本。
- 只说“已完成”，不提供证据。
```

---

# 5. 上下文恢复

执行前先恢复上下文并确认以下事实：

```text
1. 当前仓库是否已经存在 baselib-template？
2. 如果存在，是否已有 go.mod、Makefile、.github/workflows？
3. 如果不存在，创建新目录 baselib-template。
4. 当前 Go 版本是什么？
5. 当前是否有 git 仓库？
6. 当前是否能运行 go test ./...？
7. 当前是否能运行 bash 脚本？
```

如果某个上下文缺失，不要停止执行。使用合理默认值继续：

```text
module path 默认：github.com/bytechainx/baselib-template
Go 版本默认：1.23
许可证默认：MIT 或当前组织默认许可证
模板占位符：{{MODULE_NAME}}, {{MODULE_PATH}}, {{PACKAGE_NAME}}
```

---

# 6. 规格

```text
SPEC-baselib-template-v1.0
```

## 6.1 需求列表

### REQ-BT-001：独立模板仓库

`baselib-template` 必须是一个独立 Go module。

验收标准：

```text
AC-REQ-BT-001-001: 存在 go.mod。
AC-REQ-BT-001-002: module path 不包含 x.go。
AC-REQ-BT-001-003: go test ./... 可运行。
```

### REQ-BT-002：标准目录结构

必须包含：

```text
pkg/{{module}}
internal/
testkit/
examples/
contracts/
docs/
scripts/
release/manifest/
.agent/
.github/workflows/
```

验收标准：

```text
AC-REQ-BT-002-001: 所有目录存在。
AC-REQ-BT-002-002: 每个目录至少有 README 或模板文件，避免空目录丢失。
```

### REQ-BT-003：公共 API 模板

必须提供最小基础库 API 模板：

```text
Config
Validate
Sanitize
Client
New
Option
HealthCheck
错误模型
指标钩子
Version
```

验收标准：

```text
AC-REQ-BT-003-001: pkg/{{module}} 下存在 config.go/client.go/options.go/health.go/errors.go/metrics.go/version.go/doc.go。
AC-REQ-BT-003-002: go test ./... 通过。
```

### REQ-BT-004：Harness 门禁

必须提供：

```text
边界门禁
密钥门禁
契约门禁
格式门禁
静态检查门禁
单元测试门禁
竞态测试门禁
安全门禁
证据门禁
发布门禁
```

验收标准：

```text
AC-REQ-BT-004-001: scripts/check_boundary.sh 存在且可执行。
AC-REQ-BT-004-002: scripts/check_secrets.sh 存在且可执行。
AC-REQ-BT-004-003: scripts/check_contracts.sh 存在且可执行。
AC-REQ-BT-004-004: Makefile 中存在对应命令。
```

### REQ-BT-005：CI 工作流

必须提供 GitHub Actions：

```text
ci.yml
integration.yml
security.yml
release.yml
```

验收标准：

```text
AC-REQ-BT-005-001: .github/workflows/ci.yml 存在。
AC-REQ-BT-005-002: CI 至少执行 fmt/vet/test/race/boundary/security/contracts。
```

### REQ-BT-006：文档模板

必须提供：

```text
README.md
CHANGELOG.md
docs/spec.md
docs/design.md
docs/api.md
docs/config.md
docs/errors.md
docs/observability.md
docs/testing.md
docs/release.md
docs/adr/ADR-000-template.md
```

验收标准：

```text
AC-REQ-BT-006-001: 所有文档存在。
AC-REQ-BT-006-002: 文档包含占位符说明。
AC-REQ-BT-006-003: 文档明确基础库不得依赖 x.go。
```

### REQ-BT-007：发布证据

必须提供：

```text
release/manifest/template.json
scripts/generate_manifest.sh
.agent/evidence.md
```

验收标准：

```text
AC-REQ-BT-007-001: make evidence 能生成 release/manifest/latest.json。
AC-REQ-BT-007-002: Manifest 包含 module/version/commit/checks/artifacts。
```

### REQ-BT-008：自我改进

必须提供复盘模板：

```text
.agent/retrospective.md
```

验收标准：

```text
AC-REQ-BT-008-001: 复盘包含提示补丁、Harness 补丁、规则补丁、CI 门禁建议和新 Issue 候选。
```

---

# 7. Design

```text
DESIGN-baselib-template-v1.0
```

## 7.1 总体目录设计

最终必须创建：

```text
baselib-template/
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── LICENSE
├── Makefile
├── .gitignore
├── .golangci.yml
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── integration.yml
│       ├── security.yml
│       └── release.yml
├── pkg/
│   └── templatex/
│       ├── config.go
│       ├── client.go
│       ├── options.go
│       ├── health.go
│       ├── errors.go
│       ├── metrics.go
│       ├── version.go
│       ├── doc.go
│       ├── config_test.go
│       ├── client_test.go
│       └── health_test.go
├── internal/
│   ├── sanitize/
│   │   ├── sanitize.go
│   │   └── sanitize_test.go
│   ├── validation/
│   │   ├── validation.go
│   │   └── validation_test.go
│   └── runtime/
│       └── README.md
├── testkit/
│   ├── fixture.go
│   ├── assert.go
│   └── README.md
├── examples/
│   ├── basic/
│   │   └── main.go
│   ├── health/
│   │   └── main.go
│   └── config/
│       └── main.go
├── contracts/
│   ├── config.schema.json
│   ├── health.schema.json
│   ├── error.schema.json
│   └── metrics.md
├── docs/
│   ├── spec.md
│   ├── design.md
│   ├── api.md
│   ├── config.md
│   ├── errors.md
│   ├── observability.md
│   ├── testing.md
│   ├── release.md
│   └── adr/
│       └── ADR-000-template.md
├── scripts/
│   ├── check_boundary.sh
│   ├── check_secrets.sh
│   ├── check_contracts.sh
│   ├── generate_manifest.sh
│   └── run_integration.sh
├── release/
│   └── manifest/
│       └── template.json
└── .agent/
    ├── goal.md
    ├── spec.md
    ├── design.md
    ├── plan.md
    ├── tasks.md
    ├── harness.md
    ├── gates.md
    ├── evidence.md
    ├── review.md
    ├── release.md
    └── retrospective.md
```

说明：

```text
templatex 是模板自身可编译示例包。
后续生成 foundationx/postgresx 等库时，可以把 templatex 替换为实际包名。
```

---

# 8. 实现要求

## 8.1 go.mod

必须创建：

```go
module github.com/bytechainx/baselib-template

go 1.23
```

## 8.2 pkg/templatex/config.go

必须包含：

```go
package templatex

import (
	"errors"
	"time"
)

type Config struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

type SanitizedConfig struct {
	Name    string
	Timeout time.Duration
	Secret  string
}

func (c Config) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.Timeout < 0 {
		return errors.New("timeout must not be negative")
	}
	return nil
}

func (c Config) Sanitize() SanitizedConfig {
	secret := ""
	if c.Secret != "" {
		secret = "***"
	}
	return SanitizedConfig{
		Name:    c.Name,
		Timeout: c.Timeout,
		Secret:  secret,
	}
}
```

## 8.3 pkg/templatex/client.go

必须包含：

```go
package templatex

import (
	"context"
	"sync"
)

type Client struct {
	cfg    Config
	mu     sync.Mutex
	closed bool
}

func New(ctx context.Context, cfg Config, opts ...Option) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Client{cfg: cfg}, nil
}

func (c *Client) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true
	return nil
}
```

## 8.4 pkg/templatex/options.go

必须包含：

```go
package templatex

type Option func(*options)

type options struct {
	metrics Metrics
}

func defaultOptions() options {
	return options{
		metrics: NoopMetrics{},
	}
}

func WithMetrics(metrics Metrics) Option {
	return func(o *options) {
		if metrics != nil {
			o.metrics = metrics
		}
	}
}
```

## 8.5 pkg/templatex/health.go

必须包含：

```go
package templatex

import (
	"context"
	"time"
)

type HealthStatusValue string

const (
	HealthHealthy   HealthStatusValue = "healthy"
	HealthDegraded  HealthStatusValue = "degraded"
	HealthUnhealthy HealthStatusValue = "unhealthy"
)

type HealthStatus struct {
	Name      string
	Status    HealthStatusValue
	Message   string
	CheckedAt time.Time
	LatencyMs int64
	Metadata  map[string]string
}

func (c *Client) HealthCheck(ctx context.Context) HealthStatus {
	start := time.Now()

	if err := ctx.Err(); err != nil {
		return HealthStatus{
			Name:      "templatex",
			Status:    HealthUnhealthy,
			Message:   err.Error(),
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()

	if closed {
		return HealthStatus{
			Name:      "templatex",
			Status:    HealthUnhealthy,
			Message:   "client is closed",
			CheckedAt: time.Now(),
			LatencyMs: time.Since(start).Milliseconds(),
		}
	}

	return HealthStatus{
		Name:      "templatex",
		Status:    HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Now(),
		LatencyMs: time.Since(start).Milliseconds(),
	}
}
```

## 8.6 pkg/templatex/errors.go

必须包含：

```go
package templatex

import "errors"

type ErrorKind string

const (
	ErrorKindConfig      ErrorKind = "config"
	ErrorKindValidation  ErrorKind = "validation"
	ErrorKindConnection  ErrorKind = "connection"
	ErrorKindUnavailable ErrorKind = "unavailable"
	ErrorKindTimeout     ErrorKind = "timeout"
	ErrorKindAuth        ErrorKind = "auth"
	ErrorKindConflict    ErrorKind = "conflict"
	ErrorKindRateLimit   ErrorKind = "rate_limit"
	ErrorKindInternal    ErrorKind = "internal"
)

type Error struct {
	Kind      ErrorKind
	Op        string
	Message   string
	Cause     error
	Retryable bool
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return string(e.Kind) + ": " + e.Op + ": " + e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func IsKind(err error, kind ErrorKind) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Kind == kind
	}
	return false
}
```

## 8.7 pkg/templatex/metrics.go

必须包含：

```go
package templatex

type Metrics interface {
	IncCounter(name string, labels map[string]string)
	ObserveHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
}

type NoopMetrics struct{}

func (NoopMetrics) IncCounter(name string, labels map[string]string) {}

func (NoopMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {}

func (NoopMetrics) SetGauge(name string, value float64, labels map[string]string) {}
```

## 8.8 pkg/templatex/version.go

必须包含：

```go
package templatex

const (
	ModuleName = "github.com/bytechainx/baselib-template"
	Version    = "v0.1.0"
)
```

## 8.9 pkg/templatex/doc.go

必须包含：

```go
// Package templatex provides a minimal base-library template package.
//
// This package demonstrates the required structure for independent base libraries:
// Config, Validate, Sanitize, New, Close, HealthCheck, Error model, Metrics hooks,
// tests, examples, contracts, CI gates, release manifest, and agent evidence.
//
// This package must not depend on github.com/bytechainx/x.go or any x.go internal package.
package templatex
```

---

# 9. Scripts

## 9.1 scripts/check_boundary.sh

必须创建并 chmod +x：

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking forbidden dependency on x.go..."

if go list -deps ./... | grep -q "github.com/bytechainx/x.go"; then
  echo "ERROR: base library template must not depend on x.go"
  exit 1
fi

echo "checking forbidden business terms..."

FORBIDDEN_TERMS=(
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "BTCUSDT"
  "ETHUSDT"
  "Kline"
  "OrderBook"
  "Position"
  "RiskGate"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R "$term" ./pkg ./internal --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "boundary check passed"
```

注意：如果文档中需要出现这些词，Boundary Gate 只扫描 `pkg` 和 `internal`，不扫描 `docs`。

## 9.2 scripts/check_secrets.sh

必须创建并 chmod +x：

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking secrets..."

PATTERNS=(
  "password="
  "passwd="
  "secret="
  "token="
  "access_key="
  "secret_key="
  "AKIA[0-9A-Z]{16}"
  "BEGIN RSA PRIVATE KEY"
  "BEGIN OPENSSH PRIVATE KEY"
)

for pattern in "${PATTERNS[@]}"; do
  if grep -R -E "$pattern" . \
    --exclude-dir=.git \
    --exclude-dir=vendor \
    --exclude="*.sum" \
    --exclude="check_secrets.sh"; then
    echo "ERROR: possible secret found: $pattern"
    exit 1
  fi
done

echo "secret check passed"
```

## 9.3 scripts/check_contracts.sh

必须创建并 chmod +x：

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "checking contracts..."

REQUIRED_FILES=(
  "contracts/config.schema.json"
  "contracts/health.schema.json"
  "contracts/error.schema.json"
  "contracts/metrics.md"
)

for file in "${REQUIRED_FILES[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: missing contract file: $file"
    exit 1
  fi
done

echo "contract check passed"
```

## 9.4 scripts/generate_manifest.sh

必须创建并 chmod +x：

```bash
#!/usr/bin/env bash
set -euo pipefail

mkdir -p release/manifest

MODULE="$(go list -m)"
VERSION="${VERSION:-v0.1.0}"
COMMIT="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
GO_VERSION="$(go version | awk '{print $3}')"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

cat > release/manifest/latest.json <<JSON
{
  "module": "${MODULE}",
  "version": "${VERSION}",
  "commit": "${COMMIT}",
  "go_version": "${GO_VERSION}",
  "generated_at": "${GENERATED_AT}",
  "checks": {
    "fmt": "manual-or-ci",
    "vet": "manual-or-ci",
    "unit_test": "manual-or-ci",
    "race_test": "manual-or-ci",
    "boundary": "manual-or-ci",
    "secret_scan": "manual-or-ci",
    "contract": "manual-or-ci"
  },
  "artifacts": [
    "release/manifest/latest.json"
  ],
  "notes": {
    "breaking_changes": "none",
    "known_risks": []
  }
}
JSON

echo "generated release/manifest/latest.json"
```

## 9.5 scripts/run_integration.sh

必须创建并 chmod +x：

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "no external integration dependency for baselib-template"
echo "integration check passed"
```

---

# 10. Makefile

必须创建：

```makefile
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

.PHONY: integration
integration:
	./scripts/run_integration.sh

.PHONY: security
security:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed, skipping"; \
	fi
	./scripts/check_secrets.sh

.PHONY: boundary
boundary:
	./scripts/check_boundary.sh

.PHONY: contracts
contracts:
	./scripts/check_contracts.sh

.PHONY: evidence
evidence:
	./scripts/generate_manifest.sh

.PHONY: ci
ci: fmt vet lint test race boundary security contracts

.PHONY: release-check
release-check: ci integration evidence
```

---

# 11. GitHub Actions

## 11.1 .github/workflows/ci.yml

```yaml
name: CI

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  ci:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Cache Go
        uses: actions/cache@v4
        with:
          path: |
            ~/.cache/go-build
            ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: CI
        run: make ci
```

## 11.2 .github/workflows/integration.yml

```yaml
name: Integration

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  integration:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Integration
        run: make integration
```

## 11.3 .github/workflows/security.yml

```yaml
name: Security

on:
  pull_request:
  push:
    branches:
      - main

jobs:
  security:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Security
        run: make security
```

## 11.4 .github/workflows/release.yml

```yaml
name: Release Check

on:
  push:
    tags:
      - "v*"

jobs:
  release-check:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Release Check
        run: make release-check

      - name: Upload Manifest
        uses: actions/upload-artifact@v4
        with:
          name: release-manifest
          path: release/manifest/latest.json
```

---

# 12. 文档模板

## 12.1 README.md 必须包含

```markdown
# {{MODULE_NAME}}

{{MODULE_NAME}} 是一个独立 Go 基础库模板模块。

## 目标

{{MODULE_NAME}} 是一个独立 Go 基础库模板模块，为创建独立基础库提供标准骨架。

## 非目标

- 不依赖 x.go。
- 不包含 x.go 业务模型。
- 不隐式读取生产密钥。

## 标准结构

...

## 命令

make ci
make release-check

## 证据

完成声明必须包含发布清单和 CI 证据。
```

## 12.2 docs/spec.md 必须包含

```markdown
# SPEC-{{MODULE_NAME}}-v1.0

## 需求

## 验收标准

## 非目标

## 可追踪性
```

## 12.3 docs/design.md 必须包含

```markdown
# DESIGN-{{MODULE_NAME}}-v1.0

## 架构

## 公共 API

## 配置

## 错误模型

## 健康检查

## 指标

## 测试

## 发布
```

## 12.4 docs/adr/ADR-000-template.md 必须包含

```markdown
# ADR-000: 模板决策记录

## 状态

待定

## 背景

## 决策

## 后果

## 证据
```

---

# 13. .agent Goal Runtime 文件

必须创建：

```text
.agent/goal.md
.agent/spec.md
.agent/design.md
.agent/plan.md
.agent/tasks.md
.agent/harness.md
.agent/gates.md
.agent/evidence.md
.agent/review.md
.agent/release.md
.agent/retrospective.md
```

## 13.1 .agent/goal.md

包含：

```markdown
# GOAL-20260601-001

将 baselib-template 构建为 x.go 基础库的标准独立基础库模板。
```

## 13.2 .agent/harness.md

包含：

```markdown
# Harness 协议

## 门禁

- 上下文门禁
- 目标门禁
- 规格门禁
- 设计门禁
- 计划门禁
- 任务门禁
- 实现门禁
- 测试门禁
- 证据门禁
- 评审门禁
- 发布门禁
- 复盘门禁
```

## 13.3 .agent/evidence.md

包含：

```markdown
# 证据

完成声明必须包含：

- go test ./...
- go test -race ./...
- make boundary
- make security
- make contracts
- make evidence
- release/manifest/latest.json

最终声明必须使用：

DONE with evidence:
```

## 13.4 .agent/retrospective.md

包含：

```markdown
# 复盘

## 改进项

## 失败项

## 提示补丁

## Harness 补丁

## 规则补丁

## CI 门禁建议

## 新 Issue 候选
```

---

# 14. 契约

## 14.1 contracts/config.schema.json

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Base Library Config",
  "type": "object",
  "required": ["name"],
  "properties": {
    "name": {
      "type": "string"
    },
    "timeout_ms": {
      "type": "integer",
      "minimum": 0
    }
  },
  "additionalProperties": true
}
```

## 14.2 contracts/health.schema.json

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Health Status",
  "type": "object",
  "required": ["name", "status", "checked_at"],
  "properties": {
    "name": {
      "type": "string"
    },
    "status": {
      "enum": ["healthy", "degraded", "unhealthy"]
    },
    "message": {
      "type": "string"
    },
    "checked_at": {
      "type": "string"
    },
    "latency_ms": {
      "type": "integer"
    },
    "metadata": {
      "type": "object",
      "additionalProperties": {
        "type": "string"
      }
    }
  }
}
```

## 14.3 contracts/error.schema.json

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Base Library Error",
  "type": "object",
  "required": ["kind", "op", "message", "retryable"],
  "properties": {
    "kind": {
      "enum": [
        "config",
        "validation",
        "connection",
        "unavailable",
        "timeout",
        "auth",
        "conflict",
        "rate_limit",
        "internal"
      ]
    },
    "op": {
      "type": "string"
    },
    "message": {
      "type": "string"
    },
    "retryable": {
      "type": "boolean"
    }
  }
}
```

## 14.4 contracts/metrics.md

必须包含：

```markdown
# 指标契约

标准指标：

- client_requests_total
- client_request_duration_seconds
- client_errors_total
- client_retries_total
- client_inflight
- client_health_status
```

---

# 15. 测试

必须创建并通过：

```text
pkg/templatex/config_test.go
pkg/templatex/client_test.go
pkg/templatex/health_test.go
internal/sanitize/sanitize_test.go
internal/validation/validation_test.go
```

测试要求：

```text
1. Config.Validate 空 Name 返回错误。
2. Config.Sanitize 隐藏 Secret。
3. New(ctx,cfg) 对无效 cfg 返回错误。
4. Close 可以重复调用。
5. HealthCheck 对正常 client 返回 healthy。
6. HealthCheck 对 closed client 返回 unhealthy。
7. go test -race ./... 通过。
```

---

# 16. 计划

```text
PLAN-GOAL-20260601-001-v1.0
```

## 里程碑 1：创建模板骨架

```text
TASK-BT-001 创建目录结构
TASK-BT-002 创建 go.mod
TASK-BT-003 创建 README/CHANGELOG/LICENSE
```

## 里程碑 2：创建可编译示例包

```text
TASK-BT-004 实现 pkg/templatex/config.go
TASK-BT-005 实现 pkg/templatex/client.go
TASK-BT-006 实现 pkg/templatex/options.go
TASK-BT-007 实现 pkg/templatex/health.go
TASK-BT-008 实现 pkg/templatex/errors.go
TASK-BT-009 实现 pkg/templatex/metrics.go
TASK-BT-010 实现 pkg/templatex/version.go/doc.go
```

## 里程碑 3：创建测试

```text
TASK-BT-011 创建 config_test.go
TASK-BT-012 创建 client_test.go
TASK-BT-013 创建 health_test.go
TASK-BT-014 创建 internal 测试
```

## 里程碑 4：创建 Harness

```text
TASK-BT-015 创建 Makefile
TASK-BT-016 创建 check_boundary.sh
TASK-BT-017 创建 check_secrets.sh
TASK-BT-018 创建 check_contracts.sh
TASK-BT-019 创建 generate_manifest.sh
TASK-BT-020 创建 run_integration.sh
```

## 里程碑 5：创建 CI

```text
TASK-BT-021 创建 ci.yml
TASK-BT-022 创建 integration.yml
TASK-BT-023 创建 security.yml
TASK-BT-024 创建 release.yml
```

## 里程碑 6：创建文档和 .agent

```text
TASK-BT-025 创建 docs/*
TASK-BT-026 创建 contracts/*
TASK-BT-027 创建 examples/*
TASK-BT-028 创建 .agent/*
TASK-BT-029 创建 release/manifest/template.json
```

## 里程碑 7：验证和证据

```text
TASK-BT-030 运行 go test ./...
TASK-BT-031 运行 go test -race ./...
TASK-BT-032 运行 make boundary
TASK-BT-033 运行 make security
TASK-BT-034 运行 make contracts
TASK-BT-035 运行 make evidence
TASK-BT-036 输出 DONE with evidence
```

---

# 17. 验证命令

必须尽量执行：

```bash
go test ./...
go test -race ./...
make boundary
make security
make contracts
make evidence
make release-check
```

如果本地缺少 `golangci-lint` 或 `govulncheck`，可以跳过，但必须在证据中说明：

```text
golangci-lint 未安装，由 Makefile fallback 跳过
govulncheck 未安装，由 Makefile fallback 跳过
```

---

# 18. 证据协议

最终必须输出：

```text
DONE with evidence:
- 已创建 baselib-template 目录。
- 已创建 go.mod。
- 已创建标准目录结构。
- 已创建 pkg/templatex 最小可编译包。
- 已创建测试。
- go test ./... 已通过。
- go test -race ./... 已通过。
- make boundary 已通过。
- make security 已通过。
- make contracts 已通过。
- make evidence 已生成 release/manifest/latest.json。
- 未依赖 github.com/bytechainx/x.go。
- pkg/internal 中没有业务语义。
- 未检测到密钥。
```

如果某个检查失败，必须输出：

```text
NOT DONE:
- 失败门禁：
- 原因：
- 已应用修复：
- 剩余风险：
```

不得声称完成。

---

# 19. 评审门禁

Review 必须检查：

```text
1. 是否是独立 Go module？
2. 是否不依赖 x.go？
3. 是否有标准目录？
4. 是否有可编译 pkg/templatex？
5. 是否有测试？
6. 是否有 Makefile？
7. 是否有 scripts？
8. 是否有 CI？
9. 是否有 contracts？
10. 是否有 docs？
11. 是否有 .agent？
12. 是否能生成 release manifest？
13. 是否能作为 foundationx/postgresx/kafkax/redisx/taosx 的模板？
```

---

# 20. 发布门禁

Release 前必须满足：

```text
go test ./... 已通过
go test -race ./... 已通过
make boundary 已通过
make security 已通过
make contracts 已通过
make evidence 已通过
CHANGELOG.md 已更新
release/manifest/latest.json 存在
```

发布版本：

```text
v0.1.0
```

`CHANGELOG.md` 必须包含：

```markdown
## v0.1.0 - 2026-06-01

### 新增
- 初始 baselib-template 结构。
- 标准 Go 基础库包骨架。
- Makefile 命令。
- Harness 门禁脚本。
- GitHub Actions 工作流。
- 契约。
- Agent 运行时模板。
- 发布清单模板。

### 安全
- 新增密钥扫描门禁。

### 治理
- 新增证据和复盘模板。
```

---

# 21. 复盘

完成后必须创建或更新：

```text
.agent/retrospective.md
```

至少包含：

```text
## 改进项
- 基础库创建从手工变成模板化。
- 后续 foundationx/postgresx/kafkax/redisx 可复用目录、脚本、CI、文档和 Evidence。

## 失败项
- 记录执行中失败或跳过的 gate。

## 提示补丁
- 后续创建基础库时必须从 baselib-template 复制。
- 所有基础库必须保留 Boundary Gate 和 Secret Gate。

## Harness 补丁
- 后续加入 public API hash gate。
- 后续加入 config schema hash gate。

## 规则补丁
- 禁止基础库依赖 x.go。
- 禁止基础库承载业务语义。
- 禁止无 Evidence 声称 DONE。

## CI 门禁建议
- 加入 CodeQL。
- 加入 govulncheck 强制模式。
- 加入覆盖率阈值。

## 新 Issue 候选
- ISSUE-FOUNDATIONX-001 从 baselib-template 生成 foundationx。
- ISSUE-POSTGRESX-001 从 baselib-template 生成 postgresx。
```

---

# 22. Agent Teams 并行执行建议

如果使用多个 Agent，可按以下分工：

## Agent A：骨架 Agent

负责：

```text
目录结构
go.mod
README
CHANGELOG
LICENSE
.gitignore
```

## Agent B：代码 Agent

负责：

```text
pkg/templatex/*
internal/*
tests
examples
```

## Agent C：Harness Agent

负责：

```text
Makefile
scripts/*
.github/workflows/*
contracts/*
```

## Agent D：治理 Agent

负责：

```text
docs/*
.agent/*
release/manifest/*
retrospective
```

合并顺序：

```text
A → B → C → D → 验证 → 评审 → 发布
```

---

# 23. 文件生成注意事项

必须避免空目录丢失。空目录用 `README.md` 或 `.gitkeep` 保留。

脚本必须可执行：

```bash
chmod +x scripts/*.sh
```

所有 shell 脚本必须：

```bash
set -euo pipefail
```

所有 Go 文件必须：

```bash
go fmt ./...
```

所有测试必须：

```bash
go test ./...
```

---

# 24. 最终输出格式

执行结束后，必须输出以下格式：

```text
DONE with evidence:

目标：
- GOAL-20260601-001

已创建：
- baselib-template/go.mod
- baselib-template/pkg/templatex/*
- baselib-template/scripts/*
- baselib-template/.github/workflows/*
- baselib-template/docs/*
- baselib-template/contracts/*
- baselib-template/.agent/*
- baselib-template/release/manifest/latest.json

验证：
- go test ./...: passed
- go test -race ./...: passed
- make boundary: passed
- make security: passed
- make contracts: passed
- make evidence: passed

证据：
- release/manifest/latest.json

已知风险：
- ...

下一步：
- 从 baselib-template 生成 foundationx。
- 从 baselib-template 生成 postgresx。
```

如果未完成：

```text
NOT DONE:

阻塞或失败门禁：
- ...

部分证据：
- ...

必需的下一项修复：
- ...
```

---

# 25. 开始执行

现在开始执行，不要停留在解释层。请实际创建 `baselib-template` 的目录、文件、脚本、CI、文档、contracts、.agent、release manifest，并运行验证命令。

优先原则：

```text
1. 可执行 > 完美
2. 有 Evidence > 口头声明
3. 独立边界 > 方便复用
4. 模板稳定 > 业务定制
5. 后续可复制 > 一次性脚手架
```

完成后只允许用 `DONE with evidence:` 声明完成。
