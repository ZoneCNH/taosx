# taosx 下游采用

taosx 当前只声明本仓库 MVA 交付，不声明 `x.go` 或真实下游库已经采用。

## 采用条件

下游采用前必须提供：

- downstream repository 名称和 commit。
- 使用模式：`patch-only`、`dry-run` 或 `pr-plan`。
- gate 输出：`go test ./...`、`make contracts`、`make boundary` 和 release evidence 检查。
- rollback 方案。

证据 contract 位于 `contracts/downstream-adoption-proof.schema.json`。没有该证据时，`docs/downstream-matrix.md` 中的 taosx 状态只能保持本地 MVA 或计划态。
