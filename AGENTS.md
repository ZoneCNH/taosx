# 仓库指南

## 模块定位

`taosx` 是 `github.com/ZoneCNH/taosx` Go 模块，当前公开入口是 `pkg/taosx`。模块定位为 TDengine L2 存储适配器契约：提供配置、客户端端口、驱动端口、写入/查询数据结构、健康检查、错误分类和可选指标接口。

默认构造不会连接真实 TDengine。需要真实连接能力时，调用方必须通过 `taosx.WithDriver` 注入实现 `taosx.Driver` 的适配器。

## 项目结构

- `pkg/taosx/`：公开的 taosx API、配置、client、driver port、metrics port、错误和测试。
- `contracts/`：面向下游的契约测试，锁定公开行为。
- `examples/`：最小使用示例，必须避免提交真实凭证。
- `docs/`：身份、设计、API、配置、测试和发布说明。
- `scripts/`：边界检查和契约检查脚本。
- `release/`：发布证据与审计材料。

## 常用命令

- `go test ./pkg/taosx`
- `go test ./contracts`
- `go test ./examples/...`
- `go test ./...`
- `git diff --check`
- `./scripts/check_boundary.sh`
- `./scripts/check_contracts.sh`

发布前优先运行 `go test ./...`、边界脚本和契约脚本。若本地缺少某个发布工具，记录未运行原因，不要伪造证据。

## 发布版本锚点

当前发布文档锚点为 `v1.0.1`，必须与 `CHANGELOG.md` 最新版本、release manifest 和模板版本常量保持一致。

## 代码与边界规则

- 公共契约以 `pkg/taosx` 为准；文档不得声明不存在的真实驱动、连接池、STMT 或自动重试能力。
- 核心包不得直接引入未批准的 Zone 横切模块；当前直接 foundation 依赖边界按上游治理定义为 `kernel` only。
- 所有外部操作必须接受 `context.Context`，并保持 `Close` 幂等。
- 错误、健康状态和配置输出不得泄漏密码、token、DSN 密钥或真实生产端点。
- 原始 SQL 只做空值边界校验；参数化、防注入和 SQL DSL 属于具体驱动或上层调用方职责。
- 修改公开 API 时必须同步更新 `docs/api.md`、`docs/spec.md`、`contracts/` 和相关测试。

## 文档风格

文档默认中文，保留 Go 标识符、模块路径和 TDengine 技术名词。描述能力时使用当前可验证事实；未来能力必须标记为非目标、后续版本或待实现。
