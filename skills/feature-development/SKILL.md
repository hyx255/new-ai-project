# Skill: feature-development

## 名称

feature-development

## 使用时机

当需要开发新功能或实现新需求时使用。包括：
- 新业务模块开发
- 现有模块功能扩展
- API 端点实现
- 数据库表/字段新增
- 前端页面开发

## 输入

- 需求文档（`docs/requirements/`）
- 架构约束（`ARCHITECTURE.md`, `docs/decisions/`）
- 当前代码状态
- 验收标准

## 前置条件

- 需求清晰，无阻塞性歧义
- 已阅读相关架构决策（ADR）
- 已确认范围边界（IN/OUT）

## 执行步骤

1. **理解需求**
   - 阅读需求文档，提取验收标准
   - 识别范围、约束、依赖
   - 如有歧义或冲突 → STOP，询问用户

2. **设计**
   - 确定模块归属（新建/扩展）
   - 设计 Domain Model（Entity/Value Object）
   - 设计 API 契约（Request/Response）
   - 设计数据库 Schema（如需）
   - 检查与 ARCHITECTURE.md/ADR 的一致性

3. **实现**
   - 按模块分层实现：Domain → Application → Infrastructure → Interfaces
   - 同步编写测试（Unit + Integration）
   - 保持小步提交，每步可验证

4. **验证**
   - 运行测试：`go test ./...` + `npm run test`（前端）
   - 手动验证关键路径（如需要）
   - 确认满足验收标准

5. **文档更新**
   - 更新相关文档（如有变更）
   - 记录关键决策（如需新增 ADR）

## 必须遵守

- 遵循 Modular Monolith 架构
- 遵循 Defense in Depth 原则
- 遵循 Device Abstraction 四层结构
- Operation/Execution/Audit 模型
- 不引入未批准的技术依赖
- 不扩大需求范围

## 禁止

- 不臆造需求或业务行为
- 不修改已批准的架构（除非通过 ADR）
- 不跳过测试直接实现
- 不在没有设计的情况下直接编码
- 不引入微服务/消息队列/分布式事务（除非有明确 ADR）

## 验证

- 所有测试通过（Unit + Integration）
- 满足验收标准
- 代码符合架构约束
- 无新增技术债务（或已记录）

## 输出

- 可工作的功能代码
- 测试代码
- 更新的文档（如有）
- 验证结果

## 失败处理

- 需求歧义 → STOP，列出问题，询问用户
- 架构冲突 → STOP，报告冲突，不自行修改架构
- 技术决策缺失 → STOP，建议 ADR，等待确认
- 范围扩大 → PAUSE，请求确认

## 必须停止并询问用户

- 需求不完整或有歧义
- 需要新增业务模块（超出已批准范围）
- 需要修改公开 API 契约
- 需要修改数据模型或迁移
- 与已记录的 ADR 冲突
- 需要选择新技术栈/数据库/部署平台