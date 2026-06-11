# taosx Schema

schema helper 用于构造稳定、可测试的 TDengine schema SQL。

## Identifier

`ColumnSpec.Name` 和 `StableSpec.Name` 规则：

- 首字符必须是字母或 `_`。
- 后续字符只能是字母、数字或 `_`。
- 最大长度 64。

`RenderCreateStable` 会校验 `StableSpec.Name`、`Columns` 和 `Tags` 中的 identifier，输出顺序稳定，便于 golden 测试和下游 adapter 审计。
