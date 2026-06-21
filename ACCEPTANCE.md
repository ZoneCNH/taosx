# 验收标准

`taosx` 当前发布版本为 `v1.0.5`。本页是发布与验收的可读入口，所有具体 gate 仍以 `README.md`、`docs/release.md`、`docs/testing.md` 和 `docs/standard/acceptance-matrix.md` 为准。

## 发布门禁

- `GOWORK=off make release-check`
- `XLIB_CONTEXT=release_verify GOWORK=off make release-final-check`
- `XLIB_CONTEXT=release_verify GOWORK=off make release-preflight VERSION=v1.0.5`
- `make evidence`

## 验收矩阵

- P0：`XLIB_CONTEXT=local_write GOWORK=off make governance-check`
- P0：`GOWORK=off go run ./cmd/goalcli cli-contract`
- P0：`GOWORK=off go run ./cmd/goalcli issue-registry`
- P0：`GOWORK=off go run ./cmd/goalcli command-registry`
- P0：`GOWORK=off go run ./cmd/goalcli makefile-baseline`
- P1：`GOWORK=off make p1-governance-check`
- P1：`GOWORK=off go run ./cmd/goalcli policy-schema`
- P2：`GOWORK=off make p2-runtime-check`
- P2：`GOWORK=off go run ./cmd/goalcli downstream-adoption --repo kernel/configx --mode patch-only`

## 集成环境变量

真实 TDengine 集成测试必须从受控环境注入以下变量之一组：

- `TAOSX_TDENGINE_ENDPOINT`
- `TAOSX_TDENGINE_USER`
- `TAOSX_TDENGINE_PASSWORD`
- `TAOSX_TDENGINE_DATABASE`
- 或完整 `TAOSX_TDENGINE_DSN`

## 完成条件

- `pkg/taosx` 覆盖率达到 100.0%
- release gate 通过
- 文档、版本常量和 release manifest 保持同一版本锚点
- 测试和文档不输出原始密码或完整 DSN
