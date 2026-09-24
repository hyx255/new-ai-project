# ADR-003：Database Strategy（数据库策略）

## 状态

Accepted

## 日期

2026-09-22

## Context

项目需要选择数据库策略，面临以下约束：

- 广播行业客户的生产环境普遍使用 MySQL 或 GreatDB（GreatDB 是 MySQL 兼容的国产数据库）
- 开发者本地环境不应依赖外部数据库服务，否则开发启动成本高
- 集成测试需要确定性的数据库环境，不能依赖远程数据库
- 生产环境数据需要事务完整性、并发控制和可靠持久化
- 部分客户可能使用 GreatDB，必须保持 MySQL 协议兼容性
- 未来可能存在缓存需求，但当前不确定是否需要

## Decision

### 数据库选型

| 环境 | 数据库 | 理由 |
|------|--------|------|
| 生产部署 | MySQL / GreatDB | 客户环境标配，运维成熟，生态完善 |
| 本地开发 | SQLite | 零外部依赖，快速启动 |
| Unit Test | SQLite / 内存数据库 | 快速执行，不依赖外部服务 |
| Integration Test | MySQL | 验证真实 Repository 行为与生产一致性 |
| GreatDB 兼容性验证 | GreatDB | CI 阶段验证客户环境兼容性 |
| 缓存 | Redis（条件性） | 仅在确有需求时引入，不作为默认组件 |

### 多数据库兼容性规则

**核心原则**：优先使用 MySQL / GreatDB 共同兼容的能力；SQLite 仅作为本地开发和快速测试便利，不以 SQLite 兼容性约束生产数据库设计。

1. **SQL 语法策略**：
   - 数据模型以 MySQL / GreatDB 能力为基准设计
   - 优先使用 MySQL 和 GreatDB 共同支持的 SQL 语法
   - 可以使用 MySQL/GreatDB 特有能力（如 JSON 列、ENUM），但需标记为 `MySQL/GreatDB-only`
   - 迁移文件保持 SQLite 下基本可执行（用于本地开发验证），但允许存在 MySQL/GreatDB 特有语法
2. **避免跨库不兼容特性**：
   - 不使用 MySQL 存储过程和触发器（应用层处理业务逻辑）
   - 避免依赖 MySQL 特有的字符串函数（使用应用层处理）
   - JSON 列类型可以使用，但应用层必须能处理 TEXT 回退（本地开发 SQLite 场景）
3. **数据类型映射**：
   - `INTEGER` → SQLite INTEGER / MySQL INT or BIGINT
   - `VARCHAR(n)` → SQLite TEXT / MySQL VARCHAR(n)
   - `TEXT` → SQLite TEXT / MySQL TEXT
   - `BOOLEAN` → SQLite INTEGER(0/1) / MySQL TINYINT(1)
   - `DATETIME` → SQLite TEXT (ISO 8601) / MySQL DATETIME
4. **迁移文件规范**：
   - 使用版本化 SQL 文件（如 `000001_create_users.up.sql`）
   - 每个迁移必须是幂等的（可重复执行或包含存在性检查）
   - 每个迁移只向前（up），同时提供回滚脚本（down）
   - 命名格式：`<version>_<description>.<up|down>.sql`

### ORM / 数据访问策略

- 第一阶段不强制选择 ORM
- 可选方案：GORM、sqlx、原生 `database/sql`
- 无论选择哪种方式，Repository 模式必须封装数据访问逻辑
- 业务层（`domain/` 和 `application/`）不得直接 import 数据库驱动或 ORM 包
- 数据访问代码位于 `infrastructure/` 层

### Redis 引入条件

Redis 不作为默认组件。仅在以下条件满足时考虑引入：

1. 存在明确的缓存需求（高频读取、低频写入的数据）
2. 需要分布式限流或分布式锁
3. 需要会话管理且无法使用数据库方案
4. 经过 ADR 评估并获得批准

引入 Redis 时必须创建新的 ADR，记录使用场景、数据丢失容忍度和回退方案。

## Alternatives Considered

### 1. 全部使用 PostgreSQL

- **优势**：功能强大，JSON 支持好，扩展性强
- **劣势**：客户环境普遍使用 MySQL/GreatDB，引入 PostgreSQL 增加部署复杂度
- **否决原因**：增加客户运维成本，不符合客户技术栈

### 2. 只使用 MySQL（包括开发和测试）

- **优势**：环境一致性最高
- **劣势**：开发者必须安装和配置 MySQL，测试需要清理数据库状态
- **否决原因**：增加本地开发门槛，集成测试环境隔离困难

### 3. 使用内存数据库（如纯内存 SQLite）进行测试

- **优势**：测试速度最快
- **劣势**：不能测试磁盘 I/O 相关行为，与生产差异更大
- **否决原因**：文件级 SQLite 测试已足够快且更接近真实行为

### 4. 使用 MongoDB 等 NoSQL 数据库

- **优势**：灵活的 Schema，适合非结构化数据
- **劣势**：广播设备管理的数据模型以关系型为主（设备-区域-用户-操作-审计），关系型数据库更自然
- **否决原因**：数据模型不匹配

## Consequences

### 正面

- 本地开发零数据库依赖，`go run` 即可启动
- Integration Test 使用真实 MySQL，Repository 行为得到生产级验证
- GreatDB 兼容性在 CI 阶段得到验证
- 生产环境与客户需求一致（MySQL/GreatDB）
- 迁移文件是纯 SQL，不依赖特定工具

### 负面

- SQLite 和 MySQL 的类型系统存在差异，本地开发时部分 MySQL 行为无法完全模拟
- CI 需要配置 MySQL 和 GreatDB 环境
- 开发者需要了解多数据库的差异

### 风险控制

- Integration Test 在 MySQL 上进行，确保 Repository 行为与生产一致
- GreatDB 兼容性验证在 CI 阶段执行
- 代码中使用参数化查询，避免 SQL 方言差异
- MySQL/GreatDB-only 特性需在代码和迁移文件中明确标记

## Revisit Conditions

以下情况触发重新评估本决策：

1. **客户环境变化**：客户生产环境从 MySQL 迁移到其他数据库
2. **SQL 兼容性成本过高**：MySQL / GreatDB / SQLite 的兼容性维护成本超过收益
3. **数据规模**：数据量增长到 SQLite 在本地开发中无法承载（>10GB）
4. **复杂查询需求**：需要 PostgreSQL 级别的高级查询能力（如复杂窗口函数、CTE）
5. **JSON 数据需求**：需要大量使用 JSON 列类型和 JSON 查询
6. **Redis 需求出现**：当确实需要缓存或分布式锁时，创建独立 ADR 评估

## 验证方式

- 确认本地开发可以使用 SQLite 启动和运行
- 确认 Unit Test 使用 SQLite 或内存数据库
- 确认 Integration Test 在 MySQL 上运行，验证真实 Repository 行为
- 确认 CI 阶段在 MySQL 和 GreatDB 上验证所有迁移文件
- 确认 MySQL/GreatDB-only 特性在代码和迁移文件中有明确标记
- 确认 `domain/` 和 `application/` 不 import 数据库驱动包
