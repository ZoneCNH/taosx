# taosx SQL

SQL 能力由 `Statement` 和 `Query` 表达，driver adapter 负责把它们映射到 TDengine driver。

## 边界

- `Exec` 用于 DDL 和非返回行 SQL。
- `Query` 用于返回 `Rows` 的查询。
- SQL 字符串不能为空。
- 用户输入不得直接拼接为 identifier。

schema helper 提供 `Identifier` allowlist，用于生成稳定 SQL。复杂 SQL 仍可由调用方构造，但必须在 adapter 层保留 context、错误归一和脱敏。
