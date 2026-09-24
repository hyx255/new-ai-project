# Skill：release

## 名称

release

## 描述

负责 Release Readiness 检查，包括 Requirement、Implementation、Test、Review、Build、Staging、Health Check、E2E 和 Release。

在技术栈确定之前，本 Skill 不绑定 Docker、Kubernetes、云厂商或具体命令。

## 适用场景

- 变更已准备进入发布前验证。
- 需要确认发布准备度。
- 需要生成构建产物。
- 需要测试环境部署或健康检查。
- 需要发布检查清单。

## 输入

- 已批准的发布范围。
- Requirement。
- Implementation Plan。
- Test Report。
- Code Review Report。
- 构建、测试和部署说明，如已存在。
- 与运行时和部署目标相关的 ADR，如已存在。
- 经批准的环境变量、凭据和访问说明，如已存在。

## 前置条件

- Requirement 已明确。
- Implementation 已完成。
- Test Report 已产出。
- Code Review Final Status 为 APPROVED。
- 如需实际发布，可部署应用、构建命令、部署目标、产物策略和凭据必须已批准。

## 执行步骤

1. 确认 Requirement 已明确且与实现一致。
2. 确认 Implementation 已完成且无未解决 BLOCKED 项。
3. 确认 Test Report 已完成且 Final Result 可接受。
4. 确认 Code Review Final Status 为 APPROVED。
5. 如项目已有构建命令，执行 Build；否则记录为 TBD。
6. 如 Staging 环境已批准，部署到 Staging；否则记录为 TBD。
7. 如 Health Check 已定义，执行 Health Check；否则记录为 TBD。
8. 如 E2E 已定义，执行 E2E；否则记录为 TBD。
9. 记录 Release Readiness 结果、风险和回滚说明。

## 检查项

- Requirement、Implementation、Test、Review、Build、Staging、Health Check、E2E、Release 链路状态清晰。
- 发布命令与项目文档一致。
- 除非存在已批准例外，否则测试必须在部署前通过。
- 构建产物可追踪。
- 测试环境验证覆盖健康检查和关键用户流程。
- 未在没有 ADR 的情况下假设 Kubernetes、Docker、云厂商或具体平台。

## 输出

- Release Readiness checklist result。
- 构建和产物摘要，如适用。
- 测试环境部署结果，如适用。
- 健康检查和 E2E 验证结果，如适用。
- 已知风险和回滚说明。

## Definition of Done

- 必需检查已通过，或已记录批准例外。
- 构建产物可追踪，如适用。
- 当测试环境存在时，已验证环境健康。
- 已记录发布风险和回滚说明。
- 未引入未经批准的平台或云厂商假设。

## 失败处理

- 如果 Requirement、Test Report 或 Code Review 缺失，停止 Release Readiness 并报告缺失项。
- 如果 format、lint、测试或构建失败，停止发布并报告失败步骤。
- 如果部署目标未定义，只产出 Release Readiness checklist。
- 如果缺少凭据或环境访问，记录阻塞项，不臆造绕过方案。

## 必须停止并询问用户的情况

- 必须选择部署平台、产物格式或环境。
- 部署需要尚未批准的凭据或外部访问。
- 发布检查失败，但有人请求跳过。
- 回滚、数据迁移或兼容性风险不清楚。
