# Skill：testing

## 名称

testing

## 描述

负责把 Requirement、Acceptance、API 测试点和 Database 测试点转换为 Unit Test、Integration Test、E2E Test，并负责测试执行、失败分析、回归测试和 Test Report。

## 适用场景

- 新行为需要测试覆盖。
- 现有测试失败，需要分析。
- 修复后需要回归覆盖。
- 需要定义测试数据、fixture 或测试环境。
- 发布候选需要验证证据。

## 输入

- Requirement 和 Acceptance Criteria。
- Implementation Plan。
- `docs/test-report-template.md`。
- 现有测试和失败日志。
- 相关架构、模块、API 或数据库文档。
- API Development Skill 产出的 API 测试点。
- Database Skill 产出的数据库相关测试点。

## 前置条件

- 预期行为已知。
- 测试层级与风险匹配。
- 测试工具已选择，或当前任务仅产出框架无关的测试计划。

## 执行步骤

1. 将 Requirement、Acceptance、API 测试点和 Database 测试点映射为测试用例。
2. 选择最小但有效的测试层级：Unit Test、Integration Test、E2E Test 或人工验证。
3. 定义测试数据和环境需求。
4. 编写或更新测试，覆盖成功路径、失败路径、边界场景和回归场景。
5. 运行相关测试。
6. 通过对照 Requirement、Acceptance、测试和实现来分析失败。
7. 代码缺陷修改代码，测试缺陷修改测试，需求缺口进入 BLOCKED 并请求确认。
8. 使用 `docs/test-report-template.md` 记录验证结果。

## 检查项

- 测试断言的是已批准行为。
- 测试在其层级内足够确定且隔离。
- 除非明确文档化，否则测试数据不依赖隐藏外部状态。
- 不静默跳过失败测试。
- 不删除测试来规避失败。
- 不降低测试标准。
- 不修改 Expected Result 以适应错误代码，除非 Requirement 本身发生批准变更。

## 输出

- 测试计划或测试用例。
- 当技术栈存在时，输出自动化测试。
- Test Report。
- Failure Analysis。
- Regression Result。

## Definition of Done

- 相关 Acceptance Criteria 已覆盖。
- API 和 Database 测试点已被纳入测试计划或说明不适用。
- 测试成功运行，或失败已解释并给出下一步。
- 回归风险已处理。
- Test Report 已产出，并可供 Code Review 使用。

## 失败处理

- 如果预期行为不清楚，停止并请求澄清。
- 如果测试失败，将原因分类为代码缺陷、测试缺陷、环境问题或需求歧义。
- 如果环境无法运行测试，记录阻塞原因，并提供最接近的静态验证。
- 如果失败原因是 Requirement 或 Acceptance 不明确，停止并进入 BLOCKED。

## 必须停止并询问用户的情况

- 测试预期与已批准需求冲突。
- 失败测试暗示需求本身错误或不完整。
- 运行测试需要未批准的外部服务、费用、凭据或破坏性操作。
