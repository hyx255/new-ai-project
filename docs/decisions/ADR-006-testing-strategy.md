# ADR-006：Testing Strategy（测试策略）

## 状态

Accepted

## 日期

2026-09-22

## Context

项目需要建立完整的测试体系，面临以下约束：

- 业务规则需要被精确验证（权限、能力判断、状态流转）
- API 行为需要与契约一致（请求/响应、错误码、参数校验）
- 设备协议交互需要被可靠测试（但真实设备在开发和 CI 环境中不可用）
- AI Agent 的操作边界需要被验证（Agent 不能绕过 Tool 层，Tool 必须做正确校验）
- 测试必须确定性、可重复执行，不能依赖外部服务或随机值
- 测试类型需要清晰，避免重复测试同一行为
- 测试需要支持 AI Coding 工作流（AI 可以快速编写和运行测试）

## Decision

采用五类测试类型，各自关注不同的验证目标（不是严格的流水线层级）：

``text
Unit Test                    ← 业务规则、权限、Capability 匹配、错误映射
    ↓
Integration Test             ← API → Application → Repository 完整链路
    ↓
Device Adapter Contract Test ← 所有 Adapter 满足 Gateway Contract
    ↓
E2E Test                     ← 完整业务流程验证
    ↓
Agent Tool Test              ← Agent 只能通过允许的 Tool 操作系统
``

### 1. Unit Test（单元测试）

**测试对象**：

- 领域模型（Domain Model）的行为
- 业务规则和判断逻辑
- Capability 匹配算法
- 错误映射和状态流转
- 工具函数

**测试位置**：各模块 `domain/` 和 `application/` 包内的 `*_test.go` 文件

**关键原则**：

- 不依赖外部数据库（使用 mock、内存数据或 SQLite）
- 不依赖外部服务
- 不依赖 HTTP 服务器
- 不依赖设备连接
- 执行速度快（每个测试 <10ms）
- 使用 table-driven test 模式覆盖多种输入场景

**示例**：

``text
- TestSetVolume_ValidParameters
- TestSetVolume_InvalidVolumeLevel
- TestSetVolume_DeviceNotSupportVolume
- TestPermissionCheck_AdminCanRestart
- TestPermissionCheck_RegularUserCannotRestart
- TestCapabilityMatch_FilterDevicesByCapability
- TestOperationStatus_AggregateFromExecutions
``

### 2. Integration Test（集成测试）

**测试对象**：

- HTTP API → Handler → Application Service → Repository → 数据库 完整链路
- 数据库事务和并发行为
- API 参数校验和错误响应
- 模块间协作

**测试位置**：`tests/integration/` 目录

**关键原则**：

- 使用 MySQL 验证真实 Repository 行为（每个测试使用独立 schema 或 test containers）
- 启动真实 HTTP 服务器（`httptest.Server`）
- 不依赖外部服务或真实设备（设备层使用 Mock Gateway）
- 使用 test fixtures 准备测试数据
- 测试前后数据库状态可验证
- Integration Test 是验证 Repository 行为与生产一致性的主要手段

**示例**：

``text
- TestCreateDevice_API
- TestSetVolume_API_ValidRequest
- TestSetVolume_API_DeviceNotFound
- TestSetVolume_API_Unauthorized
- TestListDevices_API_Pagination
- TestOperationCreation_Integration
``

### 3. Device Adapter Contract Test（设备适配器契约测试）

**测试对象**：

- 所有 Device Adapter 是否满足 Device Gateway 定义的接口契约
- Adapter 的协议转换逻辑是否正确
- Adapter 的错误映射是否符合规范

**测试位置**：`tests/contract/` 目录

**关键原则**：

- 定义统一的 Contract Test Suite（一组标准测试用例）
- 所有 Adapter 必须通过相同的 Contract Test Suite
- 使用 Mock Transport 或预定义的协议响应
- 新 Adapter 合并前必须通过 Contract Test
- Contract Test 覆盖 Gateway 定义的所有方法
- 不允许使用 Simulator 替代 Contract Test

**示例**：

``text
- ContractTest_SetVolume_Success
- ContractTest_SetVolume_DeviceNotSupport
- ContractTest_GetStatus_Online
- ContractTest_GetStatus_Offline
- ContractTest_GetStatus_Timeout
- ContractTest_Restart_Success
- ContractTest_Restart_DeviceBusy
``

### 4. E2E Test（端到端测试）

**测试对象**：

- 完整的业务流程（从 API 请求到设备操作完成）
- 前后端协作（可选，第一阶段可仅测试后端）
- 多步骤操作链路

**测试位置**：`tests/e2e/` 目录

**关键原则**：

- 使用 Device Simulator 替代真实设备
- 使用完整数据库（MySQL，与生产环境一致）
- 测试完整业务流程（创建设备 → 配置 → 操作 → 查询结果）
- 可以使用 API 客户端模拟用户操作
- 执行速度较慢，只覆盖关键业务场景

**示例**：

``text
- TestE2E_CreateDeviceAndSetVolume
- TestE2E_BatchOperationWithFailures
- TestE2E_DeviceRestartAndHealthCheck
- TestE2E_OperationAuditTrail
``

### 5. Agent Tool Test（Agent Tool 测试）

**测试对象**：

- Tool 的参数校验是否正确
- Tool 的权限检查是否生效
- Tool 的设备能力确认是否准确
- Tool 是否只调用 Application Service（不绕过）
- Agent 是否只能通过 Tool 访问系统

**测试位置**：`tests/agent/` 目录

**关键原则**：

- 每个 Tool 的输入校验测试（正常输入、边界输入、恶意输入）
- 每个 Tool 的权限测试（有权限、无权限、部分权限）
- 每个 Tool 的能力确认测试（设备支持、设备不支持）
- 验证 Tool 返回值结构正确
- 验证 Tool 错误映射正确
- 验证 Agent 无法直接调用 Application Service
- 验证纵深防御：Tool 校验通过但 Application Service 独立拒绝的场景

**示例**：

``text
- TestTool_SetVolume_ValidInput
- TestTool_SetVolume_InvalidVolumeLevel
- TestTool_SetVolume_NoPermission
- TestTool_SetVolume_DeviceNotSupportVolume
- TestTool_ListDevices_FilterByArea
- TestAgent_CannotCallApplicationServiceDirectly
``

### 测试基础设施

| 组件 | 说明 | 位置 |
|------|------|------|
| Test Fixtures | 预定义的测试数据 | `tests/fixtures/` |
| Test Helpers | 通用测试工具函数 | `tests/helpers/` |
| Mock Transport | 模拟网络传输 | `tests/mocks/transport/` |
| Mock Gateway | 模拟 Device Gateway | `tests/mocks/gateway/` |
| Device Simulator | 模拟设备协议响应 | `tests/simulator/` |
| Test Database | 测试数据库工具（SQLite 用于 Unit/Agent Test，MySQL 用于 Integration/E2E） | `tests/helpers/database/` |

### 测试规则

1. **确定性**：所有测试必须可重复执行，不依赖随机值、当前时间或外部状态
2. **隔离性**：每个测试独立，不依赖其他测试的执行顺序
3. **快速反馈**：Unit Test 应在秒级完成，Integration Test 应在分钟级完成
4. **无外部依赖**：测试不依赖真实设备、外部服务或网络
5. **CI 集成**：所有测试类型在 CI 中自动运行
6. **不跳过失败**：不允许跳过失败测试，必须修复或标记为 BLOCKED
7. **不删除测试**：不允许通过删除测试来规避失败

### 测试优先级

当测试资源有限时，按以下优先级编写测试：

1. **Unit Test**（最高优先级）：覆盖核心业务规则和错误处理
2. **Contract Test**：覆盖所有 Adapter 的 Gateway 契约
3. **Integration Test**：覆盖关键 API 行为
4. **Agent Tool Test**：覆盖所有 Tool 的安全边界
5. **E2E Test**（最低优先级）：覆盖关键业务流程

## Alternatives Considered

### 1. 传统三层测试（Unit / Integration / E2E）

- **优势**：模型简单，团队熟悉
- **劣势**：缺少设备 Adapter 的专门测试，Adapter 质量依赖集成测试
- **否决原因**：设备协议适配是本项目核心挑战，需要专门的测试层级

### 2. 只使用 E2E 测试

- **优势**：最接近真实场景
- **劣势**：执行慢、调试难、覆盖不完整
- **否决原因**：无法满足快速反馈和精确验证的需求

### 3. 使用 BDD 框架（如 Cucumber）

- **优势**：测试用例可读性高，非技术人员可参与
- **劣势**：Go 生态中 BDD 工具不成熟，增加框架依赖
- **否决原因**：第一阶段不需要，table-driven test 已足够

## Consequences

### 正面

- 五类测试覆盖了广播设备管理的所有关键验证目标
- Device Adapter Contract Test 确保协议适配的正确性
- Agent Tool Test 确保 AI Agent 的操作安全边界
- 各类测试聚焦各自的验证目标，互不依赖
- 测试基础设施使编写新测试更简单

### 负面

- 五类测试比传统三类复杂，需要更多测试代码
- 需要维护 Contract Test Suite 和测试基础设施
- CI 执行时间较长（五类测试全部运行）
- 团队需要理解每类测试的验证目标和边界

### CI 优化策略

- 每次提交运行 Unit Test 和 Contract Test
- 每次 PR 运行 Unit Test + Integration Test + Contract Test
- 每次发布前运行全部五类测试
- 使用测试缓存减少未变更模块的重复测试

## Revisit Conditions

以下情况触发重新评估本决策：

1. **CI 执行时间过长**：当全部测试执行时间超过 15 分钟，影响开发效率时
2. **测试类型冗余**：当某类测试长期没有新增测试用例时
3. **新测试维度需求**：当需要引入性能测试、安全扫描或混沌测试时
4. **测试维护成本过高**：当测试代码的维护成本超过业务代码时
5. **团队扩展**：当团队增长到需要按模块分配测试职责时

## 验证方式

- 确认每个模块的 `domain/` 和 `application/` 有对应的 Unit Test
- 确认每个 Device Adapter 通过了 Contract Test Suite
- 确认关键 API 有 Integration Test 覆盖
- 确认每个 Agent Tool 有 Agent Tool Test 覆盖
- 确认关键业务流程有 E2E Test 覆盖
- 确认所有测试在 CI 中自动运行且全部通过
