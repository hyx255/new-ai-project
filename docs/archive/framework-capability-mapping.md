# Framework Capability Mapping

> 参考项目：[go-gin-api](https://github.com/xinliangnote/go-gin-api)
> 
> 本文档将 go-gin-api 的工程能力映射到本项目已确定的 ARCHITECTURE.md 架构中。
> go-gin-api 是"能力参考库"，不是模板——本项目架构优先于参考项目。

## 策略说明

| 策略 | 含义 |
|------|------|
| **Adopt** | 直接采用，能力映射到本项目对应位置 |
| **Adapt** | 采用核心思路，但根据本项目架构做适配 |
| **Defer** | 当前阶段不实现，但架构已预留接入点 |
| **Reject** | 不适合本项目架构或当前阶段，明确不采用 |

## 能力映射表

| Reference Capability | go-gin-api 实现 | 本项目对应位置 | 当前策略 | 原因 |
|---|---|---|---|---|
| **Gin** | 核心 HTTP 框架 | `internal/platform/http/` | **Adopt** | ARCHITECTURE.md 明确选定 Gin |
| **Config (Viper)** | `pkg/config` 基于 Viper | `internal/platform/config/` | **Adapt** | 采用 YAML + 环境变量覆盖，不引入 Viper（Skeleton 阶段保持简单） |
| **Logger (Zap)** | `pkg/zap` 封装 | `internal/platform/logging/` | **Adopt** | 已实现，Zap 结构化日志 + module/trace_id 字段 |
| **Error Code (errno)** | `pkg/errno` 错误码体系 | `internal/platform/errors/` | **Adopt** | 已实现基础错误分类；完整业务错误码 TBD |
| **Response** | `pkg/response` 统一响应 | `internal/platform/response/` | **Adopt** | 已实现统一 JSON 响应格式 |
| **Database (sql)** | `pkg/mysql` `pkg/sql` | `internal/platform/database/` | **Adapt** | 采用 `database/sql` + 多驱动；ORM 选型 TBD |
| **GORM** | `pkg/gorm` 直接集成 | TBD（`internal/platform/database/`） | **Defer** | ARCHITECTURE.md §17 明确 ORM 未选定；Repository 模式必须先于 ORM 确定 |
| **Redis** | `pkg/redis` 直接集成 | `internal/platform/cache/` | **Adapt** | 建立 Cache 抽象接口 + NoOp 实现；Redis 实现推迟到有明确需求时（ADR-003） |
| **Swagger** | `pkg/swagger` swaggo 注解 | `internal/platform/http/`（路由注册） | **Defer** | 需要 swag CLI 工具；本阶段建立注解规范，生成步骤在 Makefile 中定义 |
| **Prometheus** | `pkg/prometheus` 指标采集 | `internal/platform/metrics/` | **Adopt** | 标准 Go 可观测性基础设施 |
| **Trace** | `pkg/trace` 内部链路追踪 | `internal/platform/tracing/` | **Adapt** | 采用轻量 trace_id/request_id 传播，不引入完整分布式追踪（Jaeger/Zipkin） |
| **pprof** | `pkg/pprof` 性能分析 | `internal/platform/http/`（路由注册） | **Adopt** | Go 内建 `net/http/pprof`，仅在开发环境启用 |
| **CORS** | 中间件跨域支持 | `internal/platform/http/` middleware | **Adopt** | Web 应用基本需求 |
| **JWT/Auth** | `middleware/jwt` 认证中间件 | TBD | **Defer** | ARCHITECTURE.md §17 明确认证机制 TBD |
| **Graceful Shutdown** | `cmd/server` 信号处理 | `internal/bootstrap/` | **Adapt** | 统一在 bootstrap 中管理生命周期 |
| **Docker** | 多阶段 Dockerfile | `Dockerfile` | **Adopt** | 已创建多阶段构建 |
| **Cron** | `pkg/cron` 定时任务 | N/A | **Reject** | 当前无定时任务需求；违反 AGENTS.md "不臆造"原则 |
| **WebSocket** | WebSocket 支持 | N/A | **Reject** | 当前无实时通信需求 |
| **gRPC** | `pkg/grpc` RPC 框架 | N/A | **Reject** | Modular Monolith 进程内通信，不需要 RPC（ADR-001） |
| **GraphQL** | GraphQL API 层 | N/A | **Reject** | REST API 足够，GraphQL 引入复杂度过高 |
| **CRUD Generator** | `cmd/g` 代码生成 | N/A | **Reject** | Skeleton 阶段无需代码生成 |

## Deferred 组件的接入点

以下组件已推迟，但架构已预留清晰的接入点：

| 组件 | 接入点 | 接入条件 |
|------|--------|----------|
| GORM/ORM | `internal/platform/database/` | ADR 评估后选定，Repository 模式已就位 |
| Redis | `internal/platform/cache/` 实现 `Store` 接口 | 存在明确缓存需求时（ADR-003 §Redis 引入条件） |
| Swagger | `internal/platform/http/` 路由注册 | swag CLI 工具可用后 |
| JWT/Auth | `internal/platform/http/` middleware 链 | 认证需求明确后 |

## 与 go-gin-api 的关键架构差异

| 维度 | go-gin-api | 本项目 |
|------|-----------|--------|
| 架构风格 | 传统分层单体 | Modular Monolith（ADR-001） |
| 业务组织 | 按技术层分组 (handler/service/repo) | 按业务模块分组 (`internal/modules/<module>/`) |
| AI 能力 | 无 | Agent → Tool → Application Service（ADR-004） |
| 设备抽象 | 无 | Device Gateway → Adapter → Transport（ADR-002） |
| 操作模型 | 无 | Operation / Execution 分离（ADR-005） |
| 测试模型 | 传统测试 | 五类测试类型（ADR-006） |
| 数据库策略 | 单一 MySQL | MySQL/GreatDB 生产 + SQLite 本地（ADR-003） |

## 优先级约束

`text
AGENTS.md > ARCHITECTURE.md > ADR-001~006 > go-gin-api 参考实现
`

当参考项目与本项目架构冲突时，**必须以本项目架构为准**。
