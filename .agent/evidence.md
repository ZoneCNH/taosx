# 证据

2026-06-01 采集的完成证据：

- `go test ./...`：通过。
- `go test -race ./...`：通过。
- `make boundary`：通过。
- `make security`：通过；本地未安装 `govulncheck`，密钥扫描已通过。
- `make contracts`：通过。
- `make evidence`：通过，并生成 `release/manifest/latest.json`。
- `make release-check`：通过。

最终声明必须使用：

DONE with evidence:
