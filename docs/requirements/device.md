# Device Module MVP Requirement

**Status**: PROPOSED (REVISED)
**Date**: 2026-09-23
**Phase**: Phase 3A — Product Requirement Discovery (Review)
**Revision**: v2 — 解决概念冲突，冻结关键业务决策

---

## 1. Product Goal

为广播运维团队提供第一个可演示的业务闭环：

> **设备管理 → 设备能力 → 设备状态 → 设备操作 → 操作结果**

使运维工程师能够在系统中完成设备注册、状态查询、基础操作执行，并获得完整的操作追踪和审计记录。

MVP 的核心价值不是"功能多"，而是**端到端可验证**：从用户发出指令到设备执行结果被记录，整条链路透明可追溯。

---

## 2. Target Users

| 用户角色 | 描述 | 核心诉求 |
|----------|------|----------|
| **广播运维工程师** | 日常管理和维护广播设备的一线人员 | 快速查看设备状态、执行基础操作、追踪操作结果 |
| **系统管理员** | 负责系统配置和设备类型管理 | 定义设备类型、管理设备注册 |

> 注意：AI Agent 不在 MVP 用户范围内。Agent 将在后续阶段引入（ADR-004）。

---

## 3. Problem Statement

当前广播设备管理面临以下痛点：

1. **设备信息分散**：不同厂商、不同型号的设备信息散落在各处，缺乏统一管理
2. **操作无法追踪**：谁在什么时间对哪台设备做了什么操作，结果如何——没有完整记录
3. **状态不透明**：设备是否在线、是否可用，需要人工逐台检查
4. **操作风险不可控**：高风险操作（如重启设备）缺乏确认机制和审计能力

MVP 解决的核心问题：**让设备管理从"人工记忆"变为"系统记录"。**

---

## 4. MVP Scope

### IN SCOPE

| 能力 | 说明 |
|------|------|
| Device Type 管理 | 定义设备类型及其支持的 Capability 集合 |
| Device 注册与管理 | 注册设备、查看设备列表、查看设备详情 |
| Capability 继承 | Device 自动继承 DeviceType 定义的 Capability，不支持单设备覆盖 |
| Status 查询 | 查询设备在线/离线状态（主动查询模式） |
| 基础设备操作 | 执行 QueryStatus、SetVolume、Restart 操作 |
| Operation/Execution 追踪 | 记录每次操作的意图和设备级执行结果 |
| Audit 审计 | 独立的不可变安全审计事件，记录操作者、动作和结果 |
| Device Simulator | 用 SimulatorAdapter + Simulator 替代真实设备，支持测试和演示 |
| 基础 Web UI | Dashboard、设备列表、设备详情、操作执行、结果查看 |

### OUT OF SCOPE

| 排除项 | 原因 |
|--------|------|
| 多厂商/多协议 | MVP 只通过 Simulator 验证抽象层，不接入真实协议 |
| 真实设备接入 | MVP 使用 Simulator 替代 |
| MQTT / WebSocket | 不需要实时通信 |
| 大规模设备管理 | MVP 验证流程，不验证规模 |
| 告警中心 | 后续阶段 |
| 复杂监控 | 后续阶段 |
| AI Agent / RAG / Memory | 后续阶段 |
| Multi-Agent | 后续阶段 |
| 用户认证/授权（JWT/RBAC） | MVP 使用简化操作者标识，不构成安全边界 |
| 批量操作 | MVP 先验证单设备操作，批量作为扩展 |
| 设备固件升级 | 后续阶段 |
| 设备分组/区域管理 | 后续阶段 |
| Device Capability Override | MVP 不支持单设备覆盖 Capability，后续通过独立需求扩展 |
| 心跳/滑动窗口/状态抖动抑制 | MVP 简化 OFFLINE 规则 |

---

## 5. Out of Scope — 详细说明

以下项目明确不在第一阶段范围内：

- **MQTT / WebSocket**：设备通信通过 Simulator 模拟，不需要实时推送
- **真实设备接入**：MVP 阶段全部使用 Device Simulator（ADR-002 的 Transport 层用 SimulatorAdapter 实现）
- **AI Agent**：架构已预留（ADR-004），但 MVP 不实现 Agent、Tool、RAG
- **批量操作**：ADR-005 已定义 Operation → 多个 Execution 的模型，但 MVP 只验证单设备操作
- **认证/授权**：MVP 使用简化的操作者标识（X-Operator-ID），不实现 JWT/RBAC。这不构成安全边界。
- **设备分组/区域**：MVP 中设备不分组，操作只针对单台设备
- **Device Capability Override**：MVP 中 Device 只能继承 DeviceType 的 Capability，不支持单设备级别的覆盖
- **心跳/轮询/状态抖动抑制**：MVP 采用简化的 OFFLINE 规则（单次 QueryStatus 失败即 OFFLINE）

---

## 6. Domain Concepts

### 6.1 Device Type（设备类型）

**业务语义**：一种设备的分类定义，描述这类设备"是什么"以及"能做什么"。

Device Type 是 Device 的模板。每种 Device Type 定义了一组标准 Capability。
例如："IP 网络功放"是一个 Device Type，它支持 SetVolume、QueryStatus、Restart 三种 Capability。

**关键属性**：
- 名称（如 "IP网络功放"）
- 厂商（如 "DSPPA"）
- 型号（如 "MP2806"）
- 描述
- 支持的 Capability 列表

**业务规则**：
- Device Type 名称 + 厂商 + 型号组合唯一
- Device Type 可以新增 Capability，但不能移除已被 Device 使用的 Capability
- 删除 Device Type 前必须确认没有关联的 Device
- 如果 DeviceType 被删除，其历史 Operation / Execution / Audit 数据仍保持可查询

### 6.2 Device（设备）

**业务语义**：一台具体的、可标识的物理设备实例。

Device 是 Device Type 的实例化。每台 Device 属于某个 Device Type，拥有唯一的设备标识。

**关键属性**：
- 设备 ID（系统生成，全局唯一）
- 名称（运维人员可识别的名称）
- 所属 Device Type
- 连接地址（Simulator 使用 localhost + 端口）
- 注册时间
- 当前状态

**业务规则**：
- 每台 Device 必须属于一个 Device Type
- Device 名称不要求全局唯一（可以有同名设备），但 Device ID 唯一
- Device 创建后进入 REGISTERED 状态
- Device 可以被禁用（DISABLED），禁用后不可执行操作
- **Device 不提供物理删除**；MVP 仅通过 DISABLED 状态实现逻辑停用
- 同一 Device 同时最多允许一个写操作（SetVolume、Restart）处于 RUNNING 或 RETRYING 状态

### 6.3 Capability（设备能力）

**业务语义**：设备能够执行的一种标准操作类型。

Capability 是 Device Gateway 定义的标准化能力接口。它定义了"业务层需要什么"，而不是"协议层怎么做"。

> **已冻结的 MVP 决策**：
> - Capability 由 **DeviceType 定义**
> - Device **继承** DeviceType 的 Capability
> - Device **不能**单独修改或覆盖 Capability
> - 不存在两套 Capability 来源（只有 DeviceType 一个来源）
> - 未来如果需要单设备能力差异，通过独立需求扩展

**MVP Capability 列表**：

| Capability | 语义 | 读/写 | 高风险 |
|------------|------|-------|--------|
| QueryStatus | 查询设备当前状态 | 读 | 否 |
| SetVolume | 设置设备音量（0-100） | 写 | 否 |
| Restart | 重启设备 | 写 | **是** |

**业务规则**：
- Capability 由 DeviceType 定义，Device 从其 DeviceType 继承 Capability
- DeviceType Capability 是 Device 的**唯一** Capability 来源
- 操作前必须验证目标 Device 的 DeviceType 是否定义了所需 Capability
- 不支持的 Capability 操作应被拒绝（返回错误），而不是静默忽略

### 6.4 Status（设备状态）

**业务语义**：设备当前的运行状态信息。

Status 包含两层含义：
1. **连接状态**：设备是否可达（ACTIVE / OFFLINE / REGISTERED / DISABLED）
2. **运行状态**：设备的业务指标（如当前音量、运行时间等）

**MVP 阶段的 Status 模型**：
- 连接状态：REGISTERED / ACTIVE / OFFLINE / DISABLED
- 最后在线时间
- 当前音量（如果设备支持 SetVolume）

**业务规则**：
- Status 通过 QueryStatus 操作获取（主动查询模式）
- **单次 QueryStatus 明确失败或超时 → 设备状态变为 OFFLINE**
- **单次 QueryStatus 成功 → 设备状态变为 ACTIVE**
- MVP 不引入连续失败次数、心跳阈值、滑动窗口或状态抖动抑制
- 设备 OFFLINE 时不能执行写操作（SetVolume、Restart）
- UNKNOWN 不作为连接状态使用

### 6.5 Operation（业务操作）

**业务语义**：用户或系统发起的一次业务意图。

Operation 代表"用户想做什么"，而不是"对每台设备具体做了什么"。

> 遵循 ADR-005：Operation 是高层业务意图的表达。

**MVP 执行模型**：
- MVP 第一版允许 POST /api/operations 同步等待 Simulator 执行完成并返回 Operation 结果
- 但 **Operation / Execution 领域模型保持独立于 HTTP 请求生命周期**
- Operation 保留完整状态机（PENDING → IN_PROGRESS → COMPLETED/FAILED/PARTIALLY_FAILED）
- 为未来异步执行和批量操作保留扩展能力

**MVP 阶段特征**：
- 一个 Operation 对应一台设备（MVP 不做批量）
- Operation 创建后自动产生一个 Execution
- Operation 状态由 Execution 状态聚合

**关键属性**：
- operation_id
- type（QUERY_STATUS / SET_VOLUME / RESTART）
- initiator（操作者标识，来自 X-Operator-ID）
- device_id
- parameters（如 volume=70）
- status
- created_at / completed_at

### 6.6 Execution（设备执行）

**业务语义**：对一台设备的一次具体执行尝试。

> 遵循 ADR-005：Execution 是一个可独立追踪、重试和记录结果的最小设备执行单元。

**关键属性**：
- execution_id
- operation_id
- device_id
- capability
- status（PENDING / RUNNING / SUCCESS / FAILED / RETRYING / SKIPPED）
- retry_count
- result（执行结果数据或错误原因）
- started_at / completed_at

### 6.7 Audit（安全审计）

**业务语义**：不可变的安全审计事件，独立于 Operation / Execution 的业务执行模型。

> **已冻结的 MVP 决策 — Audit 与 Operation/Execution 职责分离**：
>
> | 概念 | 职责 | 记录内容 |
> |------|------|----------|
> | Operation / Execution | 业务执行事实 | 操作生命周期、重试、结果 |
> | Audit | 安全审计事件 | actor、time、target、action、parameters、result、trace_id |
>
> Audit 不是 Operation/Execution 的副本，而是独立的安全事件记录。

**业务规则**：
- 设备注册、禁用、操作执行均产生 Audit 记录
- Audit 记录一旦写入不可修改（append-only）
- Audit 记录包含：操作者（actor）、时间、操作类型、目标设备、参数、结果、trace_id
- QueryStatus 可产生 Audit 记录，但 QueryStatus 不属于状态变更型写操作

---

## 7. User Scenarios

### Scenario A：运维工程师注册新设备

| 项目 | 内容 |
|------|------|
| **Actor** | 运维工程师 |
| **前置条件** | 系统中已存在至少一个 Device Type（如 "IP网络功放"） |
| **User Action** | 1. 进入设备管理页面 2. 点击"注册设备" 3. 选择 Device Type 4. 填写设备名称和连接地址 5. 提交 |
| **System Behavior** | 1. 验证 Device Type 存在 2. 验证连接地址格式 3. 创建设备记录（状态 = REGISTERED） 4. 创建 Audit 记录 5. 返回设备详情 |
| **Device Behavior** | 无（注册阶段不与设备通信） |
| **Success Criteria** | 设备出现在设备列表中，状态为 REGISTERED |
| **Failure Cases** | Device Type 不存在 → 拒绝注册；连接地址为空 → 验证失败 |
| **Audit Requirements** | Audit：operator 注册了 device_id（非 Operation/Execution） |

### Scenario B：运维工程师查询设备状态

| 项目 | 内容 |
|------|------|
| **Actor** | 运维工程师 |
| **前置条件** | 设备已注册且状态不为 DISABLED |
| **User Action** | 1. 进入设备列表 2. 点击某台设备 3. 查看设备详情页 4. 点击"刷新状态" |
| **System Behavior** | 1. 创建 QueryStatus Operation 2. 生成 Execution 3. 通过 Device Gateway 向 Simulator 发送状态查询 4. 更新设备 Status（成功 → ACTIVE，失败 → OFFLINE） 5. 可选产生 Audit 记录 6. 返回最新状态 |
| **Device Behavior** | Simulator 返回预设的状态数据（在线、音量值等） |
| **Success Criteria** | 页面显示设备当前在线状态和音量信息；Operation 状态 = COMPLETED；Execution 状态 = SUCCESS |
| **Failure Cases** | 设备离线 → Execution FAILED，设备状态变为 OFFLINE；Simulator 超时 → Execution FAILED，设备状态变为 OFFLINE |
| **Audit Requirements** | Audit（可选）：operator 查询了 device_id 的状态 |

### Scenario C：运维工程师执行设备操作（SetVolume）

| 项目 | 内容 |
|------|------|
| **Actor** | 运维工程师 |
| **前置条件** | 设备已注册、ACTIVE、且 DeviceType 支持 SetVolume；该设备当前无写操作在 RUNNING/RETRYING |
| **User Action** | 1. 进入设备详情 2. 选择"设置音量" 3. 输入音量值（如 70） 4. 提交 |
| **System Behavior** | 1. 验证设备存在且 ACTIVE 2. 验证 DeviceType 支持 SetVolume 3. 验证音量值范围（0-100） 4. 验证该设备无正在执行的写操作 5. 创建 SetVolume Operation + Execution 6. 创建 Audit 记录 7. 通过 Device Gateway 发送 SetVolume 命令到 Simulator 8. 记录 Execution 结果 9. 更新 Operation 状态 |
| **Device Behavior** | Simulator 接收 SetVolume(70) 命令，更新内部音量状态，返回成功 |
| **Success Criteria** | Operation 状态 = COMPLETED；Execution 状态 = SUCCESS；设备音量更新为 70 |
| **Failure Cases** | 设备 OFFLINE → Operation FAILED；音量值超出范围 → 400 验证失败，不创建 Operation；设备有写操作进行中 → 409 DEVICE_BUSY；Simulator 返回错误 → Execution FAILED |
| **Audit Requirements** | Audit：operator 对 device_id 执行了 SetVolume(volume=70)，结果 SUCCESS |

### Scenario D：运维工程师执行高风险操作（Restart）

| 项目 | 内容 |
|------|------|
| **Actor** | 运维工程师 |
| **前置条件** | 设备已注册、ACTIVE、且 DeviceType 支持 Restart；该设备当前无写操作在 RUNNING/RETRYING |
| **User Action** | 1. 进入设备详情 2. 点击"重启设备" 3. **系统弹出确认对话框** 4. 用户确认 |
| **System Behavior** | 1. 验证设备存在且 ACTIVE 2. 验证 DeviceType 支持 Restart 3. 验证该设备无正在执行的写操作 4. 创建 Restart Operation + Execution 5. 创建 Audit 记录 6. 通过 Device Gateway 发送 Restart 命令 7. Simulator 模拟重启（短暂 OFFLINE → ONLINE） 8. **验证设备恢复 ONLINE**（通过内部 QueryStatus 确认） 9. 记录 Execution 结果 10. 更新 Operation 状态 |
| **Device Behavior** | Simulator 模拟：收到 Restart → 标记 OFFLINE → 延迟 N 秒 → 恢复 ONLINE |
| **Success Criteria** | Operation 状态 = COMPLETED；Execution 状态 = SUCCESS；**设备确认恢复 ONLINE 状态**（不仅仅是命令发送成功） |
| **Failure Cases** | 用户取消确认 → 不创建 Operation；Simulator 重启后未恢复 ONLINE → Execution FAILED；设备有写操作进行中 → 409 DEVICE_BUSY |
| **Audit Requirements** | Audit：operator 确认并执行了 Restart，结果 SUCCESS/FAILED |

> **已冻结的 MVP 决策 — Restart SUCCESS 定义**：
>
> Restart 的 SUCCESS 不能仅表示"命令发送成功"。MVP 的业务成功条件为：
> 1. Restart command successfully accepted
> 2. Device enters restart state (OFFLINE)
> 3. Device becomes ONLINE again
> 4. Status verification succeeds（内部 QueryStatus 确认设备恢复）
>
> 只有完成上述全部验证，Execution 才能标记 SUCCESS。
> 如果命令发送成功但设备没有恢复 ONLINE：Execution = FAILED。

### Scenario E：操作失败与查看结果

| 项目 | 内容 |
|------|------|
| **Actor** | 运维工程师 |
| **前置条件** | 之前有操作失败（如 Simulator 配置为返回错误） |
| **User Action** | 1. 在设备详情页查看操作历史 2. 点击失败的 Operation 3. 查看 Execution 详情和错误信息 |
| **System Behavior** | 1. 展示 Operation 列表（包含状态标记） 2. 展示选中 Operation 的 Execution 详情 3. 显示错误原因 |
| **Device Behavior** | 无 |
| **Success Criteria** | 用户可以看到失败操作的完整信息：谁、何时、做了什么、结果如何、为什么失败 |
| **Failure Cases** | N/A |
| **Audit Requirements** | 查看操作历史不产生新 Audit 记录 |

---

## 8. Business Rules

### 设备管理规则

| 编号 | 规则 |
|------|------|
| BR-001 | Device Type 的（名称 + 厂商 + 型号）组合必须全局唯一 |
| BR-002 | Device 必须属于一个且仅一个 Device Type |
| BR-003 | Device 创建时状态为 REGISTERED |
| BR-004 | DISABLED 状态的 Device 不允许执行任何操作 |
| BR-005 | Device Type 被引用（有 Device 关联）时不允许删除；无关联 Device 时允许删除，历史 Operation / Execution / Audit 数据保持可查询 |
| BR-006 | Device 不提供物理删除；MVP 仅通过 DISABLED 状态实现逻辑停用 |
| BR-007 | Device 的 Capability 完全由其 DeviceType 定义，Device 不能单独修改或覆盖 Capability |

### 操作规则

| 编号 | 规则 |
|------|------|
| BR-010 | 执行操作前必须验证 Device 的 DeviceType 支持目标 Capability |
| BR-011 | 写操作（SetVolume、Restart）在 Device 不为 ACTIVE 时拒绝执行 |
| BR-012 | SetVolume 的参数必须在 0-100 范围内 |
| BR-013 | Restart 是高风险操作，必须经过用户二次确认 |
| BR-014 | 所有操作（包括 QueryStatus）创建 Operation + Execution 记录 |
| BR-015 | Operation 状态由 Execution 状态聚合（遵循 ADR-005） |
| BR-016 | **同一 Device 同时最多允许一个写操作处于 RUNNING 或 RETRYING 状态；如果已有写操作执行中，新的写操作返回 409 DEVICE_BUSY** |

### 审计规则

| 编号 | 规则 |
|------|------|
| BR-020 | 设备注册、禁用产生 Audit 记录（独立于 Operation/Execution） |
| BR-021 | Audit 记录一旦写入不可修改（append-only） |
| BR-022 | Audit 记录包含：actor、time、action、target device、parameters、result、trace_id |
| BR-023 | 操作执行（SetVolume、Restart）产生 Audit 记录 |
| BR-024 | QueryStatus 可产生 Audit 记录，但 QueryStatus 不属于状态变更型写操作 |
| BR-025 | **Audit 与 Operation/Execution 是独立概念**：Operation/Execution 记录业务执行事实；Audit 记录安全审计事件 |

### 重试规则

| 编号 | 规则 |
|------|------|
| BR-030 | 重试粒度在 Execution 级别，不在 Operation 级别（遵循 ADR-005） |
| BR-031 | MVP 阶段最大重试次数 = 3 |
| BR-032 | 重试不重复业务判断（权限、Capability 验证已在 Operation 创建时完成） |
| BR-033 | QueryStatus 不重试（查询类操作无需重试） |

---

## 9. Lifecycle

### 9.1 Device Lifecycle

```text
REGISTERED → ACTIVE → OFFLINE
    |              |          |
    |              ↓          ↓
    |           ACTIVE ←←← OFFLINE
    |              |
    ↓              ↓
  DISABLED     DISABLED
```

| 状态 | 含义 | 可执行操作 |
|------|------|-----------|
| REGISTERED | 已注册，尚未验证连接 | QueryStatus |
| ACTIVE | 在线且可用 | 所有操作 |
| OFFLINE | 连接不可达（单次 QueryStatus 失败或超时） | QueryStatus（尝试恢复） |
| DISABLED | 管理员手动禁用 | 无 |

**状态转换触发条件**：
- REGISTERED → ACTIVE：首次 QueryStatus 成功
- ACTIVE → OFFLINE：**单次 QueryStatus 失败或超时**
- OFFLINE → ACTIVE：QueryStatus 恢复成功
- 任意 → DISABLED：管理员手动禁用
- DISABLED → REGISTERED：管理员重新启用

> **已冻结的 MVP 决策 — OFFLINE 规则**：
>
> 单次 QueryStatus 明确失败或超时 → 设备状态变为 OFFLINE。
> 单次 QueryStatus 成功 → 设备状态变为 ACTIVE。
>
> MVP 不引入：连续失败次数、心跳阈值、滑动窗口、状态抖动抑制。
> 这些能力留到后续版本。

### 9.2 Operation Lifecycle

遵循 ADR-005 定义：

```text
PENDING → IN_PROGRESS → COMPLETED
                       → PARTIALLY_FAILED
                       → FAILED
```

| 状态 | 含义 |
|------|------|
| PENDING | Operation 已创建，尚未开始执行 |
| IN_PROGRESS | 至少一个 Execution 开始执行 |
| COMPLETED | 所有 Execution 成功 |
| PARTIALLY_FAILED | 部分 Execution 成功，部分失败（MVP 不会出现，因为单设备操作） |
| FAILED | 所有 Execution 失败 |

> MVP 阶段：一个 Operation 对应一个 Execution，因此不会出现 PARTIALLY_FAILED。但系统必须支持该状态，为后续批量操作预留。

> **MVP 执行模型说明**：
> POST /api/operations 同步等待 Simulator 执行完成并返回 Operation 结果。
> 但 Operation / Execution 领域模型保持独立于 HTTP 请求生命周期。
> 不要把 Operation 设计成简单的 HTTP Request Record。

### 9.3 Execution Lifecycle

遵循 ADR-005 定义：

```text
PENDING → RUNNING → SUCCESS
                   → FAILED → RETRYING → RUNNING → SUCCESS / FAILED
                   → SKIPPED
```

| 状态 | 含义 |
|------|------|
| PENDING | Execution 已创建，等待执行 |
| RUNNING | 正在执行设备操作 |
| SUCCESS | 设备操作成功（含业务验证通过） |
| FAILED | 设备操作失败（且无更多重试机会） |
| RETRYING | 失败后正在重试 |
| SKIPPED | 跳过（如 Capability 不满足） |

---

## 10. Capability Definition

### 10.1 QueryStatus

| 项目 | 定义 |
|------|------|
| **业务语义** | 获取设备当前的运行状态信息 |
| **输入** | device_id |
| **输出** | 连接状态（ONLINE/OFFLINE）、最后在线时间、当前音量（如支持） |
| **前置条件** | Device 存在且不为 DISABLED |
| **失败条件** | 设备不可达、Simulator 超时 |
| **失败后果** | Execution = FAILED；设备状态变为 OFFLINE |
| **成功后果** | Execution = SUCCESS；设备状态变为 ACTIVE |
| **是否可重试** | 否（查询类操作，失败即返回） |
| **是否需要确认** | 否 |
| **风险等级** | 低 |
| **Operation/Execution** | 创建 Operation + Execution（BR-014） |
| **Audit** | 可选产生 Audit 记录（BR-024） |

### 10.2 SetVolume

| 项目 | 定义 |
|------|------|
| **业务语义** | 设置设备的输出音量 |
| **输入** | device_id, volume (0-100, 整数) |
| **输出** | 执行结果（成功/失败）、实际设置的音量值 |
| **前置条件** | Device 存在且 ACTIVE、DeviceType 支持 SetVolume、该设备无正在执行的写操作 |
| **失败条件** | 设备离线、参数超出范围、设备返回错误、设备繁忙 |
| **是否可重试** | 是（最多 3 次） |
| **是否需要确认** | 否（低风险操作） |
| **风险等级** | 低 |

### 10.3 Restart

| 项目 | 定义 |
|------|------|
| **业务语义** | 重启设备，使设备重新初始化，并验证设备恢复在线 |
| **输入** | device_id |
| **输出** | 执行结果（成功/失败）、设备重启后的状态 |
| **前置条件** | Device 存在且 ACTIVE、DeviceType 支持 Restart、该设备无正在执行的写操作 |
| **失败条件** | 设备离线、设备重启超时、**设备重启后未能恢复 ONLINE**、设备繁忙 |
| **是否可重试** | 是（最多 1 次，因为重复重启风险高） |
| **是否需要确认** | **是**（高风险操作） |
| **风险等级** | **高** |

> **Restart SUCCESS 业务定义**：
>
> Restart 的 SUCCESS 包含以下完整验证链：
> 1. Restart command successfully accepted by device
> 2. Device enters restart state (OFFLINE)
> 3. Device becomes ONLINE again
> 4. Status verification succeeds (内部 QueryStatus 确认)
>
> 如果命令发送成功但设备没有恢复 ONLINE → Execution = FAILED。
>
> **对 Simulator 的要求**：Simulator 必须模拟完整重启过程——
> 收到 Restart → 标记 OFFLINE → 延迟 N 秒 → 恢复 ONLINE →
> 随后的 QueryStatus 返回 ONLINE 状态。

---

## 11. Simulator Requirements

### 11.1 Simulator 是什么

Device Simulator 是一个测试基础设施，模拟真实广播设备的行为。
它通过 SimulatorAdapter 实现 Device Gateway 定义的接口，使系统可以在没有真实设备的情况下运行和测试。

> 遵循 ARCHITECTURE.md 第12节 + ADR-002：SimulatorAdapter 实现 Device Gateway interface。

### 11.2 Simulator 模拟什么

| 模拟项 | 说明 |
|--------|------|
| 连接状态 | 可控制 ONLINE / OFFLINE |
| 状态查询 | 返回预设的设备状态（音量、运行时间等） |
| SetVolume | 接收并记录音量值，返回成功 |
| Restart | 模拟完整重启过程（短暂 OFFLINE → 延迟 → 恢复 ONLINE） |
| 错误响应 | 可配置为返回特定错误 |
| 超时 | 可配置延迟响应以模拟超时 |

### 11.3 Simulator 不模拟什么

| 不模拟项 | 原因 |
|----------|------|
| 真实协议解析 | Simulator 不实现任何真实设备协议 |
| 网络传输细节 | 不需要真实的 TCP/UDP 通信 |
| 硬件行为 | 不模拟真实的硬件启动/关闭过程 |
| 多设备并发 | MVP 阶段不测试并发场景 |

### 11.4 Simulator 行为控制

Simulator 必须支持以下配置：

| 控制项 | 说明 |
|--------|------|
| mode | normal / error / timeout / offline |
| delay_ms | 响应延迟（用于模拟慢设备或超时） |
| volume | 当前音量值 |
| restart_delay_ms | 重启延迟 |

**控制方式**：
- 每台 Simulator 实例维护自己的状态
- 测试代码可以直接设置 Simulator 的行为模式
- MVP 阶段 Simulator 作为进程内组件运行（不需要独立进程）

### 11.5 Simulator 与 Gateway / Adapter 的关系

```text
Application Service
    ↓
Device Gateway (interface)
    ↓
SimulatorAdapter (实现 Gateway interface)
    ↓
Simulator (内存状态机)
```

- **SimulatorAdapter 实现 Device Gateway 定义的 interface**（不是"Simulator implements Gateway"）
- Simulator 是 SimulatorAdapter 的内部状态机
- SimulatorAdapter 负责协议适配，Simulator 负责设备行为模拟
- 未来新增真实 Adapter 时，只需实现相同的 Gateway interface

> **架构层次**：
> - Device Gateway = 业务能力接口（interface）
> - SimulatorAdapter = Gateway 的 Adapter 实现（concrete implementation）
> - Simulator = 内存状态机 / 测试设备（internal to Adapter）

### 11.6 验证要求

| 验证项 | 方法 |
|--------|------|
| Gateway 接口满足 | Adapter Contract Test（ADR-006） |
| Simulator 行为正确 | Unit Test + Integration Test |
| Operation/Execution 记录正确 | Integration Test |
| Audit 记录正确 | Integration Test |
| 错误处理正确 | 配置 Simulator 为 error 模式，验证 Operation 状态 |
| 超时处理正确 | 配置 Simulator 为 timeout 模式，验证 Execution 状态 |
| Restart 完整流程 | 配置 Simulator normal 模式，验证 Restart → OFFLINE → ONLINE 完整链路 |

---

## 12. Frontend Requirements

### 12.1 页面列表

| 页面 | 功能 | 数据来源 |
|------|------|----------|
| Dashboard | 系统概览（已有，增加设备统计） | GET /health + GET /api/devices/stats |
| Device List | 设备列表（分页、搜索） | GET /api/devices |
| Device Detail | 设备详情 + 状态 + 操作入口 | GET /api/devices/:id |
| Device Operation | 操作执行表单 | POST /api/operations |
| Operation Result | 操作结果查看 | GET /api/operations/:id |
| Device Type Management | 设备类型管理 | CRUD /api/device-types |

### 12.2 页面交互要求

**Device List**：
- 显示所有设备（名称、类型、状态、最后在线）
- 支持按名称搜索
- 状态用颜色标识（ACTIVE=绿、OFFLINE=红、DISABLED=灰、REGISTERED=蓝）

**Device Detail**：
- 显示设备基本信息
- 显示当前状态（支持手动刷新）
- 显示设备继承自 DeviceType 的 Capability 列表
- 提供操作入口（根据 DeviceType Capability 动态展示）
- 显示最近操作历史

**Device Operation**：
- SetVolume：输入框（0-100），提交按钮
- Restart：按钮 + 确认对话框
- 操作后跳转到 Operation Result
- 如果设备有写操作正在执行，显示 DEVICE_BUSY 提示

**Operation Result**：
- 显示 Operation 状态
- 显示 Execution 详情
- 失败时显示错误原因
- 显示 trace_id

### 12.3 前端约束

- 所有数据通过 API Client 从 Backend 获取
- 禁止 mock 业务数据
- 禁止前端维护业务状态（状态由 Backend 管理）
- 禁止前端直接访问数据库或设备
- 高风险操作（Restart）必须有确认对话框
- 操作结果页面展示 trace_id（从 API 响应中获取）

---

## 13. API Requirements

### 13.1 Device Type APIs

| Method | Path | 说明 |
|--------|------|------|
| GET | /api/device-types | 获取设备类型列表 |
| POST | /api/device-types | 创建设备类型 |
| GET | /api/device-types/:id | 获取设备类型详情 |
| PUT | /api/device-types/:id | 更新设备类型 |
| DELETE | /api/device-types/:id | 删除设备类型（仅当无关联 Device） |

**POST /api/device-types 请求体示例**：

```json
{
  "name": "IP网络功放",
  "vendor": "DSPPA",
  "model": "MP2806",
  "description": "支持网络音频播放和远程控制",
  "capabilities": ["QueryStatus", "SetVolume", "Restart"]
}
```

**响应**（统一格式）：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "dt_001",
    "name": "IP网络功放",
    "vendor": "DSPPA",
    "model": "MP2806",
    "capabilities": ["QueryStatus", "SetVolume", "Restart"],
    "created_at": "2026-09-23T10:00:00Z"
  }
}
```

**错误场景**：
- 400：参数验证失败
- 409：Device Type 已存在（名称+厂商+型号重复）

### 13.2 Device APIs

| Method | Path | 说明 |
|--------|------|------|
| GET | /api/devices | 获取设备列表（支持分页、搜索） |
| POST | /api/devices | 注册设备 |
| GET | /api/devices/:id | 获取设备详情 |
| PUT | /api/devices/:id | 更新设备信息 |
| PATCH | /api/devices/:id/status | 启用/禁用设备 |

**GET /api/devices 查询参数**：
- page (int, default 1)
- page_size (int, default 20)
- search (string, 按名称搜索)

**POST /api/devices 请求体示例**：

```json
{
  "name": "三楼广播主机",
  "device_type_id": "dt_001",
  "address": "simulator://localhost:9001"
}
```

**错误场景**：
- 400：参数验证失败
- 404：Device Type 不存在

### 13.3 Operation APIs

| Method | Path | 说明 |
|--------|------|------|
| POST | /api/operations | 执行设备操作（同步等待结果） |
| GET | /api/operations/:id | 获取 Operation 详情（含 Execution） |
| GET | /api/devices/:id/operations | 获取设备的操作历史 |

**POST /api/operations 请求体示例**：

```json
{
  "device_id": "dev_001",
  "type": "SET_VOLUME",
  "parameters": {
    "volume": 70
  }
}
```

**响应示例**（成功）：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "operation_id": "op_001",
    "type": "SET_VOLUME",
    "device_id": "dev_001",
    "status": "COMPLETED",
    "executions": [
      {
        "execution_id": "ex_001",
        "status": "SUCCESS",
        "result": {"volume": 70},
        "started_at": "2026-09-23T10:00:01Z",
        "completed_at": "2026-09-23T10:00:02Z"
      }
    ],
    "created_at": "2026-09-23T10:00:00Z",
    "completed_at": "2026-09-23T10:00:02Z",
    "trace_id": "trace_xxx"
  }
}
```

**错误场景**：
- 400：参数验证失败（如 volume 超出范围）
- 403：设备 DISABLED 或不为 ACTIVE
- 404：设备不存在
- 409 DEVICE_BUSY：**该设备已有写操作正在执行（RUNNING 或 RETRYING 状态）**
- 409：设备 DeviceType 不支持目标 Capability

**409 DEVICE_BUSY 响应示例**：

```json
{
  "code": "DEVICE_BUSY",
  "message": "该设备已有操作正在执行",
  "data": {
    "active_operation_id": "op_existing"
  }
}
```

### 13.4 Permission Requirements

> **已冻结的 MVP 决策 — Authentication / Authorization 边界**：
>
> X-Operator-ID 仅作为 MVP 操作者标识：
> - 用于 Operation initiator
> - 用于 Audit actor
> - **不代表 Authentication**
> - **不代表 Authorization**
> - **不构成安全边界**
>
> MVP 所有 API 当前为简化权限模型，所有操作者具有相同权限。
> 完整的认证/授权系统将通过独立 ADR 在后续阶段定义。

| 项目 | 说明 |
|------|------|
| 请求头 | X-Operator-ID（必填，标识操作者） |
| MVP 权限 | 无（所有操作者权限相同） |
| 后续扩展 | JWT / RBAC（通过独立 ADR） |

---

## 14. Acceptance Criteria

### 设备类型管理

| 编号 | 验收标准 |
|------|----------|
| AC-001 | 给定有效的设备类型信息，当用户创建时，则系统返回成功的设备类型记录 |
| AC-002 | 给定已存在的（名称+厂商+型号），当用户重复创建时，则系统返回 409 错误 |
| AC-003 | 给定一个被 Device 引用的 Device Type，当用户尝试删除时，则系统拒绝并返回错误 |
| AC-004 | 给定一个未被引用的 Device Type，当用户删除时，则系统成功删除且历史 Operation/Execution/Audit 数据保持可查询 |

### 设备管理

| 编号 | 验收标准 |
|------|----------|
| AC-010 | 给定有效的设备信息和存在的 Device Type，当用户注册设备时，则设备状态为 REGISTERED |
| AC-011 | 给定不存在的 Device Type ID，当用户注册设备时，则系统返回 404 错误 |
| AC-012 | 给定设备列表，当用户访问时，则返回所有设备及其当前状态 |
| AC-013 | 给定一台 DISABLED 设备，当用户尝试执行操作时，则系统拒绝并返回 403 |
| AC-014 | 给定一台 Device，当查看其 Capability 时，则显示其 DeviceType 定义的全部 Capability（即继承，无覆盖） |

### 设备操作

| 编号 | 验收标准 |
|------|----------|
| AC-020 | 给定 ACTIVE 设备且 DeviceType 支持 QueryStatus，当执行查询时，则返回设备状态且 Operation 状态为 COMPLETED |
| AC-021 | 给定 ACTIVE 设备且 DeviceType 支持 SetVolume，当执行 SetVolume(70) 时，则 Execution SUCCESS 且设备音量更新为 70 |
| AC-022 | 给定 SetVolume 参数 volume=150（超出范围），当提交时，则系统返回 400 验证错误，不创建 Operation |
| AC-023 | 给定 OFFLINE 设备，当尝试 SetVolume 时，则系统拒绝执行（设备不满足 ACTIVE 前提条件） |
| AC-024 | 给定设备的 DeviceType 不支持 Restart，当尝试 Restart 时，则系统拒绝并返回错误 |
| AC-025 | 给定操作执行完成，当查看 Operation 详情时，则可以看到完整的 Execution 信息和 trace_id |
| AC-026 | 给定 Restart 操作执行完成，当 Simulator 重启后设备恢复 ONLINE 时，则 Execution 状态为 SUCCESS；如果设备未恢复 ONLINE，则 Execution 状态为 FAILED |
| AC-027 | 给定一台设备已有写操作正在执行（RUNNING 或 RETRYING），当提交新的写操作时，则系统返回 409 DEVICE_BUSY |

### 状态与生命周期

| 编号 | 验收标准 |
|------|----------|
| AC-030 | 给定 REGISTERED 设备，当首次 QueryStatus 成功时，则设备状态变为 ACTIVE |
| AC-031 | 给定 ACTIVE 设备，当 QueryStatus 失败时，则设备状态变为 OFFLINE（单次失败即触发） |
| AC-032 | 给定 OFFLINE 设备，当 QueryStatus 恢复成功时，则设备状态变为 ACTIVE |

### 审计

| 编号 | 验收标准 |
|------|----------|
| AC-040 | 给定一次 SetVolume 操作，当操作完成时，则可以查到对应的 Operation、Execution 和 Audit 记录 |
| AC-041 | 给定一次失败的 Restart 操作，当查看审计时，则可以看到失败原因 |
| AC-042 | 审计记录不可修改（尝试修改应被系统拒绝或不存在修改接口） |
| AC-043 | 给定一次操作执行，当查看数据时，则 Audit 记录与 Operation/Execution 记录是独立的（Audit 包含 actor/action/trace_id，Operation/Execution 包含业务执行事实） |

### Capability

| 编号 | 验收标准 |
|------|----------|
| AC-044 | 给定一台 Device，当查询其 Capability 时，则结果完全等于其 DeviceType 定义的 Capability 集合（DeviceType Capability 是 Device 的唯一 Capability 来源） |

### 权限

| 编号 | 验收标准 |
|------|----------|
| AC-028 | 给定一个 API 请求带有 X-Operator-ID，当请求成功时，则 Operation 的 initiator 和 Audit 的 actor 记录为该值；X-Operator-ID 不作为安全授权机制 |

### Simulator

| 编号 | 验收标准 |
|------|----------|
| AC-050 | 给定 Simulator 设置为 normal 模式，当执行 SetVolume 时，则返回 SUCCESS |
| AC-051 | 给定 Simulator 设置为 error 模式，当执行 SetVolume 时，则返回 FAILED |
| AC-052 | 给定 Simulator 设置为 timeout 模式，当执行操作时，则 Execution 超时并标记为 FAILED |
| AC-053 | 给定 Simulator 设置为 offline 模式，当 QueryStatus 时，则设备状态标记为 OFFLINE |
| AC-029 | 给定 SimulatorAdapter 实现，当运行 Gateway Contract Test 时，则 SimulatorAdapter 通过所有测试用例（ADR-006） |

---

## 15. Open Questions

> 本版本已冻结的决策不再列入 Open Questions。
> 已解决的原有问题：
> - OQ-002（Device 删除策略）→ 已冻结为 DISABLED 逻辑停用
> - OQ-004（OFFLINE 定义）→ 已冻结为单次 QueryStatus 失败即 OFFLINE
> - OQ-005（Status 查询模式）→ 已冻结为主动查询
> - OQ-011（QueryStatus 审计）→ 已冻结为可选产生 Audit
> - OQ-015（设备并发操作）→ 已冻结为单设备单写操作（BR-016）

### A. Product Decision（需要在 Design 前解决的产品决策）

| 编号 | 问题 | 影响 |
|------|------|------|
| OQ-003 | Device Type 是否允许修改 Capability 列表？如允许，已关联 Device 如何处理？ | 设备类型管理规则 |
| OQ-006 | Operation 是否允许取消？（如正在执行中的 Operation） | 状态机 |

### B. Technical Design Decision（Design 阶段解决的技术决策）

| 编号 | 问题 | 影响 |
|------|------|------|
| OQ-001 | Device ID 如何生成？（UUID / 自增 / 带前缀） | 数据模型设计 |
| OQ-007 | Retry 最大次数是全局配置还是按 Capability 配置？ | 重试策略 |
| OQ-008 | Restart 重试次数是否应与 SetVolume 不同？（高风险操作少重试） | 重试策略 |
| OQ-009 | Simulator 的连接地址格式如何定义？（simulator://localhost:9001?） | 设备注册 |
| OQ-010 | 设备列表分页的默认排序是什么？（按注册时间？按名称？） | API 设计 |
| OQ-012 | MVP 是否需要操作历史的分页？ | API 设计 |
| OQ-014 | Device 的 address 字段是否应该支持多种格式？ | 数据模型 |

### C. Future Decision（MVP 之后的决策）

| 编号 | 问题 | 影响 |
|------|------|------|
| OQ-013 | 未来 Agent 执行操作时，initiator 如何标识？ | 审计模型 |
| OQ-016 | 未来是否需要 Device Capability Override？ | 能力模型扩展 |
| OQ-017 | 未来 OFFLINE 判定是否需要心跳/滑动窗口/状态抖动抑制？ | 状态机演进 |

---

## 16. Future Extension

以下能力不在 MVP 范围内，但架构已为其预留扩展点：

| 扩展项 | 架构支撑 | 引入时机 |
|--------|----------|----------|
| 批量操作 | ADR-005 Operation → 多个 Execution | 第二个迭代 |
| 设备分组/区域 | internal/modules/ 新增模块 | 需求明确后 |
| 真实设备接入 | ADR-002 Adapter + Transport | 协议标准确定后 |
| AI Agent | ADR-004 Tool + Defense in Depth | 业务模块稳定后 |
| 告警中心 | 新增模块 | 监控需求明确后 |
| 用户认证/授权 | platform/http middleware | 独立 ADR |
| 设备固件升级 | 新增 Capability | 需求明确后 |
| WebSocket 实时推送 | platform/http 扩展 | 实时性需求明确后 |
| Device Capability Override | 独立需求评审 | 单设备差异化需求明确后 |
| OFFLINE 智能判定 | 心跳/滑动窗口/抖动抑制 | 稳定性需求明确后 |

---

## Appendix: 领域关系图

```text
DeviceType ──(1:N)──→ Device
    │                    │
    │ defines            │ inherits
    ↓                    ↓
Capability ──→ Device (inherits from DeviceType)
    │
    │ enables
    ↓
Operation ──(1:N)──→ Execution
(business intent)    (device execution)
    │                    │
    │ produces           │ produces
    ↓                    ↓
Audit (independent)  Audit (independent)
```

**关系解释**：

1. **DeviceType → Device (1:N)**：一种设备类型可以有多个设备实例
2. **DeviceType defines Capability**：设备类型定义了这类设备支持的能力集合（唯一来源）
3. **Device inherits Capability**：设备从其 DeviceType 继承所有能力（不能单独修改或覆盖）
4. **Capability enables Operation**：只有 DeviceType 定义了某 Capability，Device 才能执行对应 Operation
5. **Operation → Execution (1:N)**：一个业务意图可以产生多个设备级执行（MVP 阶段为 1:1）
6. **Operation/Execution → Audit（独立）**：Audit 是独立的安全审计事件，不是 Operation/Execution 的副本

---

## Appendix B: 已冻结的产品决策汇总

| 编号 | 决策 | 冻结依据 |
|------|------|----------|
| D-001 | Capability 由 DeviceType 定义，Device 继承，不支持单设备覆盖 | 本文件 §6.3 |
| D-002 | Restart SUCCESS = 命令发送 + 设备恢复 ONLINE + 状态验证通过 | 本文件 §10.3 |
| D-003 | Operation/Execution 领域模型独立于 HTTP 生命周期 | 本文件 §6.5, §9.2 |
| D-004 | Audit 与 Operation/Execution 职责分离（业务事实 vs 安全事件） | 本文件 §6.7, §8 |
| D-005 | OFFLINE = 单次 QueryStatus 失败或超时（无心跳/滑动窗口） | 本文件 §9.1 |
| D-006 | X-Operator-ID 仅为操作者标识，不代表 Authentication/Authorization | 本文件 §13.4 |
| D-007 | 同一 Device 同时最多一个写操作（RUNNING/RETRYING） | 本文件 §8, BR-016 |
| D-008 | Device 不物理删除，仅 DISABLED | 本文件 §6.2, §8 |
| D-009 | DeviceType 删除仅当无关联 Device，历史数据保持可查询 | 本文件 §6.1, §8 |
| D-010 | QueryStatus 创建 Operation + Execution，可选 Audit，非状态变更型写操作 | 本文件 §6.7, §8 |
| D-011 | SimulatorAdapter 实现 Device Gateway（不是 Simulator 直接实现） | 本文件 §11.5 |
