# taosx 宪章

## 0. 权威顺序

当文档之间冲突时，按以下顺序判定当前仓库事实：

1. `pkg/taosx` 公开 API 与测试。
2. `contracts/` 契约测试。
3. `README.md`、`docs/identity.md`、`docs/design.md`、`docs/api.md`、`docs/config.md`、`docs/spec.md`。
4. `release/` 中的发布证据。

任何文档不得高于代码和契约测试声称额外已实现能力。

## 1. 模块身份

`taosx` 是 TDengine L2 存储适配器契约模块，不是完整 TDengine 客户端平台。当前版本交付 Go 侧 client/driver port、配置、payload、健康检查、错误和 metrics 接口。

默认 client 使用不可用驱动。没有调用方注入 `Driver` 时，运行期操作必须返回明确的 unavailable 错误，而不是假装连接成功。

## 2. 依赖纪律

核心 `pkg/taosx` 不得直接依赖未批准的 Zone 模块。按上游 foundation 矩阵，`taosx` 的直接 foundation 依赖边界是 `kernel` only；如果未来需要新增直接依赖，必须先更新上游矩阵、规格、追溯矩阵和测试证据。

横切能力通过接口注入或上层适配器组合，不在核心包内硬编码。

## 3. 行为纪律

- 所有外部 I/O 风险路径必须接受 `context.Context`。
- 配置、错误、健康状态、日志和示例不得泄漏密码或私有端点。
- `Close` 必须幂等。
- 原始 SQL 只做空值校验，不对调用方声明注入防护。
- `MaxRetries` 是配置契约字段；核心 client 当前不自动重试。

## 4. 发布纪律

发布前必须收集可复现证据：

- `go test ./...`
- `git diff --check`
- 边界检查脚本。
- 契约检查脚本。

不能运行的检查必须记录为 `Not-tested`，不能替换成口头保证。

## 5. 禁止事项

- 禁止提交真实 TDengine 凭证、生产 DSN、账户 ID 或私有端点。
- 禁止在文档中把待实现能力写成已实现能力。
- 禁止为了通过检查而放宽契约测试或删除边界约束。
- 禁止在未更新规格和追溯矩阵的情况下改变公开 API。
