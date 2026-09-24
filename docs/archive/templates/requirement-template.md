# Requirement Template

## Requirement ID

TBD

## Title

TBD

## Background

TBD

说明需求产生的业务背景、用户问题或机会。不要写技术实现方案。

## Goal

TBD

描述需求希望达到的业务结果。

## Scope

TBD

列出本需求包含的业务能力、用户行为和可验收范围。

## Out of Scope

TBD

列出本需求明确不包含的事项。

## User Story

TBD

建议格式：

```text
作为 <用户角色>
我希望 <完成某个目标>
以便 <获得某个业务价值>
```

## Functional Requirements

TBD

使用编号列出功能需求：

```text
FR-001: TBD
FR-002: TBD
```

## Business Rules

TBD

使用编号列出业务规则：

```text
BR-001: TBD
BR-002: TBD
```

## Dependencies

TBD

列出业务依赖、外部协作方、前置需求或其他已知依赖。不要假设具体技术依赖。

## Constraints

TBD

列出业务约束、合规约束、时间约束、兼容性约束或运营约束。

## Acceptance Criteria

TBD

每条验收标准必须可被后续测试验证。建议格式：

```text
AC-001: 给定 <前置条件>，当 <用户或系统行为>，则 <可观察结果>。
AC-002: TBD
```

## Open Questions

TBD

列出所有未决问题。任何影响行为、数据、权限、兼容性或验收标准的问题都必须在 Coding 前解决，或使任务进入 BLOCKED。

## Status

PROPOSED

允许值：

- PROPOSED
- PLANNED
- BLOCKED
- CANCELLED
- DONE

## Notes

- 产品需求不包含具体代码实现、框架、数据库、部署平台或云厂商选择。
- 未确定的信息必须标记为 TBD。
- 如果验收标准不可测试，需求不能进入实现阶段。
