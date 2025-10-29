# 贡献指南

感谢您对 cc-switch 项目的关注！本文档提供了参与项目贡献的指南和说明。

## 目录

- [行为准则](#行为准则)
- [快速开始](#快速开始)
- [开发环境配置](#开发环境配置)
- [项目结构](#项目结构)
- [开发流程](#开发流程)
- [代码风格指南](#代码风格指南)
- [测试](#测试)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)
- [问题反馈](#问题反馈)
- [发布流程](#发布流程)

## 行为准则

参与本项目的所有贡献者都应遵守以下行为准则：

- 尊重他人，包容不同观点
- 欢迎新人并帮助他们上手
- 专注于建设性的批评
- 优雅地接受反馈
- 以社区的最佳利益为重

## 快速开始

### 环境要求

在开始之前，请确保已安装以下工具：

- **Go**：1.23.0 或更高版本
- **Node.js**：14.0.0 或更高版本
- **Git**：用于版本控制
- **文本编辑器**：VS Code、GoLand 或您喜欢的编辑器

### 首次配置

1. **Fork 仓库**到您的 GitHub 账号
2. **克隆到本地**：
   ```bash
   git clone https://github.com/YOUR_USERNAME/cc-switch.git
   cd cc-switch
   ```

3. **添加上游远程仓库**：
   ```bash
   git remote add upstream https://github.com/HoBeedzc/cc-switch.git
   ```

4. **安装依赖**：
   ```bash
   go mod download
   go mod tidy
   ```

5. **验证配置**：
   ```bash
   go run . --help
   ```

## 开发环境配置

### 构建项目

```bash
# 构建所有平台
npm run build

# 仅构建 Go 二进制文件
npm run build:go

# 清理构建产物
npm run clean
```

### 本地运行

```bash
# 不构建直接运行
go run . --help
go run . list
go run . web

# 运行特定命令进行测试
go run . test --current
```

### 开发脚本

可用的 npm 脚本：

- `npm run dev`：开发模式运行应用
- `npm test`：运行 Go 测试
- `npm run fmt`：格式化 Go 代码
- `npm run vet`：运行 Go vet 静态分析
- `npm run clean`：清理构建产物

## 项目结构

了解项目结构有助于您浏览代码库：

```
cc-switch/
├── cmd/                    # CLI 命令实现
│   ├── root.go            # 根命令和 CLI 设置
│   ├── list.go            # 列出配置
│   ├── new.go             # 创建新配置
│   ├── use.go             # 切换配置
│   ├── test.go            # 测试 API 连接
│   ├── web.go             # Web 界面服务器
│   └── ...                # 其他命令
├── internal/
│   ├── config/            # 配置管理
│   ├── handler/           # 业务逻辑处理器
│   ├── ui/                # 用户界面层
│   ├── web/               # Web 界面
│   ├── export/            # 导出功能
│   ├── import/            # 导入功能
│   └── common/            # 共享工具
├── scripts/               # 构建和安装脚本
├── .github/workflows/     # CI/CD 工作流
└── npm/                   # npm 包分发
```

### 核心包

- **cmd/**：基于 Cobra 的 CLI 命令
- **internal/config/**：核心配置管理逻辑
- **internal/handler/**：业务逻辑和 API 测试
- **internal/ui/**：交互式提示和 CLI 界面
- **internal/web/**：Web 界面服务器和处理器
- **internal/common/**：工具函数（压缩、加密、版本）

## 开发流程

### 创建功能分支

```bash
# 更新主分支
git checkout main
git pull upstream main

# 创建功能分支
git checkout -b feature/your-feature-name
```

### 进行修改

1. **编写代码**，遵循下面的代码风格指南
2. **彻底测试**您的修改
3. **格式化代码**：运行 `npm run fmt`
4. **运行静态分析**：运行 `npm run vet`
5. **运行测试**：运行 `npm test`

### 保持 Fork 同步

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

## 代码风格指南

### Go 代码风格

遵循标准 Go 约定和最佳实践：

#### 格式化

- 使用 `gofmt` 自动格式化（运行 `npm run fmt`）
- 使用制表符缩进（gofmt 默认）
- 行长度通常不应超过 100 个字符

#### 命名约定

- **包名**：简短、小写、单个单词
- **文件名**：小写带下划线（如 `config_handler.go`）
- **函数**：CamelCase（导出）或 camelCase（非导出）
- **变量**：CamelCase 或 camelCase，使用描述性名称
- **常量**：CamelCase 或 UPPER_CASE（导出常量）

#### 代码组织

- 将相关功能组织在同一文件中
- 保持函数专注和单一职责
- 在适当的地方使用接口抽象
- 为所有导出的函数、类型和常量添加文档

#### 错误处理

- 始终检查和处理错误
- 提供带上下文的有意义错误消息
- 使用 `fmt.Errorf` 包装错误
- 除非在真正异常的情况下，否则不要使用 panic

示例：
```go
if err := someOperation(); err != nil {
    return fmt.Errorf("执行操作失败: %w", err)
}
```

#### 注释

- 使用完整句子，带适当标点
- 为所有导出的符号添加文档
- 在 `doc.go` 或包文件中添加包级文档
- 实现注释应解释"为什么"而不是"是什么"

示例：
```go
// ConfigManager 处理所有与配置相关的操作，包括
// 加载、保存和在不同的 Claude Code 配置文件之间切换。
type ConfigManager struct {
    // ...
}
```

### Web 界面指南

#### CSS

- 使用语义化的类名
- 适当时遵循 BEM 方法论
- 保持一致的间距和缩进
- 将相关样式分组

#### JavaScript

- 使用现代 ES6+ 特性
- 使用 try-catch 优雅处理错误
- 为所有操作提供用户反馈
- 保持函数小而专注

## 测试

### 编写测试

我们追求全面的测试覆盖。添加新功能时：

1. **创建测试文件**，在同一包中命名为 `*_test.go`
2. **编写表驱动测试**（在适用的情况下）
3. **测试边界情况**和错误条件
4. **模拟外部依赖**（必要时）

测试结构示例：

```go
func TestConfigManager_LoadConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {
            name:    "有效配置",
            input:   "testdata/valid.json",
            want:    &Config{...},
            wantErr: false,
        },
        {
            name:    "无效配置",
            input:   "testdata/invalid.json",
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadConfig(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行测试并显示详细输出
go test -v ./...

# 运行特定包的测试
go test ./internal/config/
```

### 测试覆盖率目标

- **核心包**（config、handler）：80%+ 覆盖率
- **工具包**（common）：70%+ 覆盖率
- **CLI 命令**：基本集成测试

## 提交规范

我们遵循 [约定式提交](https://www.conventionalcommits.org/zh-hans/) 规范。

### 提交消息格式

```
<类型>(<范围>): <主题>

<正文>

<脚注>
```

### 类型

- `feat`：新功能
- `fix`：错误修复
- `docs`：文档变更
- `style`：代码风格变更（格式化、缺少分号等）
- `refactor`：代码重构，无功能变更
- `perf`：性能改进
- `test`：添加或更新测试
- `chore`：维护任务（依赖、构建等）
- `ci`：CI/CD 变更

### 示例

```bash
# 功能添加
git commit -m "feat(web): 添加配置比较功能"

# 错误修复
git commit -m "fix(config): 解决配置切换中的竞态条件"

# 文档
git commit -m "docs: 更新 Windows 安装说明"

# 重构
git commit -m "refactor(handler): 提取通用验证逻辑"

# 多个变更（使用正文）
git commit -m "feat(export): 添加压缩支持

- 实现导出文件的 gzip 压缩
- 为 export 命令添加 --compress 标志
- 更新新标志的文档"
```

### 提交最佳实践

- 保持提交原子化（每次提交一个逻辑变更）
- 编写清晰、描述性的提交消息
- 在适当时引用问题编号（如 "fixes #123"）
- 避免提交过多不相关的变更

## Pull Request 流程

### 提交前

1. **更新分支**到最新的上游变更：
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **运行所有检查**：
   ```bash
   npm run fmt
   npm run vet
   npm test
   ```

3. **手动测试**（如适用）

4. **更新文档**（如果您更改了功能）

### 提交 Pull Request

1. **推送分支**到您的 fork：
   ```bash
   git push origin feature/your-feature-name
   ```

2. **在 GitHub 上打开 PR**，包含：
   - 遵循提交约定的清晰标题
   - 详细的变更描述
   - 相关问题的引用
   - UI 变更的截图/演示

3. **PR 模板**（示例）：
   ```markdown
   ## 描述
   简要描述此 PR 的作用

   ## 变更类型
   - [ ] 错误修复
   - [ ] 新功能
   - [ ] 破坏性变更
   - [ ] 文档更新

   ## 相关问题
   修复 #123

   ## 测试
   - [ ] 添加/更新了测试
   - [ ] 完成了手动测试

   ## 检查清单
   - [ ] 代码遵循项目风格指南
   - [ ] 完成了自我审查
   - [ ] 更新了文档
   - [ ] 没有产生新的警告
   ```

### 审查流程

- 维护者将审查您的 PR
- 通过推送新提交来处理反馈
- 一旦批准，维护者将合并您的 PR
- 可能会使用 squash 提交以保持历史清洁

### PR 指南

- 保持 PR 专注且规模合理
- 每个 PR 一个功能/修复
- 及时回应审查意见
- 对反馈和建议保持开放态度

## 问题反馈

### 错误报告

报告错误时，请包含：

- **描述**：清晰的错误描述
- **复现步骤**：详细的复现步骤
- **预期行为**：应该发生什么
- **实际行为**：实际发生了什么
- **环境信息**：
  - 操作系统（macOS、Linux、Windows）
  - Go 版本（`go version`）
  - Node.js 版本（`node --version`）
  - cc-switch 版本（`cc-switch --version`）
- **日志/截图**：任何相关的错误消息或截图

### 功能请求

请求新功能时，请包含：

- **用例**：描述您试图解决的问题
- **建议方案**：您解决问题的想法
- **替代方案**：您考虑过的其他解决方案
- **附加上下文**：任何其他相关信息

### 问题标签

我们使用标签对问题进行分类：

- `bug`：某些功能不工作
- `enhancement`：新功能或请求
- `documentation`：文档改进
- `good first issue`：适合新手
- `help wanted`：需要额外关注
- `question`：需要更多信息

## 发布流程

### 版本控制

我们遵循[语义化版本](https://semver.org/lang/zh-CN/)：

- **主版本号**：不兼容的 API 变更
- **次版本号**：向后兼容的功能新增
- **修订号**：向后兼容的问题修复

### 发布步骤（维护者）

1. 更新 `package.json` 中的版本
2. 更新 `CHANGELOG.md`
3. 创建 git 标签：`git tag v1.x.x`
4. 推送标签：`git push origin v1.x.x`
5. GitHub Actions 将自动构建和发布

### 发布前检查清单

- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] CHANGELOG 已更新
- [ ] 版本号正确递增
- [ ] 破坏性变更已记录

## 获取帮助

如果您需要帮助：

1. **查看文档**：README.md 和现有问题
2. **提问**：使用 `question` 标签开一个 GitHub issue
3. **参与讨论**：参与 GitHub Discussions（如果已启用）

## 贡献者认可

贡献者将在以下位置获得认可：

- GitHub 贡献者页面
- 重大贡献的发布说明
- CHANGELOG.md 中的特别提及

## 许可证

通过为 cc-switch 做出贡献，您同意您的贡献将根据 MIT 许可证进行许可。

---

感谢您为 cc-switch 做出贡献！您的努力使这个工具对每个人都更好。
