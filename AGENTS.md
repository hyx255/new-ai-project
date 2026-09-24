# AI Agent 项目指南

## 项目

智能广播设备管理与运维平台 —— AI 原生工程项目。

## 技术栈

- Backend: Go + Gin
- Frontend: Vue 3 + TypeScript + Vite
- Database: MySQL/GreatDB (生产) + SQLite (开发)
- 架构: Modular Monolith

## 目录结构

```
.
├── cmd/server/              # 应用入口
├── internal/
│   ├── modules/             # 业务模块
│   ├── bootstrap/           # 启动编排
│   └── platform/            # 基础设施
├── web/                     # 前端应用
├── docs/
│   ├── requirements/        # 业务需求
│   ├── decisions/           # 架构决策（ADR）
│   ├── architecture/        # 架构说明
│   └── archive/             # 历史文档
└── skills/                  # AI 开发能力
```

**业务模块内部结构**:

```
internal/modules/<module>/
├── api/           # HTTP handler、DTO、路由
├── service/       # 业务逻辑
├── repository/    # 数据持久化
├── gateway/       # 设备能力接口定义
├── adapter/       # 协议/传输实现
└── model/         # 业务实体
```

## 架构原则

1. **Modular Monolith**: 业务模块是一级组织单位
2. **Defense in Depth**: Tool → Service → Model，每层独立校验
3. **Device Abstraction**: Service → Gateway → Adapter → Transport
4. **Operation/Execution**: Operation = 业务意图, Execution = 设备执行单元
5. **Audit**: Append-only，独立于 Operation/Execution

详见: `ARCHITECTURE.md`, `docs/decisions/ADR-001~006`

## 开发原则

- **不臆造**: 不创造未定义的需求、API、数据模型
- **小步快跑**: 优先做小变更，提供验证证据
- **文档一致**: 行为变化时同步更新文档
- **测试真实**: 测试验证预期行为，不是固化实现
- **Code Review**: 不为让测试通过而修改测试预期

## Skill 使用

Skill 描述"怎么做"。开发时按需调用：

- `feature-development`: 新功能开发
- `api-development`: API 设计与实现
- `bug-fix`: Bug 修复
- `database`: 数据库设计与迁移
- `testing`: 测试策略与执行
- `code-review`: 代码审查
- `release`: 发布准备

详见: `skills/*/SKILL.md`

## 测试要求

- Unit Test: SQLite/内存，快速验证
- Integration Test: MySQL，验证真实行为
- Adapter Contract Test: 验证 Gateway 实现
- E2E Test: 完整业务流程
- Agent Tool Test: 验证 Defense in Depth

详见: `docs/decisions/ADR-006-testing-strategy.md`

## 禁止事项

- 不修改未批准的技术栈/数据库/部署平台
- 不创建/修改公开 API 契约（除非已批准）
- 不创建/修改数据模型或迁移（除非已批准）
- 不在测试预期与需求不一致时修改测试
- 不引入未经批准的依赖

## 停止并询问

遇到以下情况必须停止并询问用户：

- 需求缺失或歧义影响系统行为
- 需要选择新技术栈/数据库/部署平台
- 变更会影响公开 API 或数据模型（未批准）
- 测试预期与需求不一致
- 用户请求与已记录决策冲突

## 文档入口

- **项目规则**: `AGENTS.md`（本文件）
- **架构**: `ARCHITECTURE.md`
- **需求**: `docs/requirements/device.md`
- **决策**: `docs/decisions/`
- **能力**: `skills/`

## Definition of Done

任务完成条件：

- 范围清晰，与需求/指令一致
- 实现、文档、测试一致
- 已验证并记录结果
- 已记录风险、限制、TBD
- 未引入未批准的技术/业务假设