# Skill：code-review

## 名称

code-review

## 描述

负责从 Requirement、Design、Acceptance、Tests、架构适配、正确性、代码质量、安全、并发、异常处理、性能、兼容性和测试覆盖角度审查代码。

## 适用场景

- 代码变更已准备好接受审查。
- 风险区域需要重点检查。
- 测试失败修复需要验证。
- 发布候选需要最终工程审查。

## 输入

- Diff 或已变更文件。
- Requirement、Acceptance Criteria 和 Implementation Plan。
- Design、API 文档、模块文档和 ADR。
- Test Report。
- `docs/code-review-template.md`。
- 已知风险。

## 前置条件

- 预期行为已知。
- Test Report 已存在，或已说明无法测试的原因。
- 审查者可访问变更代码和相关文档。
- 生成文件或无关变更在可能时已分离。

## 执行步骤

1. 识别变更范围和意图。
2. 对照 Requirement、Design、Acceptance 和 Tests 检查实现。
3. 审查架构边界和依赖方向。
4. 检查正确性、边界场景、并发、事务和异常处理。
5. 检查安全、输入校验、密钥、授权和敏感数据处理。
6. 检查预期规模下的性能和资源使用。
7. 检查兼容性和 Migration 风险。
8. 检查测试是否具备有意义的覆盖和正确预期。
9. 使用 `docs/code-review-template.md` 按严重程度报告问题，并在可用时提供文件和行号引用。

## 检查项

- 代码只实现已批准行为。
- 代码满足 Requirement、Design、Acceptance 和 Tests。
- 未引入未经批准的技术或架构决策。
- 错误路径和边界场景已处理。
- 安全敏感行为明确且有测试。
- 测试验证行为，而不是实现偶然性。
- 不得仅为了让测试通过而修改测试预期。
- Code Review Final Status 只能是 APPROVED 或 CHANGES_REQUIRED。

## 输出

- Code Review Report。
- Findings，严重程度只能是 BLOCKER、HIGH、MEDIUM、LOW。
- Questions。
- Risks。
- Verification。
- Final Status：APPROVED 或 CHANGES_REQUIRED。

## Definition of Done

- 高置信度的正确性、安全和回归风险已报告。
- must-fix 问题可执行。
- 非阻塞建议已单独列出。
- 审查结论明确：APPROVED 或 CHANGES_REQUIRED。

## 失败处理

- 如果 diff 过大，请求缩小审查范围，或按区域拆分问题。
- 如果需求缺失，只审查技术风险，并将行为验证标记为 BLOCKED。
- 如果测试与需求不一致，不得为了通过而修改测试；必须升级该不一致。

## 必须停止并询问用户的情况

- 代码变更包含未被已批准需求覆盖的行为。
- 测试预期似乎编码了错误的产品行为。
- 安全、数据丢失或兼容性风险需要产品或架构决策。
- 审查需要凭据、私有系统或破坏性操作。
