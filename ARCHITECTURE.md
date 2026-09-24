# 架构基线

本文档是"智能广播设备管理与运维平台"的正式架构基线。
所有架构决策以 ADR 形式记录在 `docs/decisions/`，本文档为架构总览入口。

---

## 1. 项目定位

**项目名称**：智能广播设备管理与运维平台

**核心目标**：为广播行业提供统一的设备管理、运维监控和智能操作能力，并为未来 AI Agent 介入运维预留清晰接口。

**系统目标**：

- 统一管理多厂商、多协议的广播设备
- 提供安全、可审计的设备操作能力
- 支持 AI Agent 通过受控 Tool 执行运维任务
- 支持本地开发与生产部署两种运行形态

**目标用户**：广播运维工程师、系统管理员、AI Agent（未来）。

---

## 2. 技术栈

| 层次 | 选择 | 说明 |
|------|------|------|
| Backend 语言 | Go | 编译型、高并发、部署简单 |
| Backend 框架 | Gin | 轻量、成熟、社区活跃 |
| Frontend 语言 | TypeScript | 类型安全 |
| Frontend 框架 | Vue | 渐进式、易上手 |
| 生产数据库 | MySQL / GreatDB | 客户环境标配 |
| 本地/测试数据库 | SQLite | 零外部依赖 |
| 缓存 | Redis（条件性引入） | 仅在确有需求时引入 |
| 部署目标 | Linux / Windows / Docker | 多平台 |
| 未来 AI | Broadcast Operations Agent | 预留接口 |

> 参考项目：[go-gin-api](https://github.com/xinliangnote/go-gin-api)——仅参考其模块化、分层、通用基础设施的设计思想，不复制其代码或结构。

> 详细决策见 [ADR-003-database-strategy](docs/decisions/ADR-003-database-strategy.md)。

---

## 3. 系统架构：Modular Monolith

第一阶段采用 **Modular Monolith（模块化单体）**。不设计微服务，不提前拆分 Worker、Device Service 或 Agent Service。

``text
Frontend (Vue + TypeScript)
    ↓ HTTP/REST API
Go Backend (Single Process)
    ↓
Modular Monolith
    ├── Module A (独立业务边界)
    ├── Module B (独立业务边界)
    └── Shared Kernel (公共类型与契约)
``

核心原则：

- 业务模块是一级组织单位，模块内部再进行必要分层
- 模块间通过定义良好的 Application 层契约交互，禁止直接访问其他模块的 Infrastructure
- 每个模块可独立测试
- 只有当实际规模或技术边界证明需要时，才考虑拆分独立进程

> 详细决策见 [ADR-001-modular-monolith](docs/decisions/ADR-001-modular-monolith.md)。

---

## 4. 项目目录结构

`text
project-root/
├── cmd/                              # 应用入口（可执行文件）
│   └── server/
│       └── main.go
├── internal/                         # 不导出的应用代码
│   ├── modules/                      # ★ 业务模块（一级组织单位）
│   │   └── <module>/
│   │       ├── api/                  # HTTP handler、请求/响应 DTO、路由
│   │       ├── service/              # 业务逻辑、用例编排
│   │       ├── repository/           # 数据持久化（接口 + 实现）
│   │       ├── gateway/              # 设备能力接口定义（interface）
│   │       ├── adapter/              # 协议/传输实现（Simulator、真实设备）
│   │       └── model/                # 业务实体、枚举、状态、值对象
│   ├── bootstrap/                    # 应用启动编排（DI、路由注册）
│   └── platform/                     # 应用内部基础设施（§4.1）
│       ├── config/                   # 配置加载
│       ├── database/                 # 数据库连接与迁移
│       ├── logging/                  # 日志基础设施
│       ├── errors/                   # 标准错误定义与映射
│       ├── response/                 # 统一 API 响应结构
│       ├── http/                     # HTTP server、middleware
│       ├── tracing/                  # Trace ID 传播
│       ├── metrics/                  # 指标采集（预留）
│       └── cache/                    # 缓存抽象（预留）
├── migrations/                       # 数据库迁移文件（SQL）
├── configs/                          # 配置文件示例
├── docs/                             # 项目文档
├── scripts/                          # 自动化脚本
├── tests/                            # 测试
│   ├── unit/
│   ├── integration/
│   ├── contract/                     # Device Adapter Contract Test
│   ├── e2e/
│   └── agent/                        # Agent Tool Test
├── web/                              # 前端项目 (Vue + TypeScript)
│   ├── src/
│   │   ├── api/                      # 后端 API 调用层
│   │   ├── views/                    # 页面组件
│   │   ├── components/               # 通用 UI 组件
│   │   ├── stores/                   # 状态管理
│   │   ├── router/                   # 路由配置
│   │   └── types/                    # TypeScript 类型定义
│   └── ...
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
`

**模块内目录职责**：

| 目录 | 职责 | 依赖方向 |
|------|------|----------|
| pi/ | HTTP handler、参数解析、DTO、路由注册 | → service/ |
| service/ | 业务流程、业务规则、编排 Repository/Gateway | → model/、→ repository (interface)、→ gateway (interface) |
| 
epository/ | 数据持久化接口与实现 | → model/、→ platform/database |
| gateway/ | 设备能力接口定义（Go interface） | → model/ |
| dapter/ | Gateway 的具体实现（Simulator、真实设备） | → gateway (implements)、→ model/ |
| model/ | 业务实体、枚举、状态机、值对象 | 不依赖任何外层 |

**目录规则**：

- internal/modules/<module>/ 是业务代码的唯一归属位置
- 简单模块可以合并目录（如只保留 service/ 和 model/），复杂模块再增加完整目录
- 不要为了"架构完整"创建大量空目录
- internal/platform/ 只放应用级基础设施，**不包含任何业务逻辑**
- 模块间禁止循环依赖
- 模块间通信必须通过 Service 层，禁止直接访问其他模块的 Repository
- 不要在 internal/modules/ 之外创建业务代码
- 根目录 pkg/ 第一阶段**不创建**；只有当确实存在需要被外部 Go 项目 import 的公共库时才启用

### 4.1 Shared Kernel 约束

Shared Kernel 是各模块 service/ 层共享的接口和类型定义（如设备 ID 类型、通用枚举、跨模块事件契约），**不是**一个独立的包或目录。

**Shared Kernel 规则**：

- 只能存放真正跨模块共享且**稳定**的类型和契约
- 不允许承载业务逻辑（业务逻辑属于具体模块的 model/ 和 service/）
- 变更 Shared Kernel 类型必须评估对所有消费模块的影响
- Shared Kernel 类型可以放在发起方模块的 service/ 层，也可以提取到 internal/platform/ 中的共享子包（如 internal/platform/types/），但必须严格控制变更频率

### 4.2 internal/platform/ 职责

internal/platform/ 是应用内部的基础设施层，提供与业务无关的通用能力：

| 子包 | 职责 | 说明 |
|------|------|------|
| config/ | 配置加载与环境管理 | 支持环境变量、文件、命令行 |
| database/ | 数据库连接、迁移、健康检查 | 封装 MySQL/SQLite 驱动差异 |
| logging/ | 结构化日志 | 统一格式、关联 ID |
| errors/ | 标准错误定义与映射 | 业务错误码、HTTP 状态映射 |
| 
esponse/ | 统一 API 响应结构 | 分页、错误响应格式 |
| http/ | HTTP server 和 middleware | Gin 封装、CORS、Trace ID |
| 	racing/ | Trace ID 生成与传播 | 请求级追踪 |
| metrics/ | 指标采集 | 预留，第一阶段可不实现 |
| cache/ | 缓存抽象 | 预留，仅在 ADR 评估后实现 |

**依赖方向**：modules/ → platform/（模块依赖平台，平台不依赖模块）

## 5. 设备抽象架构

设备架构是本项目最重要的核心边界。业务层禁止依赖具体设备协议。

``text
Device Business       ← 业务规则、权限、能力判断、操作编排
        ↓
Device Gateway        ← 定义业务需要的设备能力接口（契约）
        ↓
Protocol Adapter      ← 将业务语义转换为具体厂商/协议的实现
        ↓
Transport             ← 连接、超时、重试、网络通信
        ↓
Real Device / Simulator
``

### 5.1 各层职责

| 层次 | 职责 | 禁止事项 |
|------|------|----------|
| Device Business | 设备业务规则、权限、能力判断、操作编排、审计 | 不得出现协议细节 |
| Device Gateway | 定义能力接口（Go interface）、能力查询、结果语义 | 不得包含协议转换 |
| Protocol Adapter | 业务语义 → 厂商/协议实现、协议解析、命令映射 | 不得包含业务规则 |
| Transport | TCP/UDP/HTTP 连接管理、超时、重试、心跳 | 不得包含业务语义 |

### 5.2 禁止出现在业务层的代码

- TCP / UDP 报文构造与发送
- HTTP 私有协议调用
- 厂商私有协议字段引用
- 二进制协议解析
- 协议特有的错误码直接处理

业务层应表达为：`SetVolume(deviceID, 70)`
而非：`SendTCPCommand(deviceID, rawBytes)`

### 5.3 扩展新协议的路径

1. 在 Transport 层实现或复用网络通信
2. 实现 Protocol Adapter，满足 Device Gateway 定义的 Contract
3. 通过 Adapter Contract Test
4. 注册到 Adapter Registry

> 详细决策见 [ADR-002-device-abstraction](docs/decisions/ADR-002-device-abstraction.md)。

---

## 6. Device Capability（设备能力）

每个设备声明自己支持的能力集合。业务操作前必须检查目标设备是否具备对应 Capability。

``text
Device
    ├── Capability: OnlineStatus
    ├── Capability: Volume
    ├── Capability: Playback
    ├── Capability: Restart
    └── Capability: ...
``

### 6.1 标准能力定义

| Capability | 说明 | 典型操作 |
|------------|------|----------|
| OnlineStatus | 设备在线状态查询 | GetStatus |
| Volume | 音量控制 | SetVolume, GetVolume |
| Playback | 播放控制 | Play, Pause, Stop |
| Restart | 设备重启 | Restart |
| ConfigSync | 配置同步 | SyncConfig, GetConfig |
| HealthCheck | 健康检查 | CheckHealth |

### 6.2 能力规则

- 设备注册时声明支持的 Capability 集合
- 业务操作前必须查询目标设备 Capability
- 缺少 Capability 时返回 `DEVICE_NOT_SUPPORTED` 错误
- Capability 定义在 Device Gateway 层，实现在 Adapter 层
- 未来 Agent Tool 必须基于 Capability 进行设备筛选和操作

### 6.3 Agent 操作示例

> "把三区所有支持音量控制的设备设置为 70"

Agent 执行路径：

1. 查询目标区域设备列表
2. 获取每台设备的 Capability
3. 筛选支持 Volume 能力的设备
4. 逐台调用 SetVolume(deviceID, 70)
5. 汇总执行结果

---

## 7. Device Status Model（设备状态模型）

架构中严格区分以下四种概念，不得混淆：

| 概念 | 说明 | 示例 |
|------|------|------|
| Current Device Status | 设备当前状态快照 | ONLINE, OFFLINE, FAULT, UNKNOWN |
| Device Status Event | 状态变化事件（时序记录） | 10:01 ONLINE → 10:15 OFFLINE → 10:20 ONLINE |
| Protocol / Communication Event | 通信层事件 | TIMEOUT, CONNECTION_RESET, INVALID_RESPONSE |
| Business Alarm | 业务级告警 | DEVICE_OFFLINE_TOO_LONG, HEALTH_CHECK_FAILED |

**概念边界**：

- Protocol Event 不直接等于 Business Alarm（需经过业务规则判断）
- Device Status Event 是状态变化的历史记录
- Current Status 是最新状态的快照
- Business Alarm 需要业务规则触发（如离线超过 N 分钟）

> 本任务不建表，仅明确概念边界。

---

## 8. Agent 与 Tool 边界

``text
Broadcast Operations Agent
        ↓ 意图表达
Agent Tool                ← 意图翻译：参数校验、权限检查、能力确认
        ↓ 结构化命令
Application Service       ← 业务逻辑：编排、判断、数据操作
        ↓
Operation / Execution / Repository
``

### 8.1 职责分离（Defense in Depth）

Tool 和 Application Service 各自独立执行安全检查，形成纵深防御。**Tool 已做的检查不能免除 Application Service 的校验责任**——Application Service 不得信任调用者，无论请求来自 API、Agent 还是 CLI。

| 层次 | 职责 | 禁止事项 |
|------|------|----------|
| Agent | 理解运维意图、选择 Tool、组合 Tool 完成任务 | 不得直接调用 Application Service |
| Tool | **Tool Permission**（Agent 是否有权调用此 Tool）、**参数校验**（类型/范围/必填）、**Agent Scope**（Agent 的作用域限制）、**Capability 初步确认** | 不得包含业务逻辑 |
| Application Service | **最终业务授权**、**Resource Scope**（目标资源是否在当前操作者范围内）、**Business Rule**（业务规则校验）、**Execution 创建**、**状态校验** | 不得知道调用者是谁（API/Agent/CLI 应同等对待） |

**纵深防御原则**：

- Tool 做**第一道防线**：快速拒绝明显非法的请求（参数错误、无 Tool 权限、设备无能力）
- Application Service 做**最终防线**：执行业务级权限判断（该用户能否操作该设备、操作是否符合业务规则、目标资源是否合法）
- 即使 Tool 校验通过，Application Service 仍必须独立执行完整校验
- 即使 API 入口已有中间件鉴权，Application Service 仍必须独立校验资源权限

### 8.2 核心约束

- **Agent 只能通过 Tool 操作系统**
- **Tool 是 Application Service 的入口，不是业务逻辑载体**
- Application Service 必须独立可测试，不依赖 Agent 或 Tool
- Tool 可以组合多个 Application Service 调用，但不得绕过它们
- 每个 Tool 有明确的输入参数、输出结果和错误定义

> 详细决策见 [ADR-004-agent-tool-boundary](docs/decisions/ADR-004-agent-tool-boundary.md)。

---

## 9. Operation 与 Execution

### 9.1 Operation（业务操作）

高层业务意图表达，如：

- "设置 A 区所有设备音量为 70"
- "检查设备 10.0.0.5 健康状态"
- "批量重启三楼广播设备"

Operation 包含：发起者、目标、参数、时间、上下文、审计信息。

一个 Operation 可能产生一个或多个 Execution。

### 9.2 Execution（设备执行）

一个可独立追踪、重试和记录结果的最小设备执行单元：

- 目标设备
- Capability 匹配结果
- 命令参数
- 重试策略（次数、间隔、超时）
- 状态流转：PENDING → RUNNING → SUCCESS / FAILED
- 执行结果

**关键规则**：

- 重试粒度在 Execution 级别，不在 Operation 级别
- Operation 记录业务审计，Execution 记录执行审计
- Operation 与 Execution 在概念上分离，但在第一阶段均归属于 Application Service 层

> 详细决策见 [ADR-005-operation-execution](docs/decisions/ADR-005-operation-execution.md)。

---

## 10. 数据库策略

| 环境 | 数据库 | 说明 |
|------|--------|------|
| 生产部署 | MySQL / GreatDB | 客户环境标配 |
| 本地开发 | SQLite | 零外部依赖，快速启动 |
| Unit Test | SQLite / 内存数据库 | 快速执行，不依赖外部服务 |
| Integration Test | MySQL | 验证真实 Repository 行为与生产一致性 |
| GreatDB 兼容性 | GreatDB | 验证客户环境兼容性（CI 阶段） |
| 缓存 | Redis（条件性） | 仅在确有需求时引入 |

**迁移策略**：

- 使用版本化 SQL 迁移文件
- 迁移文件必须幂等且仅向前
- 优先使用 MySQL / GreatDB 共同兼容的能力；SQLite 仅作为本地开发和快速测试便利，不以 SQLite 兼容性约束生产数据库设计
- 本地开发使用 SQLite，Integration Test 和 CI 验证使用 MySQL / GreatDB

**Redis 引入条件**：存在明确的缓存、限流或会话管理需求，且经过 ADR 评估后方可引入。

> 详细决策见 [ADR-003-database-strategy](docs/decisions/ADR-003-database-strategy.md)。

---

## 11. 日志策略

结构化日志，统一格式，统一输出：

| 字段 | 说明 | 必填 |
|------|------|------|
| timestamp | ISO 8601 时间戳 | 是 |
| level | 日志级别 | 是 |
| trace_id | 请求追踪 ID | 是 |
| module | 模块名称 | 是 |
| operation_id | Operation ID | 操作上下文 |
| execution_id | Execution ID | 执行上下文 |
| agent_session_id | Agent 会话 ID | Agent 操作上下文 |
| message | 日志消息 | 是 |
| fields | 结构化附加字段 | 否 |

**日志级别**：

- `DEBUG`：开发调试
- `INFO`：正常业务事件
- `WARN`：可恢复异常
- `ERROR`：需关注错误
- `FATAL`：系统不可用

**敏感信息**：密码、Token、密钥等敏感数据禁止出现在日志中。

**日志基础设施位置**：`internal/platform/logging/`

---

## 12. Device Simulator 架构位置

当前仅定义架构位置和职责，不实现 Simulator。

``text
Backend
    ↓
Device Gateway
    ↓
Protocol Adapter
    ↓
Device Simulator      ← 模拟设备协议响应
``

**未来用途**：

- Integration Test：测试时不需要真实设备
- Adapter Contract Test：验证 Adapter 实现的正确性
- E2E Test：端到端业务流程测试
- Agent Tool Test：测试 Agent Tool 的行为正确性

**归属**：`tests/simulator/` 或 `internal/testing/simulator/`（实现时确定）

---

## 13. 测试策略

五种测试类型，各自关注不同的验证目标（不是严格的流水线层级）：

| 测试类型 | 验证目标 | 关键验证点 |
|----------|----------|------------|
| Unit Test | 业务规则、权限、Capability、错误映射 | 领域模型行为、状态流转、算法正确性 |
| Integration Test | API → Application → Repository 完整链路 | 数据持久化、API 契约、数据库事务（使用 MySQL） |
| Adapter Contract Test | Adapter 满足 Gateway Contract | 所有 Adapter 通过统一契约测试套件 |
| E2E Test | 完整业务流程 | 端到端操作链路、跨模块协作 |
| Agent Tool Test | Agent 安全边界 | Tool 校验、Defense in Depth、Agent 隔离 |

**测试类型说明**：

- 不同测试类型关注不同验证目标，不要求严格的上下层级依赖
- Unit Test 可以在没有任何基础设施的情况下独立运行
- Integration Test 使用 MySQL 验证真实的 Repository 行为
- Adapter Contract Test 是 Adapter 合并的前置条件
- Agent Tool Test 验证纵深防御（Tool 校验 + Application Service 独立校验）

**核心原则**：

- 所有测试必须确定性、可重复执行
- Adapter Contract Test 是必须通过的测试类型
- 不允许使用 Simulator 替代 Contract Test
- Unit Test 使用 SQLite 或内存数据库；Integration Test 使用 MySQL

> 详细决策见 [ADR-006-testing-strategy](docs/decisions/ADR-006-testing-strategy.md)。

---

## 14. 前端架构

**技术栈**：Vue + TypeScript

**核心原则**：

- 前端通过 HTTP API 访问后端，不直接访问数据库或设备
- 前端不包含后端业务规则，仅负责展示和交互逻辑
- 权限展示由后端返回的能力信息决定
- 高风险操作必须展示：目标设备、风险说明、执行状态、结果、审计信息
- 第一阶段不引入复杂前端架构

**前端项目结构**：位于 `web/` 目录，详见目录结构。

---

## 15. 架构边界总结

以下是不可违反的架构边界：

``text
✗ 业务代码不得依赖具体设备协议（ADR-002）
✗ Agent 不得绕过 Application Service（ADR-004）
✗ Tool 不得包含业务逻辑（ADR-004）
✗ 模块间不得直接访问 Infrastructure 层（ADR-001）
✗ 模块间不得循环依赖（ADR-001）
✗ 前端不得绕过 API 访问后端内部（§14）
✗ 日志中不得出现敏感信息（§11）
``

**依赖方向**：

``text
interfaces → application → domain
infrastructure → application → domain
Tool → Application Service → Module
Frontend → Backend API → Module
``

---

## 16. 禁止过度设计

第一阶段禁止提前引入：

- 微服务拆分
- Kubernetes / Service Mesh
- 消息队列（Kafka / RabbitMQ / NATS）
- 多 Agent 系统
- 复杂 RAG / Memory / Event Bus
- 分布式事务
- 事件溯源（Event Sourcing）
- CQRS（除非有明确需求）

除非后续真实需求证明需要，否则不引入以上技术。任何引入必须通过 ADR 评估。

---

## 17. 仍然 TBD 的内容

以下决策将在后续任务中逐步确定：

- 具体业务模块划分（在需求明确后确定）
- Gin 框架配置细节（中间件、路由注册方式）
- ORM 选择（GORM / sqlx / 原生 database/sql）
- Migration 工具选择（golang-migrate / goose / 自研）
- 广播设备具体协议标准（IP 广播、传统广播、厂商私有协议）
- 认证/授权机制（JWT / Session / RBAC 策略）
- 部署配置细节（Docker 镜像、CI/CD Pipeline）
- AI Agent 框架选择（OpenAI API / 自研 / LangChain Go）
- Device Gateway 完整接口定义
- 完整错误码体系
- Operation / Execution 生命周期状态机
- Audit Log Schema
- Agent Tool 完整目录
- Go 测试框架选择（testify / go-cmp / 标准库）
- E2E 测试框架和运行环境
- Vue 版本（Vue 3）及状态管理（Pinia）、路由、UI 组件库
- 前端详细项目结构
- CI/CD Pipeline 设计
- Redis 引入的具体条件与时机
