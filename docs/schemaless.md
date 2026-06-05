# taosx Schemaless

`SchemalessPayload` 表达 TDengine schemaless 写入输入。

## 支持的协议

- `line`
- `telnet`
- `json`

payload 必须至少包含一行非空内容。driver adapter 负责检查当前 `DriverMode` 是否支持 schemaless 写入。

## 时间精度

`SchemalessPrecision` 允许值为 `hour`、`min`、`sec`、`ms`、`us`、`ns`。空值表示调用方不显式指定 precision，由 driver adapter 或 TDengine 默认值处理；非空值必须落在枚举范围内。

## 部分失败

partial failure contract 位于 `contracts/schemaless_failure_contract.md`。当 driver 返回 partial result 和 error 时，client 必须保留 `WriteResult`，并把错误归一为 `ErrorKindWrite`。adapter 不自动盲重试 partial write。
