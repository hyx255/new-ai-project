# ADR-005：Operation 与 Execution 概念分离

## 状态

Accepted

## 日期

2026-09-22

## Context

广播设备管理平台需要记录所有对设备的操作。面临以下约束：

- 运维操作存在两个不同的关注层面：业务意图（"批量设置三区设备音量为 70"）和设备级执行单元（"向设备 10.0.0.5 发送音量设置命令"）
- 批量操作可能涉及数十台甚至数百台设备，其中部分可能失败，需要记录每台设备的执行结果
- 失败的设备操作可能需要重试，但重试不应重复业务判断（如权限检查、能力确认）
- 审计需要两个层次：业务级审计（谁、何时、做了什么意图）和执行级审计（每台设备的执行结果）
- Agent 的操作也需要纳入 Operation / Execution 模型
- 需要支持 Operation 的查询、追溯、统计和报告

## Decision

### 概念定义

**Operation（业务操作）**：

高层业务意图的表达。代表"用户或 Agent 想做什么"。

| 属性 | 说明 |
|------|------|
| operation_id | 唯一标识 |
| type | 操作类型（SET_VOLUME, RESTART, HEALTH_CHECK, ...） |
| initiator | 发起者（用户 ID / Agent Session ID） |
| target | 目标描述（区域、设备组、单台设备） |
| parameters | 操作参数（音量值、配置内容等） |
| context | 操作上下文（来源、原因、关联任务） |
| status | 操作状态（PENDING / IN_PROGRESS / COMPLETED / PARTIALLY_FAILED / FAILED） |
| created_at | 创建时间 |
| completed_at | 完成时间 |
| audit_info | 审计信息 |

一个 Operation 可能产生一个或多个 Execution。

**Execution（设备执行）**：

一个可独立追踪、重试和记录结果的最小设备执行单元。代表"对某台设备实际做了什么"。

| 属性 | 说明 |
|------|------|
| execution_id | 唯一标识 |
| operation_id | 所属 Operation |
| device_id | 目标设备 |
| capability | 使用的 Capability |
| command | 具体命令参数 |
| retry_policy | 重试策略（次数、间隔、超时） |
| status | 执行状态（PENDING / RUNNING / SUCCESS / FAILED / RETRYING / SKIPPED） |
| retry_count | 已重试次数 |
| result | 执行结果（成功数据或失败原因） |
| started_at | 开始时间 |
| completed_at | 完成时间 |
| audit_info | 执行审计信息 |

### 层级关系

``text
Operation（"设置三区音量 70"）
    ├── Execution 1：设备 10.0.0.1 → SUCCESS
    ├── Execution 2：设备 10.0.0.2 → SUCCESS
    ├── Execution 3：设备 10.0.0.3 → FAILED → RETRYING → SUCCESS
    └── Execution 4：设备 10.0.0.4 → FAILED（设备不支持 Volume）
Operation 状态：PARTIALLY_FAILED
``

### 状态流转

**Operation 状态机**：

``text
PENDING → IN_PROGRESS → COMPLETED
                     → PARTIALLY_FAILED
                     → FAILED
``

- `PENDING`：Operation 已创建，尚未开始执行
- `IN_PROGRESS`：至少一个 Execution 开始执行
- `COMPLETED`：所有 Execution 成功
- `PARTIALLY_FAILED`：部分 Execution 成功，部分失败
- `FAILED`：所有 Execution 失败

**Execution 状态机**：

``text
PENDING → RUNNING → SUCCESS
                  → FAILED → RETRYING → RUNNING → SUCCESS / FAILED
                  → SKIPPED（如 Capability 不满足）
``

### 重试规则

- **重试粒度在 Execution 级别**，不在 Operation 级别
- 重试不重复业务判断（权限检查、能力确认已在 Operation 创建时完成）
- 重试策略由 Execution 的 `retry_policy` 定义
- 最大重试次数有上限（如 3 次），超过后标记为 `FAILED`
- 重试间隔可以配置（固定间隔或指数退避）

### 审计分离

- **Operation 级别审计**：记录谁、何时、什么意图、影响范围
- **Execution 级别审计**：记录对每台设备的具体操作、结果、错误信息
- 两层审计不重复，互为补充
- 审计信息一旦写入不可修改

### 架构归属

- Operation 和 Execution 在概念上分离，但在第一阶段均归属于 Application Service 层
- 当 Execution 需要独立调度（如异步执行、跨进程重试）时，才考虑将 Execution 提升为独立服务

## Alternatives Considered

### 1. 只记录 Operation，不区分 Execution

- **优势**：数据模型简单
- **劣势**：无法追溯每台设备的执行结果，批量操作失败时无法知道哪些设备需要重试
- **否决原因**：广播设备管理需要精确的设备级执行记录

### 2. 将 Operation 和 Execution 放在不同服务层

- **优势**：概念分离更彻底，Execution 可以独立调度和重试
- **劣势**：第一阶段增加不必要的架构复杂性
- **否决原因**：单体架构中 Application Service 可以管理两者，无需过度分层

### 3. 使用事件溯源（Event Sourcing）记录所有操作

- **优势**：完整的状态变化历史，支持时间旅行查询
- **劣势**：引入事件存储、投影等复杂基础设施，第一阶段不需要
- **否决原因**：过度设计，传统 CRUD + 审计日志已足够

## Consequences

### 正面

- Operation 是自然的工作单元，便于追踪、重试和统计
- Execution 的重试不影响 Operation 的业务完整性
- 批量操作可以清晰展示每台设备的执行状态
- 两层审计提供完整的操作追溯能力
- Agent 的操作也自然融入 Operation / Execution 模型

### 负面

- 两层状态管理增加数据模型复杂性
- 需要维护 Operation 和 Execution 的状态一致性
- 查询"某个 Operation 的完整执行结果"需要关联查询

### 数据一致性

- Operation 和 Execution 在同一事务中更新（单体架构中更容易保证）
- Operation 状态由 Execution 状态聚合得出，不单独维护
- 审计记录使用 append-only 写入，不参与业务事务

## Revisit Conditions

以下情况触发重新评估本决策：

1. **Execution 需要跨设备事务**：当一组设备的 Execution 需要原子性（全部成功或全部回滚）时
2. **Execution 规模过大**：当一个 Operation 产生的 Execution 数量超过 10,000，单 Application Service 处理效率不足时
3. **异步执行需求**：当 Execution 需要长时间运行（如小时级固件升级），需要独立的任务队列时
4. **分布式重试**：当 Execution 重试需要跨进程或跨节点调度时
5. **Operation 生命周期过长**：当 Operation 需要跨越数天甚至数周（如分批部署）时

## 验证方式

- 确认 Operation 创建时会为每台目标设备生成对应的 Execution
- 确认 Execution 重试不触发 Operation 级别的业务判断
- 确认 Operation 状态由 Execution 状态聚合得出
- 确认审计记录分别在 Operation 和 Execution 级别产生
- 确认 Agent 发起的操作也创建 Operation 和 Execution 记录
