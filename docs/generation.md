# 生成模板

## 用途

`scripts/render_template.sh` 用于把 `baselib-template` 渲染为具体基础库，例如 `foundationx`。脚本负责同步替换 module name、module path、package name、`pkg/` 目录名、imports、文档占位符和脚本中的模板名称。

## 示例

```bash
scripts/render_template.sh \
  --module-name foundationx \
  --module-path github.com/ZoneCNH/foundationx \
  --package-name foundationx \
  --out ../foundationx
```

`--out` 必须指向不存在或为空的目录，避免覆盖已有仓库内容。

## 渲染范围

- `{{MODULE_NAME}}` 替换为 `--module-name`。
- `{{MODULE_PATH}}` 和 `github.com/ZoneCNH/baselib-template` 替换为 `--module-path`。
- `{{PACKAGE_NAME}}`、`pkg/templatex` 和 `templatex` imports 替换为 `--package-name`。
- 文档、Go 代码、JSON contract、shell 脚本、Makefile 和 CI 配置同步更新。

脚本不会复制 `.git`、`.omx`、`.worktree` 和 `release/manifest/latest.json`。`latest.json` 是生成产物，生成后的库必须自己运行 release gate 生成新的 Evidence artifact。

## 验证

生成后至少运行：

```bash
GOWORK=off make release-check
```

模板自身的 `make integration` 会渲染一个临时 `foundationx`，并运行 `GOWORK=off go test ./...`，用于防止生成脚本、包路径和 imports 回归。

## 边界

生成后的基础库仍必须保持独立，不能依赖 `github.com/bytechainx/x.go`、`github.com/ZoneCNH/x.go` 或任何 `x.go/internal/*` 包。
