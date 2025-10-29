# Claude Code 配置切换工具

[English](README.md) | [中文](README.zh.md)

用于管理和切换不同 Claude Code 配置的命令行工具。

---

### 安装

通过 npm 全局安装：

```bash
npm install -g @hobeeliu/cc-switch
```

### 使用方法

#### 列出配置
```bash
cc-switch list
```
显示所有可用配置，当前配置高亮显示。

#### 初始化配置（首次设置）
```bash
cc-switch init
```
通过交互式设置初始化 Claude Code 配置。提示输入 API token 和基础 URL。

#### 创建新配置
```bash
# 从默认模板创建
cc-switch new <名称>

# 从指定模板创建
cc-switch new <名称> -t <模板>
cc-switch new <名称> --template <模板>

# 交互式模板创建（交互式填写模板字段）
cc-switch new <名称> -i
cc-switch new <名称> --interactive

# 创建并在创建后立即切换
cc-switch new <名称> -u
cc-switch new <名称> --use
```
使用模板结构创建新配置。默认模板提供基本结构，交互模式允许通过引导提示填写模板字段。使用 `--use` 标志可在创建后自动切换到新配置。

#### 切换配置
```bash
# 切换到指定配置
cc-switch use <名称>

# 切换后启动 Claude Code CLI
cc-switch use <名称> -l
cc-switch use <名称> --launch
```
切换到指定的配置。使用 `--launch` 标志在切换后自动启动 Claude Code CLI。

#### 切换到上一个配置
```bash
cc-switch use --previous
# 或
cc-switch use -p
```
切换到之前使用的配置。

#### 空配置模式（临时移除配置）
```bash
cc-switch use --empty
# 或
cc-switch use -e
```
临时移除所有 Claude Code 配置（空配置模式）。适用于在不丢失已保存配置的情况下，临时禁用 Claude Code。

#### 从空配置模式恢复
```bash
cc-switch use --restore
```
从空配置模式恢复到进入空配置模式之前活动的配置。

#### 刷新当前配置
```bash
cc-switch use --refresh
# 或
cc-switch use -f
```
通过重新应用来刷新当前配置。这在手动编辑配置文件后很有用，可确保更改正确应用。

#### 交互模式
```bash
cc-switch use
# 或
cc-switch use -i
```
进入交互模式，使用方向键选择配置。在交互模式中，可选择“空配置模式”或“恢复上一个”等特殊选项。

#### 删除配置
```bash
# 删除指定配置
cc-switch rm <名称>

# 交互式删除模式
cc-switch rm
cc-switch rm -i

# 删除当前配置并进入空配置模式
cc-switch rm -c
cc-switch rm --current

# 删除所有配置（需要手动确认）
cc-switch rm -a
cc-switch rm --all

# 跳过确认提示
cc-switch rm <名称> -y
cc-switch rm <名称> --yes
```
提供多种删除选项。除非使用 `--current` 标志，否则无法删除当前激活的配置。为安全起见，使用 `--all` 时需要输入 "DELETE ALL"。

#### 复制配置
```bash
cc-switch cp <源名称> <目标名称>
```
复制现有配置并使用新名称创建副本。原配置保持不变。

#### 移动（重命名）配置
```bash
cc-switch mv <旧名称> <新名称>
```
重命名现有配置。如果该配置当前处于激活状态，会自动更新标记。

#### 导出配置
```bash
# 导出指定配置
cc-switch export <配置名称> -o backup.ccx

# 导出所有配置
cc-switch export --all -o all-configs.ccx

# 导出当前配置
cc-switch export --current -o current.ccx
```
将配置导出为加密备份文件（.ccx 格式）。支持可选密码保护。

#### 导入配置
```bash
# 从备份文件导入
cc-switch import backup.ccx

# 导入并覆盖已存在的同名配置
cc-switch import backup.ccx --conflict=overwrite

# 导入并跳过冲突配置
cc-switch import backup.ccx --conflict=skip

# 导入并自动重命名冲突配置（默认）
cc-switch import backup.ccx --conflict=both

# 仅预览导入结果，不做更改
cc-switch import backup.ccx --dry-run

# 通过参数提供解密密码（也可交互输入）
cc-switch import backup.ccx -p <密码>
```
从加密备份文件导入配置。支持冲突处理模式、试运行（dry-run）以及加密归档。

#### 测试配置连接性
```bash
# 测试指定配置
cc-switch test <配置名称>

# 测试当前配置
cc-switch test --current

# 测试所有配置
cc-switch test --all

# 快速连接测试
cc-switch test --quick

# 失败后重试
cc-switch test -r 5                 # 最多重试 5 次
cc-switch test -r -1                # 无限重试直到成功
cc-switch test -r 3 --retry-interval 5s  # 重试 3 次，间隔 5 秒
```
测试 Claude Code API 连接性和认证情况。

#### Web 界面
```bash
# 使用默认设置启动 Web 界面
cc-switch web

# 自定义端口启动
cc-switch web --port 8080

# 自定义主机和端口启动
cc-switch web --host 0.0.0.0 --port 8080

# 启动后自动打开浏览器
cc-switch web --open

# 静默启动
cc-switch web --quiet
```
在 http://localhost:13501（或自定义主机:端口）启动现代化的基于浏览器的配置管理界面。

**Web 界面功能特性：**
- **配置管理**：创建、编辑、删除、切换配置文件
- **模板管理**：模板的完整 CRUD 操作，带安全校验
- **在线配置编辑**：在浏览器中直接编辑配置，支持 JSON 校验
- **API 连接测试**：可对所有或指定配置进行 Claude Code API 连接测试
- **实时状态**：查看当前激活配置及系统状态
- **响应式设计**：现代、移动友好的界面与导航
- **安全功能**：路径遍历防护、输入校验和安全操作

#### 模板管理
```bash
# 列出可用模板
cc-switch list -t
cc-switch list --template

# 删除模板（交互模式）
cc-switch rm -t
cc-switch rm --template

# 删除指定模板
cc-switch rm -t <模板名称>

# 复制模板
cc-switch cp -t <源模板> <目标模板>

# 移动（重命名）模板
cc-switch mv -t <旧模板> <新模板>

# 查看模板内容
cc-switch view -t <模板名称>

# 编辑模板
cc-switch edit -t <模板名称>
```
模板提供创建新配置的预配置结构。出于系统安全考虑，默认模板不可删除。

#### 显示当前配置
```bash
cc-switch current
```
显示当前激活的配置名称。

### 使用示例

```bash
# 初始化配置（首次设置）
cc-switch init

# 列出所有配置
cc-switch list

# 列出所有模板
cc-switch list -t

# 从默认模板创建工作配置
cc-switch new work

# 从指定模板创建配置
cc-switch new personal -t company-template

# 交互式填写模板创建配置
cc-switch new dev-env -i

# 创建后立即切换
cc-switch new staging -u

# 切换到工作配置
cc-switch use work

# 切换并启动 Claude Code CLI
cc-switch use work --launch

# 切换到上一个配置
cc-switch use --previous

# 复制配置作为备份
cc-switch cp work work-backup

# 重命名配置
cc-switch mv work-backup work-v2

# 测试当前配置
cc-switch test --current

# 导出所有配置
cc-switch export --all -o backup.ccx

# 导入配置
cc-switch import backup.ccx

# 启动 Web 界面
cc-switch web

# 启动后自动打开浏览器
cc-switch web --open

# 在自定义端口静默启动
cc-switch web --port 8080 --quiet

# 进入空配置模式（临时禁用所有配置）
cc-switch use --empty

# 在空配置模式下检查状态
cc-switch current

# 从空配置模式恢复
cc-switch use --restore

# 交互模式进行选择
cc-switch use

# 手动编辑后刷新当前配置
cc-switch use --refresh

# 显示当前配置
cc-switch current

# 查看配置详情
cc-switch view work

# 编辑配置
cc-switch edit work

# 删除无用配置
cc-switch rm old-config

# 删除当前配置并进入空配置模式
cc-switch rm --current

# 交互式删除
cc-switch rm -i

# 模板管理
cc-switch view -t default
cc-switch cp -t default my-template
cc-switch edit -t my-template
```

### 工作原理

工具在 `~/.claude/profiles/` 目录中管理配置：

```
~/.claude/
├── settings.json          # 当前激活配置（空配置模式下将被移除）
├── profiles/              # 存储的配置
│   ├── default.json       # 默认配置
│   ├── work.json          # 工作配置
│   ├── personal.json      # 个人配置
│   ├── templates/         # 配置模板
│   │   ├── default.json   # 默认模板（不可删除）
│   │   └── company.json   # 自定义公司模板
│   └── .empty_backup_settings.json  # 空配置模式下的备份
├── .current              # 当前配置标记
└── .empty_mode           # 空配置模式状态文件（空配置模式下存在）
```

#### 初始化

首次运行时：
- 若存在 `~/.claude/settings.json`，将创建 `default` 配置
- 若不存在任何配置，将引导您先创建配置

#### 安全性

- 配置文件以 600 权限存储（仅所有者可读写）
- 通过临时文件进行原子操作，确保配置完整性
- 切换前自动备份当前配置
- 空配置模式在移除 settings.json 前创建安全备份
- 回滚机制防止在操作中发生数据丢失

#### 查看配置详情
```bash
cc-switch view <名称>
```
显示指定配置的设置内容，不会切换到该配置。

#### 编辑配置
```bash
cc-switch edit <名称>
```
在默认文本编辑器中打开配置进行修改。

### 命令参考

| 命令 | 说明 |
|------|------|
| `init` | 通过交互式设置初始化 Claude Code 配置 |
| `list` | 列出所有可用配置 |
| `list -t, --template` | 列出所有可用模板 |
| `new <名称>` | 从默认模板创建新配置 |
| `new <名称> -t <模板>` | 从指定模板创建新配置 |
| `new <名称> -i, --interactive` | 交互式填写模板创建配置 |
| `new <名称> -u, --use` | 创建后立即切换到该配置 |
| `use <名称>` | 切换到配置 |
| `use <名称> -l, --launch` | 切换到配置并启动 Claude Code CLI |
| `use -p, --previous` | 切换到上一个配置 |
| `use -e, --empty` | 进入空配置模式（禁用配置） |
| `use --restore` | 从空配置模式恢复到之前的配置 |
| `use -f, --refresh` | 刷新当前配置（重新应用） |
| `use -i, --interactive` | 进入交互选择模式 |
| `cp <源> <目标>` | 复制配置 |
| `cp -t <源> <目标>` | 复制模板 |
| `mv <旧> <新>` | 移动（重命名）配置 |
| `mv -t <旧> <新>` | 移动（重命名）模板 |
| `rm <名称>` | 删除配置 |
| `rm -c, --current` | 删除当前配置并进入空配置模式 |
| `rm -a, --all` | 删除所有配置（需要手动确认） |
| `rm -t <模板>` | 删除模板 |
| `export [配置]` | 导出配置到备份文件 |
| `import <文件>` | 从备份文件导入配置 |
| `test [配置]` | 测试配置 API 连接 |
| `web` | 启动带配置管理的 Web 界面 |
| `current` | 显示当前配置或空配置模式状态 |
| `view <名称>` | 查看配置详情 |
| `view -t <模板>` | 查看模板详情 |
| `edit <名称>` | 在文本编辑器中编辑配置 |
| `edit -t <模板>` | 在文本编辑器中编辑模板 |

### 模板系统

模板系统用于创建标准化的配置结构，以便在不同环境下保持一致的设置。

#### 什么是模板？

模板是预先配置好的 JSON 结构，作为创建新配置的蓝本。它可以包含：

- **通用设置**：适用于各环境的 Claude Code 常用设置
- **占位符字段**：在配置创建过程中可交互填写的字段
- **环境特定值**：可根据使用场景自定义的默认值

#### 模板特性

- **默认模板**：内置模板，不能删除，提供基础结构
- **自定义模板**：按需创建（公司、项目等）
- **交互式创建**：在创建时通过引导提示填写模板字段
- **模板管理**：像配置一样对模板进行复制、编辑、查看和组织

#### 使用模板

1. **列出模板**：`cc-switch list -t` 查看可用模板
2. **从模板创建**：`cc-switch new myconfig -t mytemplate` 使用指定模板
3. **交互模式**：`cc-switch new myconfig -i` 引导式填写模板字段
4. **模板管理**：使用带 `-t` 标志的标准操作（cp、mv、edit、view）

#### 模板结构

模板与配置使用相同的 JSON 结构，但可包含在交互式创建时填写的特殊占位符。

### 空配置模式功能

空配置模式是一种特殊状态，可临时禁用所有 Claude Code 配置。

- 临时禁用 Claude Code 且不丢失已保存配置
- 测试 Claude Code 的默认行为（无自定义设置）
- 通过移除所有设置来排查配置问题

#### 空配置模式的工作原理

1. **进入空配置模式**：`cc-switch use --empty`
   - 将当前 `settings.json` 备份到 `.empty_backup_settings.json`
   - 移除 `settings.json` 以禁用 Claude Code
   - 创建 `.empty_mode` 文件以保存状态和之前的配置

2. **空配置模式状态**：命令正常工作并显示空配置模式提示
   - `cc-switch current` 显示 “空配置模式（无配置激活）”
   - `cc-switch list` 显示 “空配置模式已启用” 警告与提示

3. **退出空配置模式**：多种方式恢复
   - `cc-switch use --restore` 恢复至之前的配置
   - `cc-switch use <名称>` 切换到任意配置
   - 两种方式均会自动恢复备份并清理空配置模式状态

#### 安全特性

- **原子操作**：使用临时文件与原子重命名
- **自动回滚**：失败时自动恢复原始状态
- **备份校验**：启用空配置模式前校验备份
- **状态跟踪**：保留完整状态信息确保可靠恢复

### 系统要求

- Node.js 14.0.0 或更高版本
- 已安装并配置 Claude Code

### 开发

#### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/hobee/cc-switch.git
cd cc-switch

# 安装依赖
go mod tidy

# 构建所有平台
npm run build

# 本地测试
go run . --help
```

#### 项目结构

```
cc-switch/
├── main.go                 # 程序入口
├── cmd/                    # CLI 命令
│   ├── root.go
│   ├── list.go
│   ├── new.go
│   ├── use.go
│   ├── delete.go
│   ├── current.go
│   ├── view.go
│   ├── edit.go
│   ├── cp.go               # 复制命令
│   ├── mv.go               # 移动命令
│   ├── export.go           # 导出命令
│   ├── import.go           # 导入命令
│   ├── test.go             # 测试命令
│   ├── init.go             # 初始化命令
│   └── web.go              # Web 界面命令
├── internal/config/        # 配置管理
│   └── manager.go
├── internal/handler/       # 业务逻辑处理器
│   ├── config_handler.go
│   ├── api_tester.go       # API 测试功能
│   └── types.go
├── internal/ui/           # 用户界面层
│   ├── cli.go
│   ├── interactive.go
│   └── interfaces.go
├── internal/web/          # Web 界面
│   ├── server.go
│   ├── handlers.go
│   └── assets/            # Web 静态资源
│       ├── css/
│       │   └── main.css   # Web 界面样式
│       └── js/
│           └── main.js    # Web 界面逻辑
├── internal/export/       # 导出功能
│   ├── exporter.go
│   └── format.go
├── internal/import/       # 导入功能
│   └── importer.go
├── internal/common/       # 通用工具
│   ├── compress.go
│   └── crypto.go
├── scripts/               # 构建与安装脚本
│   ├── build.sh
│   └── install.js
└── package.json           # npm 配置
```

### 贡献

欢迎贡献！请查看我们的[贡献指南](CONTRIBUTING.zh.md)了解详细信息：

- 开发环境配置和工作流程
- 代码风格指南
- 测试要求
- Pull Request 流程
- 问题反馈

快速开始：

1. Fork 仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送分支（`git push origin feature/amazing-feature`）
5. 发起 Pull Request

英文版本请参阅 [Contributing Guide (English)](CONTRIBUTING.md)。

### 许可证

本项目采用 MIT 许可证 - 详情见 [LICENSE](LICENSE)。

### 支持

如遇到问题或有疑问：

1. 查看 [Issues](https://github.com/hobee/cc-switch/issues) 页面
2. 若未发现相同问题，请新建 issue
3. 请附上操作系统、Node.js 版本及错误信息
