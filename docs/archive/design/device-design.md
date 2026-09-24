# Device Module — Technical Design

**Status**: DRAFT
**Date**: 2026-09-24
**Phase**: Phase 3B — Technical Design
**Input**: device-mvp-requirement.md (v2, frozen)
**Dependencies**: ARCHITECTURE.md, ADR-001, ADR-002, ADR-004, ADR-005, ADR-006

---

## 1. Design Overview

### 1.1 MVP Technical Design Goal

将已冻结的 Device MVP Requirement 转换为可直接进入 API Design 和 Implementation Plan 的 Technical Design。

核心目标：
- 定义 Device Module 的完整内部结构
- 明确 Domain Model 和状态机
- 设计 Device Gateway / SimulatorAdapter 接口
- 定义 Application Service 编排逻辑
- 设计数据库 Schema（概念层）
- 定义 API Boundary（Controller → DTO → Service）
- 明确 Error Model 和 Transaction Boundary
- 设计测试矩阵

### 1.2 Module Boundary

MVP 采用 **单 Module** 设计：internal/modules/device/

所有 Device 相关概念均归属于此 Module：

| 概念 | 归属 | 理由 |
|------|------|------|
| DeviceType | device module | 设备类型定义 |
| Device | device module | 设备实例 |
| Capability | device module | 设备能力（DeviceType 定义） |
| DeviceStatus | device module | 设备状态信息 |
| Operation | device module | 业务操作意图 |
| Execution | device module | 设备执行单元 |
| AuditEvent | device module | 安全审计事件 |

> **设计决策**：不为 Operation/Execution/Audit 单独拆分 Module。
> 原因：MVP 阶段所有概念紧密耦合于设备管理领域，
> 拆分过早会增加模块间通信复杂度。
> 未来如果 Operation/Audit 需要跨模块共享，
> 再通过独立 ADR 拆分。

### 1.3 Core Call Chain

```text
HTTP Request
    ↓
Middleware (Trace ID / Access Log / Recovery / CORS)
    ↓
Controller (interfaces/)
    ├── Request DTO validation
    ├── Extract X-Operator-ID, trace_id
    ↓
Application Service (application/)
    ├── Business rule validation
    ├── Domain model operation
    ├── Device Gateway call (for device operations)
    ├── Persistence (Repository)
    ├── Audit event creation
    ↓
Response DTO
    ↓
HTTP Response (unified format)
```

### 1.4 Device Operation Call Chain

```text
HTTP POST /api/operations
    ↓
OperationController
    ↓
OperationAppService.ExecuteOperation()
    ├── 1. Load Device + DeviceType (Repository)
    ├── 2. Validate: Device exists, not DISABLED
    ├── 3. Validate: DeviceType supports requested Capability
    ├── 4. Validate: Device is ACTIVE (for write operations)
    ├── 5. Validate: No active write operation (DEVICE_BUSY check)
    ├── 6. Create Operation (PENDING) + Execution (PENDING)
    ├── 7. Create AuditEvent
    ├── 8. Persist Operation + Execution + Audit (DB transaction)
    ├── 9. Update Execution → RUNNING
    ├── 10. Call DeviceGateway.Execute(capability, params)
    │       ↓
    │   SimulatorAdapter.Execute()
    │       ↓
    │   Simulator (state machine)
    │       ↓
    │   Return result
    ├── 11. Update Execution → SUCCESS/FAILED
    ├── 12. Update Operation status (aggregate from Execution)
    ├── 13. For Restart: verify device recovered ONLINE
    ├── 14. Update Device status (if QueryStatus)
    ├── 15. Persist updates (DB transaction)
    └── 16. Return Operation + Execution result
```

### 1.5 HTTP API 与 Application Service 的关系

```text
Controller (interfaces/)
    ↓ Request DTO
Application Service (application/)
    ↓ Domain operation
Domain Model (domain/)
    ↓
Repository (infrastructure/)  +  Device Gateway (domain/ interface)
```

- Controller 只做：参数绑定、DTO 转换、调用 Application Service、返回响应
- Controller 禁止包含业务逻辑
- Application Service 编排业务流程，不实现业务规则
- Domain Model 封装业务规则和状态转换
- Repository 负责持久化
- Device Gateway 是 domain 层定义的 interface，infrastructure 层实现

---

## 2. Module Boundary

### 2.1 Module 职责

**Device Module** 负责：
- DeviceType 的 CRUD 和 Capability 定义
- Device 的注册、管理、状态维护
- 设备操作的创建、执行、追踪
- 安全审计事件的记录
- Device Gateway 接口定义和 SimulatorAdapter 实现

**Device Module 不负责**：
- 用户认证/授权（MVP 使用 X-Operator-ID 简化标识）
- 跨模块通信（MVP 单模块）
- Agent Tool 暴露（ADR-004，后续阶段）
- 真实设备协议实现（MVP 使用 Simulator）

### 2.2 Module 对外暴露的 Application Service

| Service | 方法 | 用途 |
|---------|------|------|
| DeviceTypeAppService | CreateDeviceType | 创建设备类型 |
| | UpdateDeviceType | 更新设备类型 |
| | DeleteDeviceType | 删除设备类型（无关联 Device） |
| | GetDeviceType | 获取设备类型详情 |
| | ListDeviceTypes | 设备类型列表 |
| DeviceAppService | RegisterDevice | 注册设备 |
| | UpdateDevice | 更新设备信息 |
| | DisableDevice | 禁用设备 |
| | EnableDevice | 启用设备 |
| | GetDevice | 获取设备详情 |
| | ListDevices | 设备列表（分页、搜索） |
| OperationAppService | ExecuteOperation | 执行设备操作 |
| | GetOperation | 获取 Operation 详情 |
| | ListDeviceOperations | 设备操作历史 |

### 2.3 Module 内部结构

```text
internal/modules/device/
├── domain/
│   ├── entity/
│   │   ├── device_type.go        # DeviceType entity
│   │   ├── device.go             # Device entity
│   │   ├── operation.go          # Operation entity
│   │   ├── execution.go          # Execution entity
│   │   └── audit_event.go        # AuditEvent entity
│   ├── value/
│   │   ├── capability.go         # Capability value object
│   │   ├── device_status.go      # DeviceStatus value object
│   │   └── volume.go             # Volume value object (0-100)
│   ├── errors.go                 # Domain errors
│   └── gateway.go                # Device Gateway interface
├── application/
│   ├── device_type_service.go    # DeviceType application service
│   ├── device_service.go         # Device application service
│   ├── operation_service.go      # Operation application service
│   └── dto/
│       ├── request/              # Request DTOs
│       └── response/             # Response DTOs
├── interfaces/
│   ├── device_type_handler.go    # HTTP handlers
│   ├── device_handler.go
│   ├── operation_handler.go
│   └── router.go                 # Route registration
└── infrastructure/
    ├── repository/
    │   ├── device_type_repo.go   # DeviceType repository impl
    │   ├── device_repo.go        # Device repository impl
    │   ├── operation_repo.go     # Operation repository impl
    │   └── audit_repo.go         # Audit repository impl
    └── adapter/
        ├── simulator_adapter.go  # SimulatorAdapter (implements Gateway)
        └── simulator.go          # Simulator (state machine)
```

### 2.4 依赖方向

```text
interfaces/ → application/ → domain/
infrastructure/ → application/ → domain/
```

- domain/ 不依赖任何外层
- pplication/ 只依赖 domain/
- interfaces/ 依赖 pplication/
- infrastructure/ 依赖 pplication/（实现 interface）

### 2.5 Repository 所属位置

Repository **interface** 定义在 pplication/ 层（Application Service 依赖）。
Repository **implementation** 在 infrastructure/repository/。

### 2.6 禁止跨 Module 直接访问

MVP 阶段只有 Device Module，无跨 Module 问题。
未来新增 Module 时：
- 禁止直接访问 Device Module 的 infrastructure/
- 禁止直接访问 Device Module 的 domain/ entity
- 必须通过 Device Module 的 Application Service

---

## 3. Domain Model

### 3.1 DeviceType

**类型**: Entity

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string (ULID) | 全局唯一标识 |
| name | string | 设备类型名称 |
| vendor | string | 厂商 |
| model | string | 型号 |
| description | string | 描述 |
| capabilities | []Capability | 支持的 Capability 列表 |
| created_at | time.Time | 创建时间 |
| updated_at | time.Time | 更新时间 |

**不变量**：
- (name, vendor, model) 组合全局唯一
- capabilities 不能为空（至少一个 Capability）
- 有 Device 关联时，不能移除已有 Capability

**生命周期**：
- 创建 → 被引用 → （无引用时）删除

**关键业务规则**：
- BR-001: (name + vendor + model) 唯一
- BR-005: 有 Device 关联时不允许删除
- BR-007: Capability 由此定义，Device 继承

### 3.2 Device

**类型**: Entity

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string (ULID) | 全局唯一标识 |
| name | string | 设备名称（不要求全局唯一） |
| device_type_id | string (ULID) | 所属 DeviceType |
| address | string | 连接地址 |
| status | DeviceStatus | 当前状态 |
| last_online_at | *time.Time | 最后在线时间 |
| created_at | time.Time | 注册时间 |
| updated_at | time.Time | 更新时间 |

**不变量**：
- device_type_id 必须引用存在的 DeviceType
- status 不能为 UNKNOWN（MVP 不使用 UNKNOWN 状态）
- id 全局唯一

**状态**: REGISTERED / ACTIVE / OFFLINE / DISABLED（详见 Section 4）

**生命周期**：
- REGISTERED → ACTIVE → OFFLINE → ACTIVE → ...
- 任意状态 → DISABLED
- DISABLED → REGISTERED（重新启用）

**关键业务规则**：
- BR-002: 必须属于一个 DeviceType
- BR-003: 创建时状态为 REGISTERED
- BR-004: DISABLED 不允许操作
- BR-006: 不物理删除，仅 DISABLED
- BR-016: 同时最多一个写操作 RUNNING/RETRYING

### 3.3 Capability

**类型**: Value Object

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | Capability 标识（QueryStatus / SetVolume / Restart） |

**不变量**：
- name 必须是预定义的枚举值之一

**MVP 预定义值**：
- QueryStatus — 读操作，不可重试，低风险
- SetVolume — 写操作，可重试，低风险
- Restart — 写操作，可重试（最多1次），高风险

**相等性**：按 name 值比较。

### 3.4 DeviceStatus

**类型**: Value Object (Enum)

| 值 | 含义 | 可执行操作 |
|----|------|-----------|
| REGISTERED | 已注册，未验证连接 | QueryStatus |
| ACTIVE | 在线且可用 | 所有操作 |
| OFFLINE | 连接不可达 | QueryStatus（尝试恢复） |
| DISABLED | 管理员禁用 | 无 |

### 3.5 Operation

**类型**: Entity

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string (ULID) | 全局唯一标识 |
| type | OperationType | 操作类型 |
| initiator | string | 操作者标识（X-Operator-ID） |
| device_id | string (ULID) | 目标设备 |
| parameters | JSON | 操作参数 |
| status | OperationStatus | 操作状态 |
| trace_id | string | 追踪 ID |
| created_at | time.Time | 创建时间 |
| completed_at | *time.Time | 完成时间 |

**状态**: PENDING / IN_PROGRESS / COMPLETED / PARTIALLY_FAILED / FAILED

**不变量**：
- MVP 阶段 1 Operation : 1 Execution
- status 由 Execution 状态聚合
- trace_id 不可变

### 3.6 Execution

**类型**: Entity

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string (ULID) | 全局唯一标识 |
| operation_id | string (ULID) | 所属 Operation |
| device_id | string (ULID) | 目标设备 |
| capability | Capability | 使用的 Capability |
| status | ExecutionStatus | 执行状态 |
| retry_count | int | 已重试次数 |
| max_retries | int | 最大重试次数 |
| result | JSON | 执行结果 |
| error | string | 错误原因 |
| trace_id | string | 追踪 ID |
| started_at | *time.Time | 开始时间 |
| completed_at | *time.Time | 完成时间 |

**状态**: PENDING / RUNNING / SUCCESS / FAILED / RETRYING / SKIPPED

**不变量**：
- retry_count <= max_retries
- SUCCESS/FAILED/SKIPPED 为终态
- result 在 SUCCESS 时有值，error 在 FAILED 时有值

### 3.7 AuditEvent

**类型**: Entity (Append-only)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string (ULID) | 全局唯一标识 |
| actor | string | 操作者（X-Operator-ID） |
| action | AuditAction | 操作类型 |
| target_type | string | 目标类型（device / device_type / operation） |
| target_id | string | 目标 ID |
| parameters | JSON | 操作参数 |
| result | string | 结果（success / failed） |
| trace_id | string | 追踪 ID |
| created_at | time.Time | 创建时间 |

**不变量**：
- 创建后不可修改、不可删除（append-only）
- 无 UPDATE/DELETE 接口

---

## 4. Device Domain Design

### 4.1 Device State Machine

```text
                    ┌─────────────────┐
                    │   REGISTERED    │ ← Initial state
                    └────────┬────────┘
                             │ QueryStatus SUCCESS
                             ↓
                    ┌─────────────────┐
          ┌────────│     ACTIVE      │────────┐
          │        └────────┬────────┘        │
          │                 │                  │
QueryStatus             QueryStatus         Admin
FAIL/TIMEOUT            SUCCESS             action
          │                 │                  │
          ↓                 ↓                  ↓
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│     OFFLINE     │  │     ACTIVE      │  │    DISABLED     │
└────────┬────────┘  └─────────────────┘  └────────┬────────┘
         │                                         │
         │ QueryStatus SUCCESS                     │ Admin action
         ↓                                         ↓
┌─────────────────┐                       ┌─────────────────┐
│     ACTIVE      │                       │   REGISTERED    │
└─────────────────┘                       └─────────────────┘
```

### 4.2 State Transitions

| From | To | Trigger | Validation |
|------|-----|---------|------------|
| REGISTERED | ACTIVE | QueryStatus SUCCESS | - |
| ACTIVE | OFFLINE | QueryStatus FAIL/TIMEOUT | - |
| OFFLINE | ACTIVE | QueryStatus SUCCESS | - |
| * (any) | DISABLED | Admin disable | - |
| DISABLED | REGISTERED | Admin enable | - |

### 4.3 State-Specific Operation Rules

| State | QueryStatus | SetVolume | Restart |
|-------|-------------|-----------|---------|
| REGISTERED | ✅ Allowed | ❌ Not ACTIVE | ❌ Not ACTIVE |
| ACTIVE | ✅ Allowed | ✅ Allowed | ✅ Allowed |
| OFFLINE | ✅ Allowed (recovery attempt) | ❌ Not ACTIVE | ❌ Not ACTIVE |
| DISABLED | ❌ Disabled | ❌ Disabled | ❌ Disabled |

### 4.4 QueryStatus Behavior

**Success Path**:
1. Execute QueryStatus via DeviceGateway
2. Gateway returns: ONLINE status + device metrics
3. Update Device.status = ACTIVE
4. Update Device.last_online_at = now
5. Execution.status = SUCCESS

**Failure Path**:
1. Execute QueryStatus via DeviceGateway
2. Gateway returns: error or timeout
3. Update Device.status = OFFLINE
4. Execution.status = FAILED
5. Execution.error = reason

> **MVP Decision**: Single QueryStatus failure → OFFLINE
> No heartbeat, no sliding window, no failure counting.

### 4.5 DISABLED Behavior

- DISABLED Device 不接受任何操作
- 尝试操作返回 403 FORBIDDEN
- DISABLED 不影响历史 Operation/Execution/Audit 查询
- 重新启用后状态回到 REGISTERED（需重新 QueryStatus 验证）

---

## 5. Capability Design

### 5.1 QueryStatus

| Attribute | Value |
|-----------|-------|
| **Identifier** | QueryStatus |
| **Type** | Read operation |
| **High Risk** | No |
| **Input** | device_id |
| **Output** | { online: bool, volume: int?, uptime: int?, last_seen: timestamp } |
| **Preconditions** | Device exists, status ≠ DISABLED |
| **Retry Policy** | No retry (read operation) |
| **Confirmation** | No |
| **Device State Change** | SUCCESS → ACTIVE; FAIL → OFFLINE |

**Failure Behavior**:
- Device unreachable → Execution FAILED, Device → OFFLINE
- Simulator timeout → Execution FAILED, Device → OFFLINE
- Device DISABLED → Rejected before Operation creation (403)

### 5.2 SetVolume

| Attribute | Value |
|-----------|-------|
| **Identifier** | SetVolume |
| **Type** | Write operation |
| **High Risk** | No |
| **Input** | device_id, olume (0-100, integer) |
| **Output** | { volume: int, applied_at: timestamp } |
| **Preconditions** | Device exists, status = ACTIVE, DeviceType supports SetVolume, no active write operation |
| **Retry Policy** | Max 3 retries, fixed interval |
| **Confirmation** | No |
| **Device State Change** | None (status unchanged) |

**Failure Behavior**:
- volume < 0 or > 100 → 400 Validation Error, no Operation created
- Device OFFLINE → 403 Operation rejected
- Device has active write → 409 DEVICE_BUSY
- Simulator error → Execution FAILED, retry if retry_count < max_retries
- Simulator timeout → Execution FAILED, retry if retry_count < max_retries

### 5.3 Restart

| Attribute | Value |
|-----------|-------|
| **Identifier** | Restart |
| **Type** | Write operation |
| **High Risk** | **Yes** |
| **Input** | device_id |
| **Output** | { restarted: bool, online_after_restart: bool, duration_ms: int } |
| **Preconditions** | Device exists, status = ACTIVE, DeviceType supports Restart, no active write operation |
| **Retry Policy** | Max 1 retry (high risk, limit retries) |
| **Confirmation** | **Yes** (frontend confirmation dialog) |
| **Device State Change** | During restart: temporary OFFLINE; After recovery: ACTIVE |

**Restart SUCCESS Definition** (Frozen Decision):

Restart SUCCESS requires ALL of the following:
1. ✅ Restart command accepted by device
2. ✅ Device enters restart state (OFFLINE)
3. ✅ Device becomes ONLINE again
4. ✅ Status verification succeeds (internal QueryStatus confirms)

**Implementation Flow**:
```text
1. Send Restart command via DeviceGateway
2. Wait for Simulator to complete restart (sync model)
3. Simulator: mark OFFLINE → delay → mark ONLINE
4. After restart command returns:
   5. Execute internal QueryStatus to verify device is ONLINE
   6. If QueryStatus SUCCESS → Execution = SUCCESS
   7. If QueryStatus FAIL → Execution = FAILED (device did not recover)
```

**Failure Behavior**:
- Device OFFLINE → 403 Operation rejected
- Device has active write → 409 DEVICE_BUSY
- Restart command fails → Execution FAILED
- Restart command succeeds but device not ONLINE → Execution FAILED
- User cancels confirmation → No Operation created

### 5.4 Capability Validation Flow

```text
ExecuteOperation(device_id, capability, params)
    │
    ├── 1. Load Device
    ├── 2. Load DeviceType (via device.device_type_id)
    ├── 3. Check: DeviceType.capabilities contains capability?
    │       └── NO → Return DEVICE_NOT_SUPPORTED
    ├── 4. Check: Device.status allows this operation?
    │       └── NO → Return DEVICE_NOT_ACTIVE or DEVICE_DISABLED
    ├── 5. Check: No active write operation? (for write ops)
    │       └── NO → Return DEVICE_BUSY
    └── 6. Proceed with execution
```

---

## 6. Operation / Execution Design

### 6.1 Operation Lifecycle

```text
PENDING ────→ IN_PROGRESS ────→ COMPLETED
                             ├──→ FAILED
                             └──→ PARTIALLY_FAILED (MVP unused)
```

| Status | Meaning | Trigger |
|--------|---------|---------|
| PENDING | Created, not started | Operation creation |
| IN_PROGRESS | At least one Execution running | Execution → RUNNING |
| COMPLETED | All Executions succeeded | Execution → SUCCESS |
| FAILED | All Executions failed | Execution → FAILED (final) |
| PARTIALLY_FAILED | Mixed results | MVP unused (single device) |

**状态聚合规则** (MVP: 1 Operation : 1 Execution):
- Execution PENDING → Operation PENDING
- Execution RUNNING/RETRYING → Operation IN_PROGRESS
- Execution SUCCESS → Operation COMPLETED
- Execution FAILED → Operation FAILED
- Execution SKIPPED → Operation FAILED

### 6.2 Execution Lifecycle

```text
PENDING ────→ RUNNING ────→ SUCCESS
                        ├──→ FAILED ────→ RETRYING ────→ RUNNING ────→ SUCCESS/FAILED
                        └──→ SKIPPED
```

| Status | Meaning | Trigger |
|--------|---------|---------|
| PENDING | Created, waiting | Execution creation |
| RUNNING | Executing on device | Start execution |
| SUCCESS | Device operation succeeded | Gateway returns success |
| FAILED | Device operation failed | Gateway returns error, no more retries |
| RETRYING | Failed, retrying | Gateway returns error, has retries |
| SKIPPED | Skipped | Precondition not met |

### 6.3 Operation Creation Flow

```text
POST /api/operations {device_id, type, parameters}
    │
    ├── 1. Validate request DTO
    ├── 2. Load Device + DeviceType
    ├── 3. Validate business rules (Capability, status, busy)
    ├── 4. Generate trace_id (from middleware)
    ├── 5. Create Operation:
    │       id = ULID()
    │       type = request.type
    │       initiator = X-Operator-ID
    │       device_id = request.device_id
    │       parameters = request.parameters
    │       status = PENDING
    │       trace_id = ctx.trace_id
    │       created_at = now()
    ├── 6. Create Execution:
    │       id = ULID()
    │       operation_id = operation.id
    │       device_id = operation.device_id
    │       capability = operation.type → capability
    │       status = PENDING
    │       retry_count = 0
    │       max_retries = capability.retry_policy
    │       trace_id = operation.trace_id
    ├── 7. Create AuditEvent:
    │       actor = X-Operator-ID
    │       action = OPERATION_CREATED
    │       target_type = "operation"
    │       target_id = operation.id
    │       parameters = {device_id, type, parameters}
    │       result = "created"
    │       trace_id = operation.trace_id
    ├── 8. Persist: Operation + Execution + AuditEvent (single transaction)
    └── 9. Return Operation (with Execution) as PENDING
```

### 6.4 Operation Execution Flow

```text
After creation (sync model):
    │
    ├── 1. Update Execution: status = RUNNING, started_at = now()
    ├── 2. Call DeviceGateway.Execute(capability, params)
    │       ├── SimulatorAdapter processes request
    │       └── Returns: result or error
    ├── 3. Handle result:
    │       ├── SUCCESS:
    │       │   ├── Execution.status = SUCCESS
    │       │   ├── Execution.result = gateway_result
    │       │   └── Execution.completed_at = now()
    │       ├── FAILED (with retries remaining):
    │       │   ├── Execution.status = RETRYING
    │       │   ├── Execution.retry_count++
    │       │   └── Re-execute from step 1
    │       └── FAILED (no retries):
    │           ├── Execution.status = FAILED
    │           ├── Execution.error = error_reason
    │           └── Execution.completed_at = now()
    ├── 4. For Restart: verify device recovered (internal QueryStatus)
    ├── 5. For QueryStatus: update Device status
    ├── 6. Update Operation.status (aggregate from Execution)
    ├── 7. Update Operation.completed_at = now()
    ├── 8. Update AuditEvent: result = SUCCESS/FAILED
    ├── 9. Persist updates (single transaction)
    └── 10. Return Operation (with Execution, result, trace_id)
```

### 6.5 Synchronous Execution Model

MVP HTTP POST 同步等待 Execution 完成并返回结果。

**关键点**：
- HTTP Request 等待整个 Operation 流程完成
- 但 **领域模型独立于 HTTP 生命周期**
- Operation/Execution 有完整状态机，不依赖 HTTP 连接
- 如果 HTTP 连接断开，Operation 仍然存在于数据库
- 未来可改为异步模式，不影响领域模型

**超时处理**：
- HTTP Client 超时不影响后端执行
- 后端 Gateway 调用有独立超时配置
- 如果 Gateway 超时 → Execution FAILED

---

## 7. Audit Design

### 7.1 Audit vs Operation/Execution

| Concept | Purpose | Records |
|---------|---------|---------|
| Operation / Execution | 业务执行事实 | Lifecycle, retry, result |
| AuditEvent | 安全审计事件 | Actor, action, target, trace_id |

**关键区别**：
- Operation/Execution 关注"发生了什么业务行为"
- AuditEvent 关注"谁在什么时间对什么目标做了什么操作"
- AuditEvent 是独立的、不可变的安全事件

### 7.2 AuditEvent Creation

| Action | AuditEvent |
|--------|------------|
| Device registered | DEVICE_REGISTERED |
| Device disabled | DEVICE_DISABLED |
| Device enabled | DEVICE_ENABLED |
| DeviceType created | DEVICE_TYPE_CREATED |
| DeviceType updated | DEVICE_TYPE_UPDATED |
| DeviceType deleted | DEVICE_TYPE_DELETED |
| Operation created | OPERATION_CREATED |
| Operation completed | OPERATION_COMPLETED |
| Operation failed | OPERATION_FAILED |
| QueryStatus executed | QUERY_STATUS (optional) |

### 7.3 AuditEvent Immutability

- **Append-only**: 只能创建，不能修改、不能删除
- **无 UPDATE/DELETE 接口**: Repository 只有 Insert 和 Query
- **数据库约束**: 表无 UPDATE/DELETE trigger（实现时确认）

### 7.4 QueryStatus Audit Semantics

QueryStatus 可产生 Audit 记录，但 QueryStatus 不属于状态变更型写操作。

- QueryStatus 创建 Operation + Execution（用于追踪）
- QueryStatus 可选创建 AuditEvent（配置决定）
- QueryStatus 不受 DEVICE_BUSY 限制（读操作）
- QueryStatus 在 OFFLINE 设备上仍然允许（尝试恢复）

---

## 8. Device Gateway / SimulatorAdapter Design

### 8.1 Device Gateway Interface

Device Gateway 定义在 domain/gateway.go，是业务层与设备通信的抽象接口。

```go
// DeviceGateway defines the interface for device communication.
// Implementation: SimulatorAdapter (MVP), real adapters (future).
type DeviceGateway interface {
    // QueryStatus queries device current status
    QueryStatus(ctx context.Context, deviceID string) (*StatusResult, error)

    // SetVolume sets device volume (0-100)
    SetVolume(ctx context.Context, deviceID string, volume int) (*VolumeResult, error)

    // Restart restarts the device
    Restart(ctx context.Context, deviceID string) (*RestartResult, error)
}

// StatusResult represents QueryStatus result
type StatusResult struct {
    Online   bool
    Volume   *int       // nil if not supported
    Uptime   int        // seconds
    LastSeen time.Time
}

// VolumeResult represents SetVolume result
type VolumeResult struct {
    Volume    int
    AppliedAt time.Time
}

// RestartResult represents Restart result
type RestartResult struct {
    Restarted  bool
    DurationMs int
}
```

### 8.2 Gateway Error Types

```go
// Gateway errors (domain-level)
type GatewayError struct {
    Code    GatewayErrorCode
    Message string
}

type GatewayErrorCode string

const (
    GatewayTimeout      GatewayErrorCode = "TIMEOUT"
    GatewayUnreachable  GatewayErrorCode = "UNREACHABLE"
    GatewayDeviceError  GatewayErrorCode = "DEVICE_ERROR"
    GatewayUnsupported  GatewayErrorCode = "UNSUPPORTED"
)
```

### 8.3 SimulatorAdapter Architecture

```text
Application Service
    ↓
Device Gateway (interface)
    ↓
SimulatorAdapter (implements Gateway)
    ├── Maintains Simulator instances (per device)
    ├── Translates Gateway calls to Simulator calls
    └── Returns results in Gateway format
    ↓
Simulator (in-memory state machine)
    ├── Per-device state (online, volume, mode)
    ├── Configurable behavior (normal/error/timeout)
    └── Simulates device behavior
```

### 8.4 Simulator State Machine

```text
Simulator Instance:
├── state: ONLINE / OFFLINE
├── mode: NORMAL / ERROR / TIMEOUT / OFFLINE
├── volume: int (0-100)
├── delay_ms: int
└── restart_delay_ms: int
```

**Behavior by mode**:
- NORMAL: Process commands normally, return success
- ERROR: Return error for all commands
- TIMEOUT: Delay beyond timeout threshold, return timeout error
- OFFLINE: Behave as unreachable device

### 8.5 Restart Simulation

```text
Restart command received:
    1. Mark simulator state = OFFLINE
    2. Wait restart_delay_ms (configurable)
    3. Mark simulator state = ONLINE
    4. Return RestartResult {restarted: true, duration_ms: restart_delay_ms}
```

**关键**：Restart 完成后，后续的 QueryStatus 必须返回 ONLINE。

### 8.6 Simulator Lifecycle

- **MVP**: Simulator 作为进程内组件，与 Backend 同进程
- **初始化**: Application 启动时创建 SimulatorAdapter，注册到 Gateway Registry
- **配置**: 每台 Simulator 实例维护独立状态
- **控制**: 测试代码可直接设置 Simulator 行为模式
- **归属**: infrastructure/adapter/ (implementation in device module)

### 8.7 Gateway Contract Test

```text
Gateway Contract Test Suite:
    ├── QueryStatus: normal mode → SUCCESS
    ├── QueryStatus: offline mode → UNREACHABLE
    ├── QueryStatus: timeout mode → TIMEOUT
    ├── SetVolume: normal mode → SUCCESS
    ├── SetVolume: error mode → DEVICE_ERROR
    ├── SetVolume: volume boundary (0, 100)
    ├── Restart: normal mode → SUCCESS + subsequent QueryStatus = ONLINE
    ├── Restart: error mode → DEVICE_ERROR
    └── All methods: context cancellation → TIMEOUT
```

SimulatorAdapter **必须**通过 Gateway Contract Test。
Simulator 不能替代 Contract Test。

---

## 9. Application Service Design

### 9.1 DeviceTypeAppService

**职责**: DeviceType CRUD 编排

**方法**:

```text
CreateDeviceType(ctx, request) → DeviceType
    ├── Validate: name + vendor + model uniqueness
    ├── Create DeviceType entity
    ├── Persist via Repository
    ├── Create AuditEvent (DEVICE_TYPE_CREATED)
    └── Return DeviceType

UpdateDeviceType(ctx, id, request) → DeviceType
    ├── Load DeviceType
    ├── Validate: no Device references if removing capabilities
    ├── Update DeviceType entity
    ├── Persist via Repository
    ├── Create AuditEvent (DEVICE_TYPE_UPDATED)
    └── Return DeviceType

DeleteDeviceType(ctx, id) → void
    ├── Load DeviceType
    ├── Check: no Device references (BR-005)
    ├── Delete via Repository
    ├── Create AuditEvent (DEVICE_TYPE_DELETED)
    └── Return

GetDeviceType(ctx, id) → DeviceType
    └── Load via Repository

ListDeviceTypes(ctx, filter) → [DeviceType]
    └── Query via Repository
```

### 9.2 DeviceAppService

**职责**: Device 注册、管理、状态维护

**方法**:

```text
RegisterDevice(ctx, request) → Device
    ├── Validate: DeviceType exists
    ├── Create Device entity (status = REGISTERED)
    ├── Persist via Repository
    ├── Create AuditEvent (DEVICE_REGISTERED)
    └── Return Device

UpdateDevice(ctx, id, request) → Device
    ├── Load Device
    ├── Validate: not DISABLED
    ├── Update Device entity
    ├── Persist via Repository
    └── Return Device

DisableDevice(ctx, id) → Device
    ├── Load Device
    ├── Validate: not already DISABLED
    ├── Update Device.status = DISABLED
    ├── Persist via Repository
    ├── Create AuditEvent (DEVICE_DISABLED)
    └── Return Device

EnableDevice(ctx, id) → Device
    ├── Load Device
    ├── Validate: currently DISABLED
    ├── Update Device.status = REGISTERED
    ├── Persist via Repository
    ├── Create AuditEvent (DEVICE_ENABLED)
    └── Return Device

GetDevice(ctx, id) → Device
    ├── Load Device via Repository
    ├── Load DeviceType via Repository
    └── Return Device (with DeviceType capabilities)

ListDevices(ctx, filter) → [Device]
    └── Query via Repository (with pagination, search)
```

### 9.3 OperationAppService

**职责**: Operation 创建、执行、追踪

**方法**:

```text
ExecuteOperation(ctx, request) → Operation
    ├── 1. Validate request DTO
    ├── 2. Load Device + DeviceType
    ├── 3. Validate: Device exists
    ├── 4. Validate: Device.status ≠ DISABLED
    ├── 5. Validate: DeviceType.capabilities contains request.capability
    ├── 6. Validate: Device.status = ACTIVE (for write operations)
    ├── 7. Validate: No active write operation (BR-016 DEVICE_BUSY)
    ├── 8. Create Operation (PENDING)
    ├── 9. Create Execution (PENDING)
    ├── 10. Create AuditEvent (OPERATION_CREATED)
    ├── 11. Persist (transaction)
    ├── 12. Execute:
    │       ├── Update Execution → RUNNING
    │       ├── Call DeviceGateway.Execute(capability, params)
    │       ├── Handle result (SUCCESS/FAILED/RETRYING)
    │       ├── For Restart: verify device recovered
    │       ├── For QueryStatus: update Device status
    │       ├── Update Execution → SUCCESS/FAILED
    │       ├── Update Operation status (aggregate)
    │       ├── Update AuditEvent result
    │       └── Persist updates (transaction)
    └── 13. Return Operation (with Execution, trace_id)

GetOperation(ctx, id) → Operation
    ├── Load Operation via Repository
    ├── Load Executions via Repository
    └── Return Operation (with Executions)

ListDeviceOperations(ctx, device_id, filter) → [Operation]
    └── Query via Repository (with pagination)
```

### 9.4 Application Service 依赖

| Service | Repository | Gateway |
|---------|-----------|---------|
| DeviceTypeAppService | DeviceTypeRepository | - |
| DeviceAppService | DeviceRepository, DeviceTypeRepository | - |
| OperationAppService | OperationRepository, ExecutionRepository, DeviceRepository, AuditRepository | DeviceGateway |

---

## 10. Database Design

### 10.1 Table Design (Conceptual)

#### device_types

| Column | Type | Constraints |
|--------|------|-------------|
| id | VARCHAR(26) | PK (ULID) |
| name | VARCHAR(100) | NOT NULL |
| vendor | VARCHAR(100) | NOT NULL |
| model | VARCHAR(100) | NOT NULL |
| description | TEXT | - |
| capabilities | JSON | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |
| **UNIQUE** | (name, vendor, model) | - |

#### devices

| Column | Type | Constraints |
|--------|------|-------------|
| id | VARCHAR(26) | PK (ULID) |
| name | VARCHAR(100) | NOT NULL |
| device_type_id | VARCHAR(26) | FK → device_types.id |
| address | VARCHAR(255) | NOT NULL |
| status | VARCHAR(20) | NOT NULL (REGISTERED/ACTIVE/OFFLINE/DISABLED) |
| last_online_at | TIMESTAMP | NULL |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |
| **INDEX** | (device_type_id) | - |
| **INDEX** | (status) | - |

#### operations

| Column | Type | Constraints |
|--------|------|-------------|
| id | VARCHAR(26) | PK (ULID) |
| type | VARCHAR(20) | NOT NULL (QUERY_STATUS/SET_VOLUME/RESTART) |
| initiator | VARCHAR(100) | NOT NULL |
| device_id | VARCHAR(26) | FK → devices.id |
| parameters | JSON | NULL |
| status | VARCHAR(20) | NOT NULL |
| trace_id | VARCHAR(32) | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| completed_at | TIMESTAMP | NULL |
| **INDEX** | (device_id, created_at) | - |
| **INDEX** | (status) | - |
| **INDEX** | (trace_id) | - |

#### executions

| Column | Type | Constraints |
|--------|------|-------------|
| id | VARCHAR(26) | PK (ULID) |
| operation_id | VARCHAR(26) | FK → operations.id |
| device_id | VARCHAR(26) | FK → devices.id |
| capability | VARCHAR(20) | NOT NULL |
| status | VARCHAR(20) | NOT NULL |
| retry_count | INT | NOT NULL, DEFAULT 0 |
| max_retries | INT | NOT NULL |
| result | JSON | NULL |
| error | TEXT | NULL |
| trace_id | VARCHAR(32) | NOT NULL |
| started_at | TIMESTAMP | NULL |
| completed_at | TIMESTAMP | NULL |
| **INDEX** | (operation_id) | - |
| **INDEX** | (device_id, status) | - |

#### audit_events

| Column | Type | Constraints |
|--------|------|-------------|
| id | VARCHAR(26) | PK (ULID) |
| actor | VARCHAR(100) | NOT NULL |
| action | VARCHAR(50) | NOT NULL |
| target_type | VARCHAR(50) | NOT NULL |
| target_id | VARCHAR(26) | NOT NULL |
| parameters | JSON | NULL |
| result | VARCHAR(20) | NOT NULL |
| trace_id | VARCHAR(32) | NOT NULL |
| created_at | TIMESTAMP | NOT NULL |
| **INDEX** | (target_type, target_id) | - |
| **INDEX** | (actor, created_at) | - |
| **INDEX** | (trace_id) | - |

### 10.2 JSON Column Strategy

**capabilities** (device_types):
```json
["QueryStatus", "SetVolume", "Restart"]
```

**parameters** (operations):
```json
{"volume": 70}
```

**result** (executions):
```json
{"volume": 70, "applied_at": "2026-09-24T10:00:00Z"}
```

### 10.3 Database Strategy

| Environment | Database | Use Case |
|-------------|----------|----------|
| Production | MySQL / GreatDB | Real deployment |
| Local Dev | SQLite | Zero dependency |
| Unit Test | SQLite / In-memory | Fast execution |
| Integration Test | MySQL | Verify real Repository behavior |

### 10.4 SQL Compatibility

优先使用 MySQL / GreatDB 共同兼容的能力。
SQLite 仅作为本地开发和快速测试便利，不以 SQLite 兼容性约束生产数据库设计。

---

## 11. Concurrency Model

### 11.1 DEVICE_BUSY Rule (BR-016)

**规则**: 同一 Device 同时最多允许一个写操作处于 RUNNING 或 RETRYING 状态。

**实现策略**:

```text
Before creating write Operation:
    1. Query executions table:
       SELECT * FROM executions
       WHERE device_id = ?
         AND status IN ('RUNNING', 'RETRYING')
         AND capability IN ('SetVolume', 'Restart')
    2. If result count > 0:
       → Return 409 DEVICE_BUSY
    3. Else:
       → Proceed with Operation creation
```

### 11.2 Concurrency Scope

- **MVP**: 单进程应用，无分布式锁需求
- **并发控制**: 数据库查询 + 应用层检查
- **Race condition**: 极低概率（单设备写操作频率低）
- **未来**: 如需高并发，可引入行锁或分布式锁

### 11.3 QueryStatus 不受 DEVICE_BUSY 限制

- QueryStatus 是读操作，不检查 DEVICE_BUSY
- QueryStatus 可以与写操作并发执行
- QueryStatus 不产生 DEVICE_BUSY 冲突

---

## 12. ID Strategy

### 12.1 ID Generation Strategy

**选择**: ULID (Universally Unique Lexicographically Sortable Identifier)

**理由**:
- 全局唯一，无中心化生成
- 时间有序（按创建时间排序）
- 字符串格式，易于调试
- 26 字符，比 UUID 短
- 无需数据库自增依赖

**ULID 结构**:
```text
01ARZ3NDEKTSV4RRFFQ69G5FAV
│     │
│     └─ Randomness (80 bits)
└─────── Timestamp (48 bits, milliseconds since Unix epoch)
```

### 12.2 ID Usage

| Entity | ID Format | Example |
|--------|-----------|---------|
| DeviceType | ULID |  1ARZ3NDEKTSV4RRFFQ69G5FAV |
| Device | ULID |  1ARZ3NDEKTSV4RRFFQ69G5FAV |
| Operation | ULID |  1ARZ3NDEKTSV4RRFFQ69G5FAV |
| Execution | ULID |  1ARZ3NDEKTSV4RRFFQ69G5FAV |
| AuditEvent | ULID |  1ARZ3NDEKTSV4RRFFQ69G5FAV |

### 12.3 ID Generation Library

**Go**: github.com/oklog/ulid/v2

```go
import "github.com/oklog/ulid/v2"

func NewID() string {
    return ulid.Make().String()
}
```

---

## 13. API Boundary

### 13.1 Controller → DTO → Service Pattern

```text
HTTP Request
    ↓
Controller (interfaces/)
    ├── Bind request body → Request DTO
    ├── Validate DTO (binding tags)
    ├── Extract: X-Operator-ID, trace_id from context
    ├── Call Application Service
    ↓
Application Service (application/)
    ├── Convert Request DTO → Domain parameters
    ├── Execute business logic
    ├── Return Domain result
    ↓
Controller
    ├── Convert Domain result → Response DTO
    ├── Return HTTP Response (unified format)
```

**禁止**:
`
HTTP Request → Domain Entity
`

### 13.2 X-Operator-ID Boundary

| Aspect | Definition |
|--------|-----------|
| **Purpose** | MVP 操作者标识 |
| **Source** | HTTP Header X-Operator-ID |
| **Used by** | Operation.initiator, AuditEvent.actor |
| **NOT** | Authentication |
| **NOT** | Authorization |
| **NOT** | RBAC |
| **Security** | 不构成安全边界 |

**Middleware 行为**:
- 提取 X-Operator-ID header
- 如果缺失，使用默认值 nonymous
- 存入 context，供 Controller 使用

### 13.3 API Endpoint Overview

| Method | Path | Handler | Service Method |
|--------|------|---------|---------------|
| GET | /api/device-types | DeviceTypeHandler.List | DeviceTypeAppService.ListDeviceTypes |
| POST | /api/device-types | DeviceTypeHandler.Create | DeviceTypeAppService.CreateDeviceType |
| GET | /api/device-types/:id | DeviceTypeHandler.Get | DeviceTypeAppService.GetDeviceType |
| PUT | /api/device-types/:id | DeviceTypeHandler.Update | DeviceTypeAppService.UpdateDeviceType |
| DELETE | /api/device-types/:id | DeviceTypeHandler.Delete | DeviceTypeAppService.DeleteDeviceType |
| GET | /api/devices | DeviceHandler.List | DeviceAppService.ListDevices |
| POST | /api/devices | DeviceHandler.Create | DeviceAppService.RegisterDevice |
| GET | /api/devices/:id | DeviceHandler.Get | DeviceAppService.GetDevice |
| PUT | /api/devices/:id | DeviceHandler.Update | DeviceAppService.UpdateDevice |
| PATCH | /api/devices/:id/status | DeviceHandler.UpdateStatus | DeviceAppService.DisableDevice / EnableDevice |
| POST | /api/operations | OperationHandler.Create | OperationAppService.ExecuteOperation |
| GET | /api/operations/:id | OperationHandler.Get | OperationAppService.GetOperation |
| GET | /api/devices/:id/operations | OperationHandler.ListByDevice | OperationAppService.ListDeviceOperations |

### 13.4 Response Format

所有 API 响应使用统一格式（已有 internal/platform/response/）：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

Error:
```json
{
  "code": 2001,
  "message": "device not found"
}
```

---

## 14. Frontend Technical Design

### 14.1 Frontend Architecture

```text
Browser
    ↓
Vue 3 + TypeScript + Vite
    ↓
Views (pages)
    ↓
API Client (web/src/api/)
    ↓
HTTP REST API
    ↓
Go Backend
```

**约束**:
- 前端不实现业务规则
- 前端不维护业务状态
- 前端不直接访问数据库或设备
- 所有数据通过 API Client 从 Backend 获取

### 14.2 Page Design

| Page | Route | Function |
|------|-------|----------|
| Dashboard | /dashboard | System overview (existing + device stats) |
| DeviceType List | /device-types | List all device types |
| DeviceType Detail | /device-types/:id | View/edit device type + capabilities |
| Device List | /devices | List all devices (pagination, search) |
| Device Detail | /devices/:id | View device + status + capabilities + operations |
| Operation Result | /operations/:id | View operation + execution + trace_id |

### 14.3 Frontend Components

| Component | Purpose |
|-----------|---------|
| DeviceStatusBadge | Display device status with color |
| CapabilityList | Display device capabilities |
| OperationStatusBadge | Display operation status |
| ExecutionDetail | Display execution result/error |
| ConfirmDialog | Confirmation for high-risk operations |
| OperationForm | SetVolume form, Restart button |

### 14.4 Frontend Constraints

**禁止实现**:
- Capability business rule validation
- Device state machine logic
- Operation state machine logic
- Retry logic
- Device permission decision

**必须实现**:
- Restart 确认对话框（高风险操作）
- DEVICE_BUSY 提示（显示已有操作在执行）
- Loading / Error state
- trace_id 展示

### 14.5 Restart Confirmation Flow

```text
User clicks "Restart Device"
    ↓
Show ConfirmDialog:
    "确认重启设备 [device_name]？
     重启过程中设备将短暂离线。
     此操作需要确认。"
    ↓
User confirms → POST /api/operations
User cancels → No API call
```

### 14.6 Operation Result Display

Operation Result 页面至少展示：
- operation_id
- execution_id
- trace_id
- status (with color badge)
- result (if SUCCESS)
- error (if FAILED)
- created_at / completed_at
- device name
- capability
- parameters

---

## 15. Error Model

### 15.1 Domain Errors

| Error Code | Name | HTTP Status | Description |
|-----------|------|-------------|-------------|
| 2001 | DEVICE_NOT_FOUND | 404 | Device ID does not exist |
| 2002 | DEVICE_DISABLED | 403 | Device is DISABLED |
| 2003 | DEVICE_NOT_ACTIVE | 403 | Device is not ACTIVE (for write ops) |
| 2004 | DEVICE_BUSY | 409 | Device has active write operation |
| 2005 | DEVICE_NOT_SUPPORTED | 400 | DeviceType does not support capability |
| 2006 | DEVICE_OFFLINE | 403 | Device is OFFLINE (for write ops) |
| 2007 | DEVICE_TIMEOUT | 504 | Device communication timeout |
| 2008 | INVALID_VOLUME | 400 | Volume out of range (0-100) |
| 2009 | OPERATION_FAILED | 500 | Operation execution failed |
| 2010 | DEVICE_TYPE_NOT_FOUND | 404 | DeviceType ID does not exist |
| 2011 | DEVICE_TYPE_EXISTS | 409 | (name+vendor+model) already exists |
| 2012 | DEVICE_TYPE_IN_USE | 409 | DeviceType has associated Devices |

### 15.2 Error Layer Mapping

```text
Domain Error
    ↓
Application Error (AppError with Code, Message, HTTPStatus)
    ↓
HTTP Response
```

**Domain → Application mapping**:
| Domain Error | AppError Code | HTTP Status |
|-------------|---------------|-------------|
| DEVICE_NOT_FOUND | 2001 | 404 |
| DEVICE_DISABLED | 2002 | 403 |
| DEVICE_BUSY | 2004 | 409 |
| INVALID_VOLUME | 2008 | 400 |
| OPERATION_FAILED | 2009 | 500 |
| ... | ... | ... |

### 15.3 Error Response Format

```json
{
  "code": 2004,
  "message": "device has active operation",
  "data": {
    "active_operation_id": "01ARZ3NDEK..."
  }
}
```

### 15.4 Error Code Ranges

| Range | Category |
|-------|----------|
| 0 | Success |
| 1000-1999 | System errors (existing platform) |
| 2000-2999 | Device module business errors |
| 3000-3999 | Reserved for future modules |

---

## 16. Transaction Boundary

### 16.1 Database Transaction Scope

**原则**: Database transaction 与 Device execution 是两个不同的概念。

```text
┌─── DB Transaction 1 ───┐
│ Create Operation       │
│ Create Execution       │
│ Create AuditEvent      │
│ (all or nothing)       │
└────────────────────────┘
         ↓
   ┌── Device Execution ──┐
   │ Gateway call         │  ← NOT in DB transaction
   │ (network I/O)        │
   └──────────────────────┘
         ↓
┌─── DB Transaction 2 ───┐
│ Update Execution       │
│ Update Operation       │
│ Update Device (if QS)  │
│ Update AuditEvent      │
│ (all or nothing)       │
└────────────────────────┘
```

### 16.2 Transaction Rules

| Operation | Transaction Scope |
|-----------|------------------|
| Operation + Execution creation | Single DB transaction |
| Gateway call | NOT in DB transaction |
| Result persistence | Single DB transaction |
| DeviceType CRUD | Single DB transaction |
| Device CRUD | Single DB transaction |
| AuditEvent creation | Same transaction as parent operation |

### 16.3 Gateway Call Isolation

**为什么 Gateway 调用不在 DB transaction 中**：
- Gateway 调用涉及网络 I/O（即使是 Simulator 也模拟延迟）
- 长时间持有 DB connection 会影响连接池
- 如果 Gateway 超时，不应该阻塞其他 DB 操作
- 事务应该尽可能短

### 16.4 Partial Failure Handling

如果 Gateway 调用成功但 DB transaction 2 失败：
- Execution 状态可能不一致
- 需要补偿机制（future）
- MVP 接受此风险（Simulator 环境）

---

## 17. Testing Design

### 17.1 Test Matrix

遵循 ADR-006 五类测试类型。

| Test Type | Scope | MVP Focus |
|-----------|-------|-----------|
| Unit Test | Domain model behavior | State transitions, business rules, validation |
| Integration Test | API → Repository | Full request flow, DB operations |
| Adapter Contract Test | Gateway interface | SimulatorAdapter correctness |
| E2E Test | Complete user scenarios | Device registration, operation execution |
| Agent Tool Test | Agent boundary | MVP: Not implemented |

### 17.2 Unit Test Cases

**Device State Machine**:
- REGISTERED → ACTIVE (QueryStatus SUCCESS)
- ACTIVE → OFFLINE (QueryStatus FAIL)
- OFFLINE → ACTIVE (QueryStatus SUCCESS)
- * → DISABLED (Admin disable)
- DISABLED → REGISTERED (Admin enable)

**Capability Validation**:
- DeviceType defines QueryStatus, SetVolume, Restart
- Device inherits all capabilities from DeviceType
- Operation rejected if capability not in DeviceType

**Business Rules**:
- BR-001: DeviceType (name+vendor+model) uniqueness
- BR-002: Device must belong to DeviceType
- BR-004: DISABLED device rejects all operations
- BR-005: DeviceType with Devices cannot be deleted
- BR-016: DEVICE_BUSY when active write operation exists
- BR-006: Device no physical delete, only DISABLED

**Validation**:
- Volume 0-100 range (0, 50, 100 valid; -1, 101 invalid)
- Device name not empty
- DeviceType capabilities not empty

**Operation Lifecycle**:
- PENDING → IN_PROGRESS → COMPLETED
- PENDING → IN_PROGRESS → FAILED
- Execution retry logic (RETRYING → RUNNING → SUCCESS/FAILED)

**Execution Lifecycle**:
- PENDING → RUNNING → SUCCESS
- PENDING → RUNNING → FAILED → RETRYING → RUNNING → SUCCESS/FAILED
- PENDING → SKIPPED

### 17.3 Integration Test Cases

**DeviceType CRUD**:
- Create DeviceType with capabilities
- Get DeviceType by ID
- List DeviceTypes (pagination)
- Update DeviceType (add capability)
- Delete DeviceType (no Devices)
- Delete DeviceType (with Devices → reject)

**Device CRUD**:
- Register Device (with valid DeviceType)
- Register Device (with invalid DeviceType → 404)
- Get Device by ID (with DeviceType)
- List Devices (pagination, search, status filter)
- Disable Device
- Enable Device

**Operation Execution**:
- Execute QueryStatus (SUCCESS → ACTIVE)
- Execute QueryStatus (FAIL → OFFLINE)
- Execute SetVolume (SUCCESS)
- Execute SetVolume (volume out of range → 400)
- Execute SetVolume (device OFFLINE → 403)
- Execute Restart (SUCCESS + verify ONLINE)
- Execute Restart (device not recovered → FAILED)
- Execute operation (DEVICE_BUSY → 409)

**Audit**:
- AuditEvent created for Device registration
- AuditEvent created for Operation execution
- AuditEvent append-only (no update/delete API)

### 17.4 Adapter Contract Test Cases

**SimulatorAdapter must pass**:
- QueryStatus: normal mode → StatusResult{online: true}
- QueryStatus: offline mode → GatewayError{UNREACHABLE}
- QueryStatus: timeout mode → GatewayError{TIMEOUT}
- SetVolume: normal mode → VolumeResult{volume: 70}
- SetVolume: error mode → GatewayError{DEVICE_ERROR}
- SetVolume: boundary values (0, 100)
- Restart: normal mode → RestartResult{restarted: true}
- Restart: subsequent QueryStatus → online: true
- Restart: error mode → GatewayError{DEVICE_ERROR}
- Context cancellation → GatewayError{TIMEOUT}

### 17.5 E2E Test Cases

**Scenario A: Register New Device**:
1. Create DeviceType
2. Register Device
3. Verify Device status = REGISTERED
4. Verify AuditEvent created

**Scenario B: Query Device Status**:
1. Register Device
2. Execute QueryStatus (SUCCESS)
3. Verify Device status = ACTIVE
4. Verify Operation COMPLETED
5. Verify Execution SUCCESS

**Scenario C: Execute SetVolume**:
1. Create DeviceType with SetVolume
2. Register Device
3. QueryStatus (make ACTIVE)
4. Execute SetVolume(70)
5. Verify Operation COMPLETED
6. Verify Execution SUCCESS
7. Verify result contains volume=70

**Scenario D: Execute Restart**:
1. Create DeviceType with Restart
2. Register Device
3. QueryStatus (make ACTIVE)
4. Execute Restart
5. Verify Operation COMPLETED
6. Verify Execution SUCCESS
7. Verify Device recovered ONLINE (QueryStatus)

**Scenario E: Operation Failure**:
1. Register Device
2. QueryStatus (make ACTIVE)
3. Set Simulator to error mode
4. Execute SetVolume
5. Verify Operation FAILED
6. Verify Execution FAILED
7. Verify error message

### 17.6 Test Database Strategy

| Test Type | Database |
|-----------|----------|
| Unit Test | In-memory / SQLite |
| Integration Test | MySQL (real Repository behavior) |
| Adapter Contract Test | No DB required |
| E2E Test | MySQL (full stack) |

---

## 18. Design Decisions

### 18.1 Frozen Decisions (from Requirement)

| ID | Decision | Source |
|----|----------|--------|
| D-001 | Capability: DeviceType defines, Device inherits | Requirement §6.3 |
| D-002 | Restart SUCCESS = command + recovery + verification | Requirement §10.3 |
| D-003 | Operation/Execution independent from HTTP lifecycle | Requirement §6.5 |
| D-004 | Audit separate from Operation/Execution | Requirement §6.7 |
| D-005 | OFFLINE = single QueryStatus failure | Requirement §9.1 |
| D-006 | X-Operator-ID is identifier only, not security | Requirement §13.4 |
| D-007 | Single write operation per device | Requirement §8 BR-016 |
| D-008 | Device no physical delete, only DISABLED | Requirement §6.2 |
| D-009 | DeviceType delete only when no Devices | Requirement §6.1 |
| D-010 | QueryStatus creates Op+Exec, optional Audit | Requirement §6.7 |
| D-011 | SimulatorAdapter implements Gateway | Requirement §11.5 |

### 18.2 New Design Decisions

| ID | Decision | Rationale |
|----|----------|-----------|
| DD-001 | **Single Device Module** | All concepts tightly coupled; split when cross-module sharing needed |
| DD-002 | **ULID for all IDs** | Globally unique, time-ordered, no central generation |
| DD-003 | **Capability stored as JSON array** | Simple, flexible for future extension |
| DD-004 | **DEVICE_BUSY via DB query** | Check executions table for RUNNING/RETRYING |
| DD-005 | **Two DB transactions per Operation** | Creation + Result persistence; Gateway call not in transaction |
| DD-006 | **Simulator in-process** | MVP simplicity; same process as Backend |
| DD-007 | **Restart verification via QueryStatus** | After restart, internal QueryStatus confirms ONLINE |
| DD-008 | **QueryStatus Audit optional** | Configurable; not state-changing write |
| DD-009 | **Synchronous HTTP execution** | MVP simplicity; domain model independent |
| DD-010 | **Error codes 2000-2999 for Device module** | Reserved range for future modules |

### 18.3 Rejected Alternatives

**Alternative: Split Operation/Execution into separate Module**
- Rejected: MVP scope too small; concepts tightly coupled to Device
- Revisit: When Operation/Audit needs cross-module sharing

**Alternative: UUID for IDs**
- Rejected: Longer (36 chars), not time-ordered
- Chosen: ULID (26 chars, time-ordered)

**Alternative: Gateway call inside DB transaction**
- Rejected: Network I/O in transaction blocks connections
- Chosen: Two separate transactions

**Alternative: Async Operation execution**
- Rejected: MVP simplicity; sync model sufficient
- Chosen: Sync HTTP, but domain model independent

---

## 19. Open Questions

### 19.1 A. Blocking (Must resolve before Phase 3C)

**None.**

所有 blocking decisions 已在 Requirement 和 Design 阶段解决。

### 19.2 B. Non-blocking (Can resolve during Implementation)

| ID | Question | Impact | Default |
|----|----------|--------|---------|
| OQ-B01 | DeviceType Capability 修改规则 | UpdateDeviceType logic | Allow add, reject remove if referenced |
| OQ-B02 | QueryStatus Audit 是否默认开启 | Audit creation | Configurable, default: false |
| OQ-B03 | Device list 默认排序 | ListDevices query | created_at DESC |
| OQ-B04 | Operation list 默认排序 | ListDeviceOperations query | created_at DESC |
| OQ-B05 | Simulator address 格式 | Device.address validation | simulator://localhost:{port} |
| OQ-B06 | Retry interval strategy | Execution retry logic | Fixed interval (1s, 2s, 3s) |
| OQ-B07 | Gateway timeout duration | SimulatorAdapter config | 5s (configurable) |
| OQ-B08 | Restart verification delay | Post-restart QueryStatus | 1s after restart completes |

### 19.3 C. Future (Post-MVP)

| ID | Question | Context |
|----|----------|---------|
| OQ-F01 | Agent initiator 标识 | When Agent introduced |
| OQ-F02 | Device Capability Override | When per-device customization needed |
| OQ-F03 | OFFLINE 智能判定 (heartbeat/sliding window) | When stability requirements increase |
| OQ-F04 | Operation cancellation | When long-running operations introduced |
| OQ-F05 | Async Operation execution | When scale/performance requires |
| OQ-F06 | Cross-module Operation/Audit | When multiple modules need operations |
| OQ-F07 | Real device protocol adapters | When real devices integrated |
| OQ-F08 | Batch operations | When multi-device operations needed |

---

## 20. Architecture Consistency Check

### 20.1 Technical Design ↔ ARCHITECTURE.md

| Aspect | Consistent? | Notes |
|--------|-------------|-------|
| Modular Monolith | ✅ | Single Device module, follows module structure |
| Module internal layers | ✅ | domain/application/interfaces/infrastructure |
| Dependency direction | ✅ | interfaces→application→domain |
| Device Gateway 4-layer | ✅ | Business→Gateway→Adapter→Transport (Simulator) |
| Operation/Execution model | ✅ | Follows ADR-005 |
| Agent Tool Boundary | ✅ | MVP: not implemented; architecture preserved |
| Defense in Depth | ✅ | Application Service validates independently |
| Database strategy | ✅ | MySQL production, SQLite local, MySQL integration test |
| Testing strategy | ✅ | 5 test types, Adapter Contract Test required |
| Simulator position | ✅ | Infrastructure layer, implements Gateway |
| Frontend boundary | ✅ | HTTP API only, no business logic in frontend |

### 20.2 Technical Design ↔ ADR-002 (Device Abstraction)

| Aspect | Consistent? | Notes |
|--------|-------------|-------|
| 4-layer architecture | ✅ | Business→Gateway→Adapter→Transport |
| Gateway interface in domain | ✅ | domain/gateway.go |
| Adapter in infrastructure | ✅ | infrastructure/adapter/ |
| Business layer no protocol code | ✅ | Gateway abstracts protocol |
| Adapter Contract Test | ✅ | Required for SimulatorAdapter |

### 20.3 Technical Design ↔ ADR-004 (Agent Tool Boundary)

| Aspect | Consistent? | Notes |
|--------|-------------|-------|
| Agent not in MVP | ✅ | MVP: not implemented |
| Defense in Depth | ✅ | Application Service independent validation |
| Tool→Application Service | ✅ | Architecture preserved for future |

### 20.4 Technical Design ↔ ADR-005 (Operation/Execution)

| Aspect | Consistent? | Notes |
|--------|-------------|-------|
| Operation = business intent | ✅ | High-level intent expression |
| Execution = device execution unit | ✅ | Independently traceable, retryable |
| Status machines | ✅ | Operation: PENDING→IN_PROGRESS→COMPLETED/FAILED; Execution: PENDING→RUNNING→SUCCESS/FAILED/RETRYING/SKIPPED |
| Retry at Execution level | ✅ | Execution.retry_count, max_retries |
| Operation status aggregated | ✅ | From Execution status |

### 20.5 Technical Design ↔ ADR-006 (Testing Strategy)

| Aspect | Consistent? | Notes |
|--------|-------------|-------|
| 5 test types | ✅ | Unit, Integration, Contract, E2E, Agent |
| Agent Tool Test MVP | ✅ | Not implemented (MVP scope) |
| Adapter Contract Test required | ✅ | SimulatorAdapter must pass |
| Integration Test uses MySQL | ✅ | Real Repository behavior |
| Unit Test uses SQLite/memory | ✅ | Fast execution |

### 20.6 Technical Design ↔ device-mvp-requirement.md

| Aspect | Consistent? | Notes |
|--------|-------------|-------|
| MVP scope | ✅ | No scope expansion |
| Capability model | ✅ | DeviceType defines, Device inherits |
| Restart SUCCESS definition | ✅ | Command + recovery + verification |
| OFFLINE rule | ✅ | Single QueryStatus failure |
| DEVICE_BUSY rule | ✅ | BR-016 implemented |
| Audit separation | ✅ | Independent AuditEvent |
| Device no physical delete | ✅ | Only DISABLED |
| X-Operator-ID boundary | ✅ | Identifier only, not security |

### 20.7 Consistency Check Result

**✅ PASSED**

Technical Design 与所有架构文档一致，无冲突。

---

## 21. Phase 3C Readiness

### 21.1 Blocking Decisions

**None.**

所有 blocking decisions 已解决。

### 21.2 Non-blocking Decisions

8 项 non-blocking decisions可在 Implementation 阶段解决，不阻塞 API Design。

### 21.3 Architecture Consistency

**✅ PASSED**

Technical Design 与 ARCHITECTURE.md、ADR-002/004/005/006、device-mvp-requirement.md 一致。

### 21.4 Readiness Status

**✅ READY FOR PHASE 3C (API Design)**

Technical Design 已完成，可作为 Phase 3C API Design 的稳定输入。

---

## Appendix A: Glossary

| Term | Definition |
|------|-----------|
| DeviceType | 设备类型定义，包含 Capability 集合 |
| Device | 设备实例，继承 DeviceType 的 Capability |
| Capability | 设备能力（QueryStatus/SetVolume/Restart） |
| Operation | 业务操作意图 |
| Execution | 设备执行单元（可追踪、重试） |
| AuditEvent | 安全审计事件（不可变） |
| DeviceGateway | 设备通信接口（domain 层定义） |
| SimulatorAdapter | Gateway 实现（infrastructure 层） |
| Simulator | 设备模拟器（内存状态机） |
| ULID | Universally Unique Lexicographically Sortable Identifier |
| DEVICE_BUSY | 同一设备有写操作执行中 |

---

## Appendix B: File Structure Preview

```text
internal/modules/device/
├── domain/
│   ├── entity/
│   │   ├── device_type.go
│   │   ├── device.go
│   │   ├── operation.go
│   │   ├── execution.go
│   │   └── audit_event.go
│   ├── value/
│   │   ├── capability.go
│   │   ├── device_status.go
│   │   └── volume.go
│   ├── errors.go
│   └── gateway.go
├── application/
│   ├── device_type_service.go
│   ├── device_service.go
│   ├── operation_service.go
│   └── dto/
│       ├── request/
│       │   ├── create_device_type.go
│       │   ├── register_device.go
│       │   └── execute_operation.go
│       └── response/
│           ├── device_type.go
│           ├── device.go
│           └── operation.go
├── interfaces/
│   ├── device_type_handler.go
│   ├── device_handler.go
│   ├── operation_handler.go
│   └── router.go
└── infrastructure/
    ├── repository/
    │   ├── device_type_repo.go
    │   ├── device_repo.go
    │   ├── operation_repo.go
    │   └── audit_repo.go
    └── adapter/
        ├── simulator_adapter.go
        └── simulator.go
```

---

**Document End**

**Status**: DRAFT
**Next Phase**: Phase 3C — API Design
**Blocking Issues**: None
