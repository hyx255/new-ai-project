# Module Design Template

## Module

TBD

关联的 Module Spec：TBD

## Architecture

TBD

描述模块在系统中的技术边界、职责和依赖方向。

## Component Structure

TBD

描述组件、层次和职责。不要绑定未决技术栈。

## Data Flow

TBD

描述数据如何进入、处理、持久化或返回。

## API Flow

TBD

描述 API 请求从入口到响应的处理流程。

## Data Model

TBD

描述概念数据模型、数据约束和持久化需求。具体数据库实现必须依赖已批准 ADR。

## Transaction Boundary

TBD

描述事务边界、一致性要求和失败时的回滚要求。

## Cache Strategy

TBD

如不需要缓存，明确写为：不需要缓存。不得在没有需求或 ADR 的情况下引入缓存技术。

## Concurrency

TBD

描述并发场景、冲突处理、幂等性和一致性策略。

## Error Handling

TBD

描述错误分类、错误传播、用户可见错误和内部错误处理。

## Security

TBD

描述认证、授权、输入校验、敏感数据和审计要求。

## Performance

TBD

描述已知性能目标、容量假设和性能风险。

## Failure Recovery

TBD

描述失败后的恢复、补偿、重试、回滚或人工处理策略。

## Alternatives Considered

TBD

列出已考虑方案及未选择原因。

## Technical Risks

TBD

列出技术风险、影响和缓解措施。

## Notes

- Design 必须建立在已批准的 Spec 之上。
- 如果 Spec 存在未解决问题，不允许自行猜测，任务必须进入 BLOCKED。
- 任何技术栈、数据库、部署平台或外部服务选择必须依赖 ADR。
