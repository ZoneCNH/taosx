# taosx 快速开始

本指南用于本地复核 `github.com/ZoneCNH/taosx` 的 TDengine adapter MVA。

## 环境

- Go 版本遵循 `.tool-versions` 与 CI 配置：`1.23.x`。
- 发布、Evidence 和 release 语境必须显式使用 `GOWORK=off`。

## 最小验证

```bash
go test ./pkg/taosx ./contracts ./examples/...
```

示例覆盖：

- `examples/basic` 输出 module path。
- `examples/config` 输出脱敏密码。
- `examples/health` 在未注入真实 driver 时输出 `degraded`。

## 完整本地 gate

```bash
go test ./...
./scripts/check_boundary.sh
./scripts/check_contracts.sh
GOWORK=off go run ./cmd/goalcli score --min 9.8
```

`release/manifest/latest.json` 与校验和是运行时生成产物，按 `.gitignore` 不提交。
