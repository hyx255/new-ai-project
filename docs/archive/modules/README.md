# 模块文档

本目录将在真实业务模块出现后，用于存放各模块文档。

在模块具备已批准的需求和范围前，不要创建模块目录。

每个模块建议使用以下结构：

```text
docs/modules/<module-name>/
|-- spec.md
|-- design.md
|-- api.md
`-- acceptance.md
```

创建模块文档时，应从本目录的模板复制内容：

- `spec-template.md`
- `design-template.md`
- `api-template.md`
- `acceptance-template.md`

## 文件标准

- `spec.md`：模块目标、范围、用户故事、功能需求、非功能需求和不在范围内的事项。
- `design.md`：实现设计、边界、依赖、数据流、错误处理和风险。
- `api.md`：API 契约、请求/响应示例、校验、错误和兼容性说明。
- `acceptance.md`：验收标准、测试场景、人工验证说明和发布检查。

模块文档必须与产品需求和架构决策保持一致。

如果产品需求、业务规则、架构决策或验收标准存在未解决问题，不得自行猜测，任务必须进入 `BLOCKED`。
