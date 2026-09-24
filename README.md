# AI 原生广播设备管理平台

智能广播设备管理与运维平台 —— AI Coding 实验项目。

## 项目状态

- ✅ 项目基础工程完成（Go + Gin + Vue 3）
- ✅ 架构基线确定（Modular Monolith）
- ✅ Device MVP 需求冻结
- ✅ Backend/Frontend Foundation 完成
- 🔄 准备进入业务模块开发

## 技术栈

| 层次 | 技术 |
|------|------|
| Backend | Go + Gin |
| Frontend | Vue 3 + TypeScript + Vite |
| Database | MySQL/GreatDB (生产) + SQLite (开发) |
| Architecture | Modular Monolith |

## 快速开始

```bash
# Backend
go mod download
go run ./cmd/server

# Frontend
cd web
npm install
npm run dev
```

访问 http://localhost:5173 查看 Dashboard。

## 项目结构

```
.
├── cmd/server/              # 应用入口
├── internal/
│   ├── modules/             # 业务模块
│   └── platform/            # 基础设施
├── web/                     # 前端应用
├── docs/
│   ├── requirements/        # 业务需求
│   ├── decisions/           # 架构决策
│   └── architecture/        # 架构说明
└── skills/                  # AI 开发能力
```

## AI Coding 工作方式

```
Human → 需求、架构方向、关键决策、验收
AI Agent → 需求分析、任务拆解、实现、测试、审查、构建、部署
```

**开始任务时**：
1. AI 阅读 `AGENTS.md` 了解项目规则
2. AI 阅读 `docs/requirements/` 了解需求
3. AI 调用 `skills/` 中的能力执行开发

**文档优先级**：
```
AGENTS.md → ARCHITECTURE.md → docs/requirements/ → skills/ → 代码
```

## 当前业务

**Device Module MVP**:
- 设备类型管理
- 设备注册与管理
- 设备状态查询
- 设备操作（QueryStatus, SetVolume, Restart）
- 操作追踪与审计
- Device Simulator

详见: `docs/requirements/device.md`

## 架构决策

已确定的关键决策：
- ADR-001: Modular Monolith
- ADR-002: Device Abstraction Layer
- ADR-003: Database Strategy
- ADR-004: Agent Tool Boundary
- ADR-005: Operation/Execution
- ADR-006: Testing Strategy

详见: `docs/decisions/`

## 开发能力

项目提供以下 AI 开发 Skill：
- `feature-development`: 新功能开发
- `api-development`: API 设计与实现
- `bug-fix`: Bug 修复
- `database`: 数据库设计与迁移
- `testing`: 测试策略与执行
- `code-review`: 代码审查
- `release`: 发布准备

详见: `skills/*/SKILL.md`