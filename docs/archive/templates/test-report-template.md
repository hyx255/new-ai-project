# Test Report Template

## Task ID

TBD

## Test Scope

TBD

说明本次测试覆盖的 Requirement、Acceptance、API、Database 或回归范围。

## Environment

TBD

说明测试环境、运行方式和必要配置。不得假设未批准的技术栈。

## Test Cases

```text
Case ID | Source | Type | Scenario | Expected Result | Actual Result | Status
TBD     | TBD    | TBD  | TBD      | TBD             | TBD           | TBD
```

Type 允许值：

- Unit
- Integration
- E2E
- Manual

Status 允许值：

- PASSED
- FAILED
- SKIPPED

## Passed

TBD

## Failed

TBD

## Skipped

TBD

## Failure Analysis

TBD

每个失败项必须分析原因：

```text
Failure ID: TBD
Related Case: TBD
Category: Code Defect | Test Defect | Environment Issue | Requirement Ambiguity
Analysis: TBD
Next Action: TBD
```

## Regression Result

TBD

说明修复后的回归测试范围和结果。

## Final Result

TBD

允许值：

- PASSED
- FAILED
- BLOCKED

## Rules

- 测试失败时，Agent 必须分析失败原因。
- 禁止删除测试来规避失败。
- 禁止降低测试标准。
- 禁止修改 Expected Result 以适应错误代码。
- 只有 Requirement 本身发生批准变更时，才允许更新 Expected Result。
