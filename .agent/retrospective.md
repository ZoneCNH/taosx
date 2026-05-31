# 复盘

## 改进项

- 基础库创建从手工变成模板化。
- 后续 `foundationx`、`postgresx`、`kafkax`、`redisx` 可复用目录、脚本、CI、文档和证据。

## 失败项

- 最终验证运行中没有必需门禁失败。
- 可选的 `govulncheck` 因本地未安装而跳过。

## 提示补丁

- 后续创建基础库时必须从 `baselib-template` 复制。
- 所有基础库必须保留边界门禁和密钥门禁。

## Harness 补丁

- 后续加入公共 API 哈希门禁。
- 后续加入配置 schema 哈希门禁。

## 规则补丁

- 禁止基础库依赖 `x.go`。
- 禁止基础库承载业务语义。
- 禁止无证据声称 `DONE`。

## CI 门禁建议

- 加入 CodeQL。
- 加入 govulncheck 强制模式。
- 加入覆盖率阈值。

## 新 Issue 候选

- ISSUE-FOUNDATIONX-001 从 `baselib-template` 生成 `foundationx`。
- ISSUE-POSTGRESX-001 从 `baselib-template` 生成 `postgresx`。
