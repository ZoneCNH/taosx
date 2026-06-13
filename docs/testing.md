# taosx 测试策略

## 必跑检查

发布或合入前应优先运行：

```bash
go test ./pkg/taosx
go test ./contracts
go test ./examples/...
go test ./...
git diff --check
./scripts/check_boundary.sh
./scripts/check_contracts.sh
```

如果某个脚本在当前环境不可用，记录具体原因和替代验证，不要把未运行检查写成已通过。

## 单元测试重点

- `Config.Normalize`、`Validate`、`Sanitized`、`RedactedDSN`。
- 默认 unavailable driver 行为。
- `Exec`、`Query`、`WriteBatch`、`SchemalessWrite` 的边界校验。
- `Rows` 扫描、列信息和关闭行为。
- metrics 注入与 no-op 默认实现。
- 错误分类、操作名和敏感信息脱敏。
- `Close` 幂等和关闭后操作。

## 契约测试重点

`contracts/` 用来锁定下游可依赖行为：

- 默认 client 不会假装连接真实 TDengine。
- 注入 driver 后所有操作按接口委托。
- 空 batch 是 validation error。
- schemaless payload 必须有 lines 和合法协议。
- 健康状态和错误分类稳定。

## 集成测试边界

真实 TDengine WebSocket 集成测试通过 build tag 隔离，默认测试不会访问网络或读取本地凭据：

```bash
TAOSX_INTEGRATION=1 go test -tags=integration ./pkg/taosx -run TestTDengineWebSocketIntegration -count=1
```

测试只支持官方 `github.com/taosdata/driver-go/v3/taosWS` 路径。运行前必须注入以下环境变量之一：

- `TAOSX_TDENGINE_DSN`：完整 WebSocket DSN。
- `TAOSX_TDENGINE_ENDPOINT`、`TAOSX_TDENGINE_USER`、`TAOSX_TDENGINE_PASSWORD`、`TAOSX_TDENGINE_DATABASE`：由测试组装 DSN。

凭据必须来自本地受控 secret 文件或 CI secret store；测试失败信息不得输出原始密码或完整 DSN。容器化集成环境、性能压测、连接池、STMT 和生产级重试策略仍属于后续发布范围。
