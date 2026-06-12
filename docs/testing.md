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

当前仓库没有内置真实 TDengine 集成环境。真实 TDengine driver、容器化集成测试和性能压测属于后续适配器或发布管线扩展范围。新增这些能力前必须同步更新规格、设计、追溯矩阵和 CI 证据。
