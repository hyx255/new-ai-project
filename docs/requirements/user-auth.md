# User Management & Authentication Requirement

## 1. 目标

实现基础用户管理与登录认证能力：

- 用户 CRUD
- JWT 登录认证
- 用户启用 / 禁用
- 用户逻辑删除
- 管理员重置密码
- 用户主动修改密码
- 首次登录强制修改密码
- 前端登录页与用户管理页

本阶段只做基础用户认证和管理，不实现完整 RBAC、数据权限、Refresh Token、OAuth、SSO、MFA。

---

## 2. 用户角色

### ADMIN
可以：
- 创建用户
- 查询用户列表 / 详情
- 修改用户
- 启用 / 禁用其他用户
- 重置其他用户密码
- 删除其他用户

限制：
- 不能删除自己
- 不能禁用自己

### USER
可以：
- 登录
- 查看自己的当前用户信息
- 修改自己的密码

不可以：
- 访问用户管理功能
- 创建 / 删除 / 修改其他用户
- 重置其他用户密码

---

## 3. User 模型

至少包含：

```text
id
username
password_hash
role
status
must_change_password
token_version
deleted_at
created_at
updated_at
```

状态：

```text
ACTIVE
DISABLED
```

逻辑删除：

```text
deleted_at = NULL      → 正常用户
deleted_at != NULL     → 已删除用户
```

已删除用户：
- 不允许登录
- 不允许继续访问系统
- 默认用户列表不显示
- 历史业务记录保留

---

## 4. 用户名规则

- username 必填
- username 全局唯一
- 已逻辑删除用户的 username 仍视为占用
- username 冲突返回 409

---

## 5. 密码规则

密码必须满足：

- 长度不少于 8 位
- 至少 1 个大写字母
- 至少 1 个小写字母
- 至少 1 个数字

密码不得明文存储，只保存 `password_hash`。

任何 API 都不得返回 `password_hash`。

---

## 6. 用户创建

仅 ADMIN 可以创建用户。

```http
POST /api/users
```

创建时至少提供：

```json
{
  "username": "user01",
  "initial_password": "Password123",
  "role": "USER"
}
```

创建后默认：

```text
status = ACTIVE
must_change_password = true
token_version = 1
deleted_at = NULL
```

不提供用户自主注册接口。

---

## 7. 登录

```http
POST /api/auth/login
```

Request：

```json
{
  "username": "admin",
  "password": "Password123"
}
```

登录必须校验：

1. 用户存在
2. 用户未被逻辑删除
3. 密码正确
4. 用户状态为 ACTIVE

用户不存在或密码错误时，对外统一提示：

```text
用户名或密码错误
```

登录成功返回：

- JWT Access Token
- 当前用户基础信息
- must_change_password

---

## 8. JWT

JWT 至少包含：

```text
sub
role
token_version
exp
```

要求：

- JWT 必须过期
- JWT Secret 从配置读取
- 不硬编码 Secret
- 本阶段不实现 Refresh Token

---

## 9. Token 失效

使用 `token_version` 让旧 JWT 立即失效。

认证时：

```text
JWT.token_version
        ↓
User.token_version
        ↓
一致 → 有效
不一致 → 401
```

以下操作必须执行：

```text
token_version + 1
```

包括：

- 用户修改密码
- 管理员重置密码
- 用户被禁用
- 用户被逻辑删除

---

## 10. 首次登录强制改密

新建用户：

```text
must_change_password = true
```

首次登录成功后仍返回 JWT，但只能访问：

```http
GET  /api/auth/me
POST /api/auth/change-password
```

访问其他业务接口返回 403。

修改密码成功后：

```text
password_hash = 新密码 Hash
must_change_password = false
token_version += 1
```

当前 JWT 立即失效。

前端：

```text
清除 Token
→ 跳转 Login
→ 使用新密码重新登录
```

---

## 11. 修改密码

```http
POST /api/auth/change-password
```

Request：

```json
{
  "current_password": "OldPassword123",
  "new_password": "NewPassword123"
}
```

必须校验：

- 当前密码正确
- 新密码符合密码规则

成功后：

- 更新 password_hash
- must_change_password = false
- token_version += 1
- 当前 JWT 失效
- 用户重新登录

---

## 12. 管理员重置密码

```http
POST /api/users/:id/reset-password
```

仅 ADMIN 可执行。

重置后：

```text
password_hash = 新密码 Hash
must_change_password = true
token_version += 1
```

该用户旧 JWT 全部失效，下次登录后必须再次修改密码。

---

## 13. 启用 / 禁用

```http
PATCH /api/users/:id/status
```

支持：

```text
ACTIVE → DISABLED
DISABLED → ACTIVE
```

规则：

- 仅 ADMIN 可执行
- 管理员不能禁用自己
- 禁用时 token_version += 1
- DISABLED 用户不能登录

---

## 14. 删除用户

```http
DELETE /api/users/:id
```

规则：

- 仅 ADMIN 可执行
- 逻辑删除
- 管理员不能删除自己
- 删除时：
  - deleted_at = 当前时间
  - token_version += 1
- 删除后旧 JWT 失效
- 历史数据保留

---

## 15. Authentication Middleware

公开接口：

```http
POST /api/auth/login
GET  /health
```

其他业务 API 默认需要认证。

Middleware 负责：

1. 解析 Bearer Token
2. 校验 JWT
3. 校验 exp
4. 加载当前 User
5. 检查 deleted_at
6. 检查 status
7. 校验 token_version
8. 写入 Current User Context

如果：

```text
must_change_password = true
```

仅允许：

```http
GET  /api/auth/me
POST /api/auth/change-password
```

---

## 16. Authorization

以下 API 仅 ADMIN 可访问：

```http
POST   /api/users
GET    /api/users
GET    /api/users/:id
PUT    /api/users/:id
DELETE /api/users/:id
PATCH  /api/users/:id/status
POST   /api/users/:id/reset-password
```

USER 访问返回 403。

前端隐藏按钮或菜单不是最终权限控制，后端必须再次校验。

---

## 17. API 汇总

### Auth

```http
POST /api/auth/login
GET  /api/auth/me
POST /api/auth/change-password
```

### User

```http
POST   /api/users
GET    /api/users
GET    /api/users/:id
PUT    /api/users/:id
DELETE /api/users/:id
PATCH  /api/users/:id/status
POST   /api/users/:id/reset-password
```

---

## 18. 前端

### Login
路由建议：

```text
/login
```

登录成功：

```text
must_change_password = true
→ /change-password

否则
→ /dashboard
```

### Change Password

```text
/change-password
```

修改成功：

```text
清除 Token
→ /login
```

### User Management

建议：

```text
/users
/users/create
/users/:id/edit
```

支持：

- 用户列表
- 新建
- 编辑
- 启用 / 禁用
- 重置密码
- 删除

普通 USER 不显示用户管理菜单。

---

## 19. 本阶段不实现

- 自助注册
- Refresh Token
- Token Blacklist
- Redis Session
- OAuth
- SSO
- LDAP
- MFA
- 完整 RBAC
- 菜单权限
- 数据权限
- 区域权限
- Device Scope
- Agent Permission

---

## 20. 验收标准

- ADMIN 可以创建 / 查询 / 修改 / 禁用 / 启用 / 重置密码 / 删除用户
- ADMIN 不能删除自己
- ADMIN 不能禁用自己
- USER 访问用户管理 API 返回 403
- username 重复返回 409
- 密码规则正确校验
- 正确账号密码可登录
- 错误密码 / 不存在用户登录失败
- DISABLED 用户不能登录
- 已删除用户不能登录
- JWT 过期返回 401
- token_version 不一致返回 401
- 新用户首次登录必须改密码
- 首次改密完成后旧 JWT 失效
- 重置密码后旧 JWT 失效
- 禁用后旧 JWT 失效
- 删除后旧 JWT 失效
- password_hash 永不返回客户端
- 日志不记录明文密码、完整 JWT、JWT Secret
- 前端 Login / Change Password / User List 可以正常使用
