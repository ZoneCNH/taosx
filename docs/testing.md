# taosx 测试策略

`taosx` 的测试目标是锁定 TDengine adapter contract、标准运行时 gate 和 Evidence 生成边界。

## 必需覆盖范围

- `go test ./pkg/taosx` 覆盖配置、脱敏、driver mode、client lifecycle、context、错误、SQL helper、batch、schemaless 和 health。
- `go test ./contracts` 覆盖 JSON schema 与公共常量映射。
- `go test ./examples/...` 覆盖 README 和 docs 中的最小示例。
- `go test ./scripts` 覆盖 render_template 运行态目录排除和下游控制面文件继承。
- `go test ./cmd/goalcli` 覆盖治理 gate、迁移 guard、debt evidence 和 release manifest fixture。
- Release manifest 测试必须在临时 fixture 仓库构造所需 `.omc` state，不得依赖当前工作区的 Agent 运行态文件。

## taosx contract

- `Config.Validate` 必须拒绝空 endpoint、空 database、非法 DriverMode、负 timeout 和负 retry。
- `Config.Sanitized` 与 `RedactedDSN` 必须屏蔽密码。
- `Close` 必须幂等。
- 未注入 driver 时执行方法必须返回 retryable `ErrorKindUnavailable`。
- context deadline exceeded 必须映射为 `ErrorKindTimeout`。
- schema helper 必须拒绝非法 identifier，并生成稳定 SQL。
- `Health` 在未注入真实 driver 时返回 `degraded`。

## Gate 命令

```bash
go test ./...
./scripts/check_boundary.sh
./scripts/check_contracts.sh
GOWORK=off make docker-toolchain-check
GOWORK=off make integration DOWNSTREAM=kernel
GOWORK=off go run ./cmd/goalcli score --min 9.8
```

生成的库必须保持测试独立于 `x.go`，且不得读取 `/home/k8s/secrets/env/*`。

## TDengine WebSocket integration

真实 TDengine 连接测试使用 `integration` build tag，默认不运行。执行前必须显式导出环境变量，不要 `source` Markdown 格式的密钥说明文件，也不要在日志中打印 DSN 或密码。

```bash
export TAOSX_INTEGRATION=1
export TAOSX_TDENGINE_ENDPOINT=localhost:6041
export TAOSX_TDENGINE_USER=<user>
export TAOSX_TDENGINE_PASSWORD=<password>
export TAOSX_TDENGINE_DATABASE=<database>
GOWORK=off go test -tags=integration ./pkg/taosx -run TestTDengineWebSocketIntegration -count=1 -v
```

可选地使用完整 DSN：

```bash
export TAOSX_INTEGRATION=1
export TAOSX_TDENGINE_DSN='root:taosdata@ws(localhost:6041)/db'
GOWORK=off go test -tags=integration ./pkg/taosx -run TestTDengineWebSocketIntegration -count=1 -v
```

该测试通过 `database/sql` 加载官方 `github.com/taosdata/driver-go/v3/taosWS` driver，验证 `taosx.New(..., WithDriver(...))` 的 `Health` 和真实 `SHOW TABLES` 查询路径。
