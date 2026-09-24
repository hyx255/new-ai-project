# ADR-002：Device Abstraction Layer（设备抽象层）

## 状态

Accepted

## 日期

2026-09-22

## Context

广播设备管理平台需要管理多厂商、多协议的广播设备。面临以下约束：

- 不同厂商设备使用不同的通信协议（TCP 私有协议、UDP 广播协议、HTTP API、SNMP 等）
- 业务逻辑需要与具体协议解耦，否则每次新增厂商或协议变更都需要修改业务代码
- 未来可能新增设备类型和协议，系统必须支持低成本扩展
- AI Agent 需要基于设备能力（Capability）而非协议细节来操作设备
- 业务层开发者应关注"做什么"（设置音量），而非"怎么做"（发送什么字节）
- 测试时不能依赖真实设备，需要 Simulator 替代

## Decision

建立四层设备抽象架构，业务层禁止依赖具体设备协议：

``text
Device Business       ← 业务规则、权限、能力判断、操作编排、审计
        ↓
Device Gateway        ← 定义业务需要的设备能力接口（Go interface）
        ↓
Protocol Adapter      ← 将业务语义转换为具体厂商/协议的实现
        ↓
Transport             ← TCP/UDP/HTTP 连接管理、超时、重试、心跳
        ↓
Real Device / Simulator
``

### 各层职责

**Device Business（业务层）**：

- 负责设备业务规则（哪些用户可以操作哪些设备）
- 负责能力判断（目标设备是否支持所需操作）
- 负责操作编排（组合多个能力调用完成业务流程）
- 负责审计记录
- 使用 Device Gateway 接口操作设备，不关心底层协议

**Device Gateway（网关层）**：

- 定义设备能力的 Go interface（如 `VolumeController`, `PlaybackController`）
- 定义能力查询接口（`GetCapabilities(deviceID)`）
- 定义标准结果语义（成功、失败、不支持、超时）
- 不包含协议转换逻辑
- 是所有 Adapter 必须满足的契约

**Protocol Adapter（适配器层）**：

- 实现 Device Gateway 定义的 interface
- 负责业务语义到厂商/协议的映射
- 负责协议解析和二进制数据处理
- 负责厂商特有的命令构造和响应解析
- 不包含业务规则（如权限判断）

**Transport（传输层）**：

- 管理 TCP/UDP/HTTP 连接
- 负责超时控制、重试策略、心跳检测
- 负责连接池和连接生命周期管理
- 不包含业务语义，只做字节流传输

### 禁止出现在业务层的代码

以下代码严禁出现在 `internal/modules/*/domain/` 和 `internal/modules/*/application/` 中：

- TCP / UDP 报文构造与发送
- HTTP 私有协议 URL 和参数
- 厂商私有协议字段名
- 二进制协议解析逻辑
- 协议特有的错误码直接处理

### 业务语义表达规范

``text
✓ 正确：SetVolume(deviceID, 70)
✓ 正确：GetPlaybackStatus(deviceID)
✓ 正确：Restart(deviceID)

✗ 错误：SendTCPCommand(deviceID, []byte{0x01, 0x02, 0x03})
✗ 错误：http.Post("http://device/vendor/control", body)
✗ 错误：ParseVendorResponse(rawBytes)
``

### 扩展新协议/新厂商的路径

1. 在 Transport 层实现或复用网络通信组件
2. 实现新的 Protocol Adapter，满足 Device Gateway Contract
3. 通过 Adapter Contract Test（必须）
4. 注册到 Adapter Registry（运行时配置）

### Device Capability 集成

- Device Gateway 定义标准 Capability 枚举和查询接口
- 每个 Adapter 在初始化时声明自己支持哪些 Capability
- Device Business 在操作前调用 `GetCapabilities(deviceID)` 判断
- Agent Tool 基于 Capability 进行设备筛选和操作

## Alternatives Considered

### 1. 业务代码直接集成各协议 SDK

- **优势**：开发速度最快，无需抽象层
- **劣势**：每新增一个协议都要修改业务代码，测试困难，协议 bug 直接影响业务
- **否决原因**：严重违反开闭原则，无法支持多厂商扩展

### 2. 只使用 Protocol Adapter 不区分 Transport

- **优势**：层级更少，实现更简单
- **劣势**：网络通信逻辑（超时、重试、连接管理）与协议解析混杂，难以复用
- **否决原因**：Transport 层可以被多个 Adapter 复用（如多个基于 TCP 的协议共享连接管理），拆分后更易测试和维护

### 3. 使用 Plugin Architecture 动态加载协议

- **优势**：协议可以热插拔，无需重编译
- **劣势**：Go 语言原生不支持动态插件，需要 CGO 或进程间通信，增加复杂性
- **否决原因**：第一阶段不需要热插拔，注册表 + 配置切换已足够灵活

## Consequences

### 正面

- 业务代码完全不依赖具体协议，新增厂商只需新增 Adapter
- 业务层单元测试可以 mock Device Gateway 接口
- 每个 Adapter 可独立进行 Contract Test
- Transport 层可复用（如多个 TCP 协议共享连接管理器）
- Agent Tool 可以基于 Capability 进行设备操作，无需了解协议

### 负面

- 每次设备交互多一层抽象开销（性能影响极小，可忽略）
- 开发者需要理解四层边界，学习成本略高
- 简单的单协议场景也需走完整链路

### 架构边界维护

- Code Review 检查业务层是否出现协议相关代码
- Adapter Contract Test 确保所有 Adapter 满足 Gateway Contract
- 静态分析检查 `domain/` 和 `application/` 是否 import `adapter/` 或 `transport/` 包

## Revisit Conditions

以下情况触发重新评估本决策：

1. **Adapter 数量过多**：当 Adapter 超过 20 个，且存在大量公共协议处理逻辑时，考虑提取 Protocol Framework
2. **性能瓶颈**：当抽象层的性能开销在高频设备通信场景中成为瓶颈时
3. **Capability 碎片化**：当 Capability 定义过于碎片化，导致组合爆炸时
4. **协议变更频繁**：当厂商协议频繁变更导致 Adapter 维护成本过高时
5. **实时性要求极高**：当设备操作的延迟要求极低（<1ms），抽象层开销不可接受时

## 验证方式

- 确认 `domain/` 和 `application/` 包不 import `adapter/` 或 `transport/` 包
- 确认业务代码中不存在 `net.Dial`, `http.Post` 等直接网络调用
- 确认所有 Adapter 实现了 Device Gateway 定义的 interface
- 确认 Adapter Contract Test 覆盖了 Gateway 定义的所有方法
