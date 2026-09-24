# ADR-004：Agent Tool Boundary（Agent Tool 边界）

## 状态

Accepted

## 日期

2026-09-22

## Context

项目未来将引入 AI Agent（Broadcast Operations Agent）来辅助广播设备运维。Agent 需要能够操作系统完成运维任务，但同时必须受控，不能绕过业务规则。面临以下约束：

- Agent 需要具备操作设备的能力（设置音量、查询状态、重启设备等）
- Agent 的操作必须经过权限检查和设备能力确认
- Agent 不得绕过业务层直接访问数据库或设备协议
- Agent 的操作必须可审计、可追溯
- Application Service 必须独立于 Agent 存在（Web API 和 CLI 也应使用相同的 Application Service）
- Tool 是 Agent 与系统交互的唯一入口
- Tool 的设计必须为 Agent 的输入校验和意图翻译提供明确的边界

## Decision

### 三层边界模型（Defense in Depth）

``text
Broadcast Operations Agent
        ↓ 自然语言/结构化意图
Agent Tool              ← 第一道防线：Tool Permission + 参数校验 + Agent Scope + Capability 初步确认
        ↓ 结构化命令
Application Service     ← 最终防线：最终业务授权 + Resource Scope + Business Rule + Execution 创建 + 状态校验
        ↓
Operation / Execution / Repository
``

**纵深防御原则**：Tool 和 Application Service 各自独立执行安全检查。Tool 已做的检查不能免除 Application Service 的校验责任。Application Service 不得信任调用者——无论请求来自 API、Agent 还是 CLI。

### 各层职责

**Agent（智能体）**：

- 理解运维意图（来自用户指令或自动策略）
- 选择合适的 Tool 执行操作
- 组合多个 Tool 完成复杂任务
- 处理 Tool 返回的结果和错误
- **禁止**：不得直接调用 Application Service，不得直接访问数据库，不得直接操作设备

**Agent Tool（工具层）—— 第一道防线**：

- **Tool Permission**：验证 Agent 是否有权调用此 Tool
- **参数校验**：校验输入参数的类型、范围、必填项
- **Agent Scope**：验证 Agent 的作用域限制（如只能操作特定区域）
- **Capability 初步确认**：查询目标设备 Capability，快速过滤明显不兼容的操作
- 调用一个或多个 Application Service 方法
- 返回结构化结果（成功/失败/部分成功）
- **禁止**：不得包含业务逻辑（如判断某操作在当前业务上下文中是否允许）
- **重要**：Tool 的校验是快速拒绝，不是最终授权

**Application Service（应用服务层）—— 最终防线**：

- **最终业务授权**：独立判断当前操作者是否有权执行此操作（不信任 Tool 的权限检查结果）
- **Resource Scope**：验证目标资源是否在当前操作者的可操作范围内
- **Business Rule**：执行业务规则校验（如设备状态是否允许此操作、操作时间窗口是否合法）
- **Execution 创建**：创建 Operation 和 Execution 记录
- **状态校验**：验证操作前提条件（如设备在线、无冲突操作正在进行）
- 调用 Device Gateway 操作设备
- 返回业务结果
- **禁止**：不得知道调用者是谁（API 用户、Agent、CLI 应同等对待）
- **要求**：必须独立可测试，不依赖 Agent 或 Tool
- **核心**：Application Service 是安全边界的核心，即使 Tool 校验被绕过，Application Service 仍然能阻止非法操作

### Tool 设计原则

1. **单一职责**：每个 Tool 对应一个明确的运维操作
2. **第一道防线**：Tool 必须完成 Tool Permission、参数校验、Agent Scope 和 Capability 初步确认
3. **非最终授权**：Tool 的校验是快速拒绝，不是最终授权——Application Service 仍必须独立校验
4. **结构化输入输出**：Tool 的输入和输出必须使用结构化类型，不使用自由文本
5. **错误映射**：Tool 将 Application Service 的错误映射为 Agent 可理解的错误类型
6. **不信任传递**：Tool 传递给 Application Service 的调用上下文中必须包含操作者身份信息，以便 Application Service 独立校验

### Tool 分类

| 类别 | 示例 Tool | 说明 |
|------|-----------|------|
| 查询类 | ListDevices, GetDeviceStatus, GetCapabilities | 只读操作，获取设备信息 |
| 操作类 | SetVolume, RestartDevice, SyncConfig | 对设备执行操作 |
| 运维类 | BatchOperation, HealthCheck, Diagnostics | 批量或综合运维操作 |

### Agent 操作示例

> 用户指令："把三区所有支持音量控制的设备设置为 70"

Agent 执行路径：

``text
1. Agent 选择 Tool: ListDevices(area="三区")
   → 返回设备列表

2. Agent 选择 Tool: GetCapabilities(deviceIDs=[...])
   → 返回各设备能力

3. Agent 筛选支持 Volume 的设备

4. Agent 选择 Tool: SetVolume(deviceID, 70) × N
   → 每个 Tool 内部（第一道防线）：
     a. Tool Permission 校验
     b. 参数校验
     c. Agent Scope 校验
     d. Capability 初步确认
     e. 调用 ApplicationService.SetVolume()
   → Application Service 内部（最终防线）：
     f. 最终业务授权校验
     g. Resource Scope 校验
     h. Business Rule 校验
     i. 创建 Execution
     j. 返回结果

5. Agent 汇总结果，报告给用户
``

### 关键约束

- **Agent 只能通过 Tool 访问系统**——没有后门、没有直接 API
- **Tool 不能包含业务逻辑**——业务逻辑属于 Application Service
- **Application Service 不依赖 Agent**——Web API 和 CLI 使用相同的 Application Service
- **Tool 可以组合多个 Application Service**——但不得绕过它们
- **Defense in Depth**——Tool 做第一道防线（快速拒绝），Application Service 做最终防线（独立授权），两者互不信任
- **Application Service 不得信任调用者**——即使 Tool 已完成权限检查，Application Service 仍必须独立执行完整校验

## Alternatives Considered

### 1. Agent 直接调用 Application Service

- **优势**：减少一层抽象，开发更快
- **劣势**：Agent 可以绕过权限检查和能力确认，安全风险高，审计困难
- **否决原因**：Agent 直接调用 Application Service 意味着 Agent 可以传入任意参数，绕过业务规则

### 2. Agent 通过 HTTP API 操作系统（与 Web 前端相同接口）

- **优势**：复用现有 API 层，无需专门的 Tool 层
- **劣势**：API 面向人类用户设计，返回格式可能不适合 Agent 解析，且 Agent 无法获得结构化的中间状态
- **否决原因**：Agent 需要结构化的 Tool 接口，而非面向人类用户的 HTTP API

### 3. Agent 直接嵌入 Application Service 中

- **优势**：Agent 与业务逻辑紧密集成，效率最高
- **劣势**：Agent 和业务逻辑耦合，Agent 升级影响业务，业务变更影响 Agent
- **否决原因**：违反关注点分离原则

## Consequences

### 正面

- Agent 操作全部经过 Tool 层，审计完整
- Tool 层提供统一的参数校验和权限检查，不依赖 Agent 的"自觉"
- Application Service 独立可测试，不依赖 Agent 框架
- 新增 Agent 能力只需新增 Tool，不修改 Application Service
- Web API、CLI、Agent 三种入口共享同一个 Application Service

### 负面

- Tool 层引入额外的代码和测试
- Agent 的每个操作多一层调用开销
- Tool 的定义需要与 Application Service 保持同步

### 安全边界（Defense in Depth）

- **第一道防线（Tool）**：Tool Permission + 参数校验 + Agent Scope + Capability 初步确认——快速拒绝明显非法的请求
- **最终防线（Application Service）**：最终业务授权 + Resource Scope + Business Rule + 状态校验——即使 Tool 校验被绕过，Application Service 仍能阻止非法操作
- Tool 的输入参数校验是确定性的，不依赖 Agent 的输出质量
- Application Service 不信任任何调用者（API / Agent / CLI 同等对待）
- 所有通过 Tool 的操作都会产生 Operation 和审计记录
- 即使 Agent 被"越狱"，攻击者仍需突破 Tool 和 Application Service 两道独立防线

## Revisit Conditions

以下情况触发重新评估本决策：

1. **Tool 数量爆炸**：当 Tool 数量超过 50 个，维护 Tool 定义和测试的成本过高时，考虑引入 Tool 自动生成机制
2. **Agent 需要实时流式交互**：当 Agent 需要流式输出（如长时间操作的进度反馈），Tool 的请求-响应模式不足以满足时
3. **多 Agent 协作**：当需要多个 Agent 协作完成复杂任务时，当前单 Agent 的 Tool 模型可能不足
4. **Tool 需要跨系统操作**：当 Tool 需要调用外部系统（如其他管理平台 API），Tool 的职责边界需要重新定义
5. **Application Service 需要感知调用者**：当不同调用者（API/Agent/CLI）需要不同的业务行为时

## 验证方式

- 确认 Agent 代码中不存在直接 import Application Service 包的调用
- 确认每个 Tool 包含 Tool Permission、参数校验、Agent Scope 和 Capability 初步确认
- 确认 Application Service 独立执行最终业务授权、Resource Scope 和 Business Rule 校验
- 确认 Application Service 的所有方法可独立测试（不依赖 Agent 或 Tool）
- 确认 Agent Tool Test 覆盖纵深防御：Tool 校验通过但 Application Service 拒绝的场景
- 确认不存在 Application Service 信任 Tool 校验结果而跳过自身校验的情况
