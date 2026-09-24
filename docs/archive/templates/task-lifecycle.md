# Task Lifecycle

本文档定义 AI 研发任务的统一状态模型。

## 状态流转

```text
PROPOSED
  -> PLANNED
  -> IMPLEMENTING
  -> TESTING
  -> REVIEWING
  -> DONE
```

异常状态：

```text
BLOCKED
CANCELLED
```

## PROPOSED

### 进入条件

- 用户提出需求、变更请求或工程任务。

### 允许执行什么

- 阅读 `AGENTS.md`。
- 阅读相关 `docs/` 和 `skills/`。
- 使用 `docs/product/requirement-template.md` 澄清需求。
- 识别 Open Questions。

### 必须产生什么产物

- Requirement 草案，或问题清单。

### 退出条件

- Requirement 足够清晰，可进入 PLANNED。
- 或需求存在阻塞问题，进入 BLOCKED。
- 或用户取消，进入 CANCELLED。

## PLANNED

### 进入条件

- Requirement 已明确。
- 验收标准可测试。

### 允许执行什么

- 创建或更新 Module Spec。
- 创建或更新 Design。
- 创建或更新 API 文档。
- 创建或更新 Acceptance。
- 创建 Implementation Plan。
- 识别架构决策、数据库决策、API 决策和依赖风险。

### 必须产生什么产物

- Requirement。
- Implementation Plan。
- 如果任务涉及模块：Module Spec、Design、API、Acceptance。
- Open Questions，如有。

### 退出条件

- 用户批准 Plan，或 Agent 被授权直接执行。
- 如果关键问题未解决，进入 BLOCKED。

## IMPLEMENTING

### 进入条件

- Implementation Plan 已完成并获批准，或 Agent 被明确授权执行。
- 必要技术决策已存在。

### 允许执行什么

- 创建或修改代码。
- 创建或修改必要配置。
- 创建或修改测试。
- 更新与实现一致的文档。

### 必须产生什么产物

- 代码变更。
- 与实现一致的文档变更。
- 初步测试或验证记录。

### 退出条件

- 实现完成并准备测试，进入 TESTING。
- 如果发现需求、架构或依赖阻塞，进入 BLOCKED。

## TESTING

### 进入条件

- 实现已完成到可验证状态。
- 测试计划存在。

### 允许执行什么

- 将 Requirement、Acceptance、API、Database 测试点转换为测试用例。
- 执行 Unit Test、Integration Test、E2E Test 或人工验证。
- 分析失败原因。
- 修复代码缺陷或测试缺陷。

### 必须产生什么产物

- Test Report。
- 失败分析，如有。
- 回归测试结果。

### 退出条件

- 测试通过并记录结果，进入 REVIEWING。
- 失败原因需要需求或验收变更，进入 BLOCKED。

## REVIEWING

### 进入条件

- 测试已完成。
- Test Report 已存在。
- 代码变更可审查。

### 允许执行什么

- 对照 Requirement、Design、Acceptance 和 Tests 审查代码。
- 输出 Code Review Report。
- 修复审查发现的问题。

### 必须产生什么产物

- Code Review Report。
- 审查结论：APPROVED 或 CHANGES_REQUIRED。

### 退出条件

- Code Review 为 APPROVED，进入 DONE。
- Code Review 为 CHANGES_REQUIRED，回到 IMPLEMENTING 或 TESTING。
- 审查发现需求或架构阻塞，进入 BLOCKED。

## DONE

### 进入条件

- Requirement 已满足。
- Test Report 表明必要测试已通过或批准例外已记录。
- Code Review Final Status 为 APPROVED。
- 文档、代码和测试一致。

### 允许执行什么

- 汇报完成情况。
- 进入 Release Readiness，如任务需要发布。

### 必须产生什么产物

- 完成摘要。
- 验证证据。
- 已知风险或后续事项。

### 退出条件

- 无。DONE 是正常终态。

## BLOCKED

### 进入条件

- 需求不明确。
- 架构不明确。
- 技术决策缺失。
- 依赖缺失。
- 验收标准不可测试。
- 需要用户确认业务、技术、安全、数据或发布决策。

### 允许执行什么

- 汇总阻塞原因。
- 提出具体问题。
- 等待用户确认。

### 必须产生什么产物

- Blocked 原因。
- 需要用户回答的问题。
- 已完成和未完成范围说明。

### 退出条件

- 用户补充信息后回到 PROPOSED 或 PLANNED。
- 用户取消任务，进入 CANCELLED。

## CANCELLED

### 进入条件

- 用户取消任务。
- 需求被废弃。

### 允许执行什么

- 停止当前任务。
- 汇总已产生的文档或变更。

### 必须产生什么产物

- 取消原因。
- 已完成产物清单。
- 后续清理建议，如有。

### 退出条件

- 无。CANCELLED 是终态。
