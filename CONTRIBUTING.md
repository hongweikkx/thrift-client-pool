# 贡献指南

感谢你对 thrift-client-pool 项目的兴趣！

## 如何贡献

### 报告 Bug

如果发现 bug，请通过 GitHub Issues 报告。请提供：
- 问题的详细描述
- 复现步骤
- 预期行为
- 实际行为
- 环境信息（Go 版本、操作系统等）

### 提交改进

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'feat: add some amazing feature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码风格

- 遵循 Go 的官方代码风格指南
- 运行 `go fmt` 格式化代码
- 使用 `go vet` 检查代码
- 为所有导出的函数添加文档注释

### 测试

- 为新功能添加测试
- 确保所有测试通过：`go test -v ./...`
- 保持测试覆盖率

### 提交消息格式

遵循约定式提交（Conventional Commits）：

```
feat: 新增功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整（不影响功能）
refactor: 重构代码
perf: 性能优化
test: 测试相关
chore: 构建、依赖等
```

## 开发流程

1. 克隆项目并创建特性分支
2. 在 `pool.go` 和 `example/` 中进行更改
3. 运行测试：`go test -v ./...`
4. 验证编译：`go build ./...`
5. 提交更改并推送
6. 提交 Pull Request

## 许可证

你的贡献将按照项目的现有许可证（MIT）进行许可。

---

如有任何问题，请通过 GitHub Issues 联系我们。
