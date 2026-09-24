# Auth Development Skill

## 1. Purpose

用于指导：

- User Authentication
- Login
- JWT
- Password
- Change Password
- Reset Password
- Authentication Middleware
- User Status
- Token Invalidation
- Basic Authorization

职责：

```text
Requirement      → 定义做什么
Skill            → 定义怎么做
AGENTS.md        → 项目全局规则
ARCHITECTURE.md  → 代码结构和依赖方向
```

---

## 2. 开发前必须读取

开始认证相关开发前必须读取：

1. `AGENTS.md`
2. `ARCHITECTURE.md`
3. 当前 User/Auth Requirement
4. 当前 Error / Response 规范
5. 当前 Database / Migration 机制
6. 当前 Config
7. 当前 Frontend API Client / Router

Requirement 未要求的能力，不提前实现。

---

## 3. 架构约束

必须遵循：

```text
API
↓
Service
↓
Repository
↓
Database
```

Token：

```text
AuthService
↓
TokenProvider
```

密码：

```text
Auth/User Service
↓
Password Hasher / Password Policy
```

禁止：

- Handler 直接访问 Repository
- Handler 写 SQL
- Repository 处理 JWT
- Repository 比较密码
- Repository 做权限判断

---

## 4. Handler

Handler 只负责：

- 解析 Request
- 基础格式校验
- 调用 Service
- Error Mapping
- Response

禁止：

- 比较密码
- Hash 密码
- 生成 JWT
- 修改 token_version
- 实现复杂权限
- 直接访问 Repository

---

## 5. Service

Service 负责：

- Login
- 用户创建 / 修改
- Enable / Disable
- Soft Delete
- Change Password
- Reset Password
- Password Policy
- Password Compare
- token_version 更新
- 用户状态判断
- self delete / self disable 判断
- 基础角色授权

---

## 6. Repository

Repository 只负责持久化：

- Create
- GetByID
- GetByUsername
- List
- Update
- UpdatePassword
- UpdateStatus
- IncrementTokenVersion
- SoftDelete

禁止：

- Password Compare
- Password Policy
- JWT Generate / Parse
- Authorization
- Login Flow

---

## 7. Repository Error 约定

统一：

```text
找到：
entity, nil

不存在：
nil, repository.ErrNotFound

数据库错误：
nil, wrapped error
```

如果 error 可能被 `%w` 包装，上层必须：

```go
errors.Is(err, repository.ErrNotFound)
```

禁止直接使用：

```go
err == repository.ErrNotFound
```

---

## 8. Password

必须使用成熟密码哈希算法，优先使用：

```text
bcrypt
```

禁止：

- 明文密码
- MD5
- SHA1
- SHA256(password)
- Base64
- 自定义可逆加密

数据库只保存：

```text
password_hash
```

API Response 禁止返回：

```text
password_hash
```

---

## 9. Password Policy

密码规则集中实现，例如：

```text
PasswordPolicy.Validate(password)
```

禁止在 Handler / Service / Repository 多处复制规则。

前端可以做同样的即时提示，但后端必须最终校验。

---

## 10. JWT

JWT 必须：

- 有 exp
- Secret 配置化
- 校验签名
- 包含 sub
- 包含 token_version

禁止：

- Secret 硬编码
- 永不过期 JWT
- JWT 中包含密码或 password_hash

JWT 生成 / 解析 / 校验集中在 TokenProvider。

---

## 11. Token Invalidation

Requirement 要求旧 JWT 立即失效时，优先使用：

```text
token_version
```

签发：

```text
User.token_version
→ JWT.token_version
```

认证：

```text
JWT.token_version
→ User.token_version
→ Compare
```

不一致：

```text
401
```

不要为了 MVP 引入：

- Token Blacklist
- Redis Token Store
- Refresh Token Rotation

除非 Requirement 明确要求。

---

## 12. Authentication Middleware

Middleware 只负责：

- 获取 Authorization Header
- 解析 Bearer Token
- 校验 JWT
- 校验 exp
- 加载 Current User
- 校验 deleted_at
- 校验 status
- 校验 token_version
- 写入 Current User Context

禁止在 Middleware 中：

- 创建 / 删除用户
- 修改 / 重置密码
- 写 SQL
- 实现复杂业务流程

---

## 13. Authentication 与 Authorization 分离

```text
Authentication
→ 你是谁

Authorization
→ 你能不能做这件事
```

Middleware 负责认证。

业务权限由 Service 或项目现有 Authorization 层负责。

不能把全部权限逻辑塞进 JWT Middleware。

---

## 14. Current User Context

Context 中只放最小必要信息，例如：

```text
user_id
username
role
must_change_password
```

禁止放：

- password_hash
- 明文密码
- 完整 JWT

---

## 15. 首次登录强制改密

如果 Requirement 使用：

```text
must_change_password = true
```

后端必须限制业务 API。

禁止只依赖前端跳转。

只有 Requirement 指定的接口允许访问。

---

## 16. Change Password

标准流程：

```text
Load Current User
↓
Verify Current Password
↓
Validate New Password
↓
Hash New Password
↓
Update password_hash
↓
Update must_change_password
↓
Increment token_version
```

---

## 17. Reset Password

管理员 Reset 与用户 Change Password 分开实现。

Reset：

- 不需要旧密码
- 必须校验管理员权限
- 新密码必须符合 Password Policy
- 根据 Requirement 更新 must_change_password
- 根据 Requirement 更新 token_version

禁止提供“查询旧密码”。

---

## 18. Soft Delete

逻辑删除必须经过 Service。

Service 负责：

- 权限校验
- self delete 校验
- token_version 更新
- 业务状态处理

Repository 只负责更新数据库。

普通查询默认：

```text
deleted_at IS NULL
```

---

## 19. 登录错误安全

客户端禁止区分：

```text
用户不存在
密码错误
```

统一为类似：

```text
invalid username or password
```

日志也禁止记录密码、password_hash、完整 Token。

---

## 20. HTTP 语义

遵循项目现有 Error/Response 规范。

一般：

```text
400 → 参数非法 / 密码规则不满足
401 → 未认证 / Token 无效 / Token 过期 / Token 已失效
403 → 已认证但无权限
404 → 目标不存在
409 → username 冲突 / 状态冲突 / self delete / self disable
```

---

## 21. Frontend

认证相关页面必须复用统一 API Client。

API Client 统一处理：

```text
Authorization: Bearer <token>
```

禁止页面自己拼 Header。

统一处理：

```text
401
→ 清理 Token
→ 清理当前用户
→ 跳转 Login
```

403 不应自动登出，只提示无权限。

Route Guard 只能用于 UX，不能替代后端权限校验。

---

## 22. Logging

禁止记录：

- password
- current_password
- new_password
- password_hash
- 完整 JWT
- 完整 Authorization Header
- JWT Secret

允许记录：

- user_id
- username
- role
- operation
- safe error category
- trace_id

---

## 23. Migration

User/Auth Schema 必须通过：

```text
migrations/
```

管理。

禁止：

- bootstrap 执行 CREATE TABLE
- bootstrap 执行 ALTER TABLE
- Server 启动自动修改 Schema

沿用项目已有 `cmd/migrate`。

---

## 24. 默认管理员初始化

如果系统需要初始 ADMIN，但 Requirement 没有定义初始化方式：

必须先提出问题。

禁止自行：

- 硬编码 admin/admin
- 在 server 启动时偷偷创建管理员
- 写死默认密码

---

## 25. 测试要求

至少覆盖：

### Login
- 正常登录
- 用户不存在
- 密码错误
- Disabled User
- Deleted User

### Password
- 少于 8 位
- 缺少大写
- 缺少小写
- 缺少数字
- Change Password
- Reset Password

### JWT
- Missing Token
- Invalid Token
- Expired Token
- token_version mismatch
- Disable 后旧 Token 失效
- Delete 后旧 Token 失效
- Reset Password 后旧 Token 失效

### Authorization
- ADMIN 正常访问管理接口
- USER 访问管理接口 → 403
- self delete 拒绝
- self disable 拒绝

### First Login
- must_change_password=true
- me 可访问
- change-password 可访问
- 其他业务 API 被拒绝
- 改密后旧 JWT 失效
- 新密码重新登录成功

### Security
- password_hash 不返回
- 日志不包含密码
- 日志不包含完整 JWT

---

## 26. 完成验证

后端至少运行：

```bash
go test ./...
go vet ./...
go build ./...
```

前端至少运行：

```bash
npx vue-tsc --noEmit
npm run build
```

并进行真实 HTTP 联调。

---

## 27. 编码前必须输出

在开始编码前输出：

1. Requirement 理解
2. User 数据模型
3. Login Flow
4. JWT / token_version Flow
5. First Login Flow
6. AuthService / UserService 边界
7. Repository 设计
8. Middleware 设计
9. API 设计
10. Frontend 设计
11. Test Plan
12. Security Checkpoints
13. 预计修改文件
14. 当前歧义
15. Scope 外内容

存在重大歧义时先确认，不自行扩展。

---

## 28. 完成后必须输出

至少包括：

1. Implementation Summary
2. Added / Modified Files
3. Database Schema
4. APIs
5. Login Flow
6. JWT / token_version
7. First Login Flow
8. Authorization
9. Frontend Pages
10. Test Results
11. Runtime Verification
12. Security Review
13. Skill Compliance
14. Requirement Deviations
15. Known Issues / TBD

---

## 29. Forbidden

禁止：

- 明文密码落库
- password_hash 返回客户端
- JWT Secret 硬编码
- 永不过期 JWT
- Handler 比较密码
- Handler 生成 JWT
- Handler 直接访问 Repository
- Repository 做 Password Policy
- Repository 处理 JWT
- Middleware 承担复杂业务
- 仅依靠前端权限
- bootstrap 自动 Migration
- 日志记录密码或完整 Token
- 未经 Requirement 要求提前实现 Refresh Token
- 提前实现 OAuth / SSO / MFA
- 提前实现 Token Blacklist
- 提前实现 Redis Session
- 提前实现完整 RBAC

---

## 30. Core Principle

```text
Requirement 决定做什么
Skill 决定怎么做
Architecture 决定代码结构
AGENTS 决定项目规则
```

以及：

```text
当前 Requirement 没有要求的能力，不提前实现。
```
