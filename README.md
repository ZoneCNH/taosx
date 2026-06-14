# taosx

`taosx` 是 `github.com/ZoneCNH/taosx` 的 L2 TDengine adapter 基础库，当前发布版本为 `v1.0.1`。它提供 TDengine 客户端工厂、配置校验、脱敏、错误分类、SQL 构造、批量写入、schemaless 写入、健康检查和可选可观测性注入，不承载业务时序模型或应用编排。

## 能力边界

- 默认 DriverMode 为 `websocket`，面向 TDengine WebSocket 连接。
- `native_legacy` 仅用于需要兼容本地 native driver 的旧运行时。
- `rest_sql_only` 仅用于 REST SQL fallback，不承诺批量或 schemaless 能力等价。
- 不依赖 `x.go`、业务 repository、业务消息 schema 或应用 profile runtime。
- 不隐式读取 `/home/k8s/secrets/env/*`，不记录原始密码、token 或完整生产 DSN。

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ZoneCNH/taosx/pkg/taosx"
)

func main() {
	cfg := taosx.Config{
		Name:       "metrics",
		Endpoint:   "taos.example.internal:6041",
		Database:   "meters",
		Username:   "root",
		Password:   "secret",
		Timeout:    5 * time.Second,
		MaxRetries: 1,
	}

	client, err := taosx.New(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
	defer client.Close(context.Background())

	fmt.Println(cfg.RedactedDSN())
}
```

默认构造不会创建真实 TDengine 网络连接；未注入 `Driver` 时，执行类方法返回 `unavailable` retryable 错误，健康状态为 `degraded`。生产集成应通过 `taosx.WithDriver(...)` 注入具体 driver 适配层。

## 公共 API

- `Config.Normalize()`：补齐默认 name、driver_mode 和 timeout。
- `Config.Validate()`：校验 driver_mode、endpoint、database、timeout、retry 等字段。
- `Config.Sanitized()` / `Config.Sanitize()`：返回可写入日志、Evidence 和错误上下文的脱敏副本。
- `Config.RedactedDSN()`：生成脱敏 TDengine DSN，密码固定替换为 `***`。
- `New(ctx, cfg, opts...)`：创建上下文感知 client，支持 `WithDriver`、`WithMetrics`、`WithClock`。
- `Client`：提供 `Exec`、`Query`、`WriteBatch`、`SchemalessWrite`、`Health`、`Close`。
- `Statement`、`Query`、`Batch`、`SchemalessPayload`：表达 SQL、查询、批量写入和 schemaless 写入参数。
- `ColumnSpec`、`StableSpec`、`RenderCreateStable`：提供 identifier allowlist 和稳定 STABLE SQL 生成。
- `Error`、`ErrorKind`、`IsKind`、`IsRetryable`：提供稳定错误 contract。

## 文档入口

- [配置](docs/config.md)：DriverMode、字段、校验和脱敏规则。
- [API](docs/api.md)：Client、driver 注入、SQL/query/write contract。
- [错误](docs/errors.md)：错误分类、retry 语义和 secret redaction。
- [可观测性](docs/observability.md)：可选 metrics 注入与健康状态。
- [快速开始](docs/quickstart.md)：本地验证和示例命令。
- [测试策略](docs/testing.md)：taosx 单测、schema、示例和 gate 覆盖。
- [下游矩阵](docs/downstream-matrix.md)：taosx 在 L2 adapter 矩阵中的状态和采用证据边界。

## 本地验证

```bash
GOWORK=off make docs-check
GOWORK=off make dependency-check
GOWORK=off make standard-impact-check
GOWORK=off make release-check
go test ./pkg/taosx ./contracts ./examples/...
go test ./pkg/taosx -coverprofile=/tmp/taosx.cover
go tool cover -func=/tmp/taosx.cover
go test ./...
TAOSX_INTEGRATION=1 go test -tags=integration ./pkg/taosx -run TestTDengineWebSocketIntegration -count=1
./scripts/check_boundary.sh
./scripts/check_contracts.sh
GOWORK=off go run ./cmd/goalcli score --min 9.8
```

v1.0.1 发布验证要求 `pkg/taosx` 覆盖率达到 100.0%，并使用本地受控 dev 环境通过官方 `taosWS` WebSocket 真实集成测试。真实 TDengine 集成测试必须从本地受控环境注入 `TAOSX_TDENGINE_ENDPOINT`、`TAOSX_TDENGINE_USER`、`TAOSX_TDENGINE_PASSWORD` 和 `TAOSX_TDENGINE_DATABASE`，或注入完整 `TAOSX_TDENGINE_DSN`。测试和文档不得输出原始密码或完整 DSN。

发布、Evidence 和 release 语境下的命令必须显式使用 `GOWORK=off`，避免本地 `go.work` 改写 module 解析。

标准运行时仍保留依赖、下游同步和 Docker Toolchain Runtime gate：

- 依赖漂移由 `renovate.json`、`.github/dependabot.yml`、`GOWORK=off make dependency-check` 和 `GOWORK=off make standard-impact-check` 共同约束。
- 下游同步以 `docs/downstream-sync-policy.md`、`docs/downstream-matrix.md` 和 `downstream_sync_required` 为准；`kernel` 是必须跟踪的下游采用目标之一。
- 深度 fuzz 使用 `FUZZ_SMOKE_TIME` 调整窗口。
- Docker Toolchain Runtime 规范见 `docs/standard/docker-toolchain-standard.md`，本地/CI gate 包括 `make docker-toolchain-check`、`make docker-ci` 和 `make docker-release-check`。

## Evidence

`release/manifest/latest.json`、`release/manifest/latest.json.sha256` 与 `release/standard-impact/latest.md` 是生成产物，不提交源码历史。完成声明必须基于真实 gate 输出，并包含 `DONE with evidence:`。

`standard_impact.downstream_release_decision`（只允许 `required` / `not_required`）和 `standard_impact.repository_rules_release_decision`（只允许 `audit_required` / `not_required`）必须与 release Evidence 保持一致。
