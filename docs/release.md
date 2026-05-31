# 发布模板

## 占位符

- `{{MODULE_NAME}}`
- `{{MODULE_PATH}}`
- `{{PACKAGE_NAME}}`

## Release Gate

- `make ci`
- `make integration`
- `make evidence`

推荐入口是：

```bash
GOWORK=off make release-check
```

`GOWORK=off` 用于证明模板不依赖父级 workspace。

## Evidence

发布 Evidence 生成到 `release/manifest/latest.json`，该文件是生成产物，不提交到源码历史。提交到仓库的是 `release/manifest/template.json`；CI release workflow 会上传 `latest.json` 作为 artifact。

`latest.json` 至少包含：

- `module`
- `version`
- `commit`
- `go_version`
- `generated_at`
- `generated_by`
- `tree_state`
- `checks`
- `artifacts`
- `notes`

`make release-check` 成功后会以 `CHECK_STATUS=passed` 生成 manifest。若单独运行 `make evidence`，未显式传入的检查状态默认为 `unknown`。因为 `latest.json` 不再提交，manifest 中的 `commit` 可以指向实际执行 release gate 的 HEAD，避免自引用提交哈希导致的永久漂移。

`make integration` 会调用 `scripts/render_template.sh` 生成临时 `foundationx`，并在生成目录内运行 `GOWORK=off go test ./...`。这一步用于证明模板替换、包目录迁移和 imports 对齐仍然可用。

## 规则

- 没有 Evidence artifact 不得发布。
- `tree_state` 为 `dirty` 时可以生成 Evidence，但发布前必须明确说明未提交或生成中的文件。
- 不得在 release manifest、PR、Issue 或变更日志条目中包含原始凭据。
- 不得依赖 `github.com/bytechainx/x.go` 或 `github.com/ZoneCNH/x.go`。
