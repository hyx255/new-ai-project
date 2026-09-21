# AI 原生软件项目

本仓库是一个 AI 原生软件项目的起点。

项目目标是建立一种工作方式：Human 负责需求、架构方向、关键决策和最终验收；AI Agent 负责大部分可重复的工程工作，包括需求分析、任务拆解、编码、测试、代码审查、构建、测试环境部署和自动验证。

当前尚未实现任何业务功能。

## 项目结构

```text
.
|-- AGENTS.md
|-- README.md
|-- ARCHITECTURE.md
|-- docs/
|   |-- product/
|   |-- architecture/
|   |-- modules/
|   `-- decisions/
|-- skills/
|   |-- project-management/
|   |-- api-development/
|   |-- database/
|   |-- testing/
|   |-- code-review/
|   `-- release/
|-- scripts/
|-- tests/
|   |-- unit/
|   |-- integration/
|   `-- e2e/
`-- .gitignore
```

## AI 原生开发理念

预期工作流如下：

```text
Human
  -> 需求、架构方向、关键决策、最终验收

AI Agent
  -> 需求分析
  -> 任务拆解
  -> 实现计划
  -> 编码
  -> 测试
  -> 代码审查
  -> 问题修复
  -> 构建
  -> 测试环境部署
  -> 自动验证
```

## 如何使用 Codex

每个任务开始时，应向 Codex 提供相关需求或决策上下文。Codex 在修改文件前，应阅读 `AGENTS.md`、`docs/` 下的相关文档，以及 `skills/` 中匹配的工作流。

对于较大的工作，应先让 Codex 产出或更新实现计划，再进入编码。对于审查任务，应让 Codex 使用 code-review Skill，优先关注缺陷、回归风险、缺失测试和未经批准的假设。

## `AGENTS.md` 的作用

`AGENTS.md` 是 AI Agent 的最高层项目说明。它定义项目原则、仓库结构、核心规则、Definition of Done，以及 Agent 必须停止并询问 Human 的情况。

## `docs/` 的作用

`docs/` 存放项目知识：

- `docs/product/`：产品需求、用户故事、验收标准。
- `docs/architecture/`：架构、技术规范、编码规则、API/数据库约定。
- `docs/modules/`：模块级规格、设计、API 和验收文档。
- `docs/decisions/`：架构决策记录。

## `skills/` 的作用

`skills/` 存放可复用的 AI Agent 工作流。Skill 描述如何完成稳定的工程活动，例如项目管理、API 设计、数据库设计、测试、代码审查和发布。

Skill 不承载业务知识；业务知识应放在 `docs/`。

## 后续开发流程

1. 在 `docs/product/` 中编写或更新产品需求。
2. 在 `docs/decisions/` 中记录关键架构或技术选择。
3. 当真实业务模块出现时，在 `docs/modules/` 中创建模块文档。
4. 使用 `skills/project-management/` 分析范围并创建实现计划。
5. 使用相关技术 Skill 完成实现。
6. 通过测试和代码审查完成验证。
7. 当项目具备可部署应用后，通过 release 工作流完成构建、测试环境部署和验证。

技术选择会在真实需求出现后再确定，因此当前刻意保持 TBD。
