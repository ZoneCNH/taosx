# TDengine Profile

taosx 使用 `DriverMode` 描述 TDengine 接入方式：

- `websocket`：默认模式，面向现代 TDengine driver adapter。
- `native_legacy`：为历史 native driver adapter 保留。
- `rest_sql_only`：只声明 SQL/REST 子集，不声明 schemaless 和完整批量写入能力。

本仓库不直接引入真实 TDengine driver 依赖。真实接入由下游或后续 adapter 层通过 `WithDriver` 注入，并必须保留 taosx 的错误、health、metrics 和脱敏 contract。
