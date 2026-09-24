# Device Management Requirement

## 1. 目标

实现用户管理功能。

---

## 2. Device Type

字段：

- id
- name
- vendor
- model
- description
- created_at
- updated_at

规则：

1. name + vendor + model 唯一
2. 已经被 Device 使用的 DeviceType 不允许删除

接口：

- POST /api/device-types
- GET /api/device-types
- GET /api/device-types/:id
- PUT /api/device-types/:id
- DELETE /api/device-types/:id

---

## 3. Device

字段：

- id
- name
- device_type_id
- address
- status
- created_at
- updated_at

状态：

- REGISTERED
- ACTIVE
- OFFLINE
- DISABLED

规则：

1. 新建设备默认 REGISTERED
2. 每个设备必须属于一个 DeviceType
3. DISABLED 状态暂时不能执行设备操作
4. 删除 DeviceType 前必须检查是否存在设备

接口：

- POST /api/devices
- GET /api/devices
- GET /api/devices/:id
- PUT /api/devices/:id
- PATCH /api/devices/:id/status

---

## 4. 后端要求

遵循：

API
↓
Service
↓
Repository
↓
Database

业务 Model 放在 model 中。

API 不直接访问 Repository。

Service 不写 SQL。

Repository 不包含业务逻辑。

---

## 5. 前端

提供：

- DeviceType 列表页
- DeviceType 新增/编辑
- Device 列表页
- Device 新增/编辑
- Device 状态显示

---

## 6. 验收标准

1. 可以创建设备类型
2. 可以创建属于该类型的设备
3. 重复 DeviceType 被拒绝
4. DeviceType 被设备引用时不可删除
5. 新建设备默认 REGISTERED
6. 可以将设备设置为 DISABLED
7. 后端单元测试通过
8. API 测试通过
9. 前端可以完成基本操作