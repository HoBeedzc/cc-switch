# Claude Code Configuration Switcher | Claude Code 配置切换工具

A command-line tool for managing and switching between different Claude Code configurations.

用于管理和切换不同 Claude Code 配置的命令行工具。

[English](#english) | [中文](#中文)

---

## English

### Installation

Install globally via npm:

```bash
npm install -g cc-switch
```

### Usage

#### List Configurations
```bash
cc-switch list
```
Shows all available configurations with the current one highlighted.

#### Create New Configuration
```bash
cc-switch new <name>
```
Creates a new configuration by copying the current Claude Code settings.

#### Switch Configuration
```bash
cc-switch use <name>
```
Switches to the specified configuration.

#### Switch to Previous Configuration
```bash
cc-switch use --previous
# or
cc-switch use -p
```
Switches to the previously used configuration.

#### Empty Mode (Temporary Configuration Removal)
```bash
cc-switch use --empty
# or
cc-switch use -e
```
Temporarily removes all Claude Code configurations (empty mode). This is useful when you want to disable Claude Code temporarily without losing your saved configurations.

#### Restore from Empty Mode
```bash
cc-switch use --restore
```
Restores from empty mode to the previous configuration that was active before entering empty mode.

#### Interactive Mode
```bash
cc-switch use
# or
cc-switch use -i
```
Enters interactive mode where you can select configurations using arrow keys. In interactive mode, you can also select special options like "Empty Mode" or "Restore Previous".

#### Delete Configuration
```bash
cc-switch delete <name>
```
Deletes the specified configuration (cannot delete currently active one).

#### Show Current Configuration
```bash
cc-switch current
```
Displays the name of the currently active configuration.

### Examples

```bash
# List all configurations
cc-switch list

# Create a work configuration
cc-switch new work

# Switch to work configuration
cc-switch use work

# Switch to previous configuration
cc-switch use --previous

# Enter empty mode (disable all configurations temporarily)
cc-switch use --empty

# Check status in empty mode
cc-switch current

# Restore from empty mode
cc-switch use --restore

# Use interactive mode for selection
cc-switch use

# Show current configuration
cc-switch current

# View configuration details
cc-switch view work

# Edit configuration
cc-switch edit work

# Delete unused configuration
cc-switch delete old-config
```

### How It Works

The tool manages configurations in `~/.claude/profiles/` directory:

```
~/.claude/
├── settings.json          # Current active configuration (removed in empty mode)
├── profiles/              # Stored configurations
│   ├── default.json       # Default configuration
│   ├── work.json          # Work configuration
│   ├── personal.json      # Personal configuration
│   └── .empty_backup_settings.json  # Backup when in empty mode
├── .current              # Current configuration marker
└── .empty_mode           # Empty mode state file (present in empty mode)
```

#### Initialization

On first run:
- If `~/.claude/settings.json` exists, it creates a `default` profile
- If no configuration exists, it guides you to create one first

#### Security

- Configuration files are stored with 600 permissions (owner read/write only)
- Atomic operations using temporary files ensure configuration integrity
- Automatic backup of current configuration before switching
- Empty mode creates secure backup before removing settings.json
- Rollback mechanisms prevent data loss during operations

#### View Configuration Details
```bash
cc-switch view <name>
```
Displays the settings for a specific configuration without switching to it.

#### Edit Configuration
```bash
cc-switch edit <name>
```
Opens the configuration in your default text editor for modification.

### Commands Reference

| Command | Description |
|---------|-------------|
| `list` | List all available configurations |
| `new <name>` | Create a new configuration |
| `use <name>` | Switch to a configuration |
| `use -p, --previous` | Switch to previous configuration |
| `use -e, --empty` | Enter empty mode (disable configurations) |
| `use --restore` | Restore from empty mode to previous configuration |
| `use -i, --interactive` | Enter interactive selection mode |
| `delete <name>` | Delete a configuration |
| `current` | Show current configuration or empty mode status |
| `view <name>` | View configuration details |
| `edit <name>` | Edit configuration in text editor |

### Empty Mode Feature

Empty mode is a special state where all Claude Code configurations are temporarily disabled. This is useful in scenarios where:

- You want to temporarily disable Claude Code without losing your saved configurations
- You need to test Claude Code's default behavior without any custom settings
- You want to troubleshoot configuration issues by temporarily removing all settings

#### How Empty Mode Works

1. **Entering Empty Mode**: Use `cc-switch use --empty` to enter empty mode
   - Current `settings.json` is safely backed up to `.empty_backup_settings.json`
   - The `settings.json` file is removed, disabling Claude Code
   - A `.empty_mode` file is created to track the state and previous configuration

2. **Empty Mode Status**: All commands work normally and show empty mode indicators
   - `cc-switch current` shows "Empty mode (no configuration active)"
   - `cc-switch list` shows "Empty mode active" warning with helpful tips

3. **Exiting Empty Mode**: Multiple ways to restore configurations
   - `cc-switch use --restore` restores to the previous configuration
   - `cc-switch use <name>` switches to any specific configuration
   - Both methods automatically restore the backup and clean up empty mode state

#### Safety Features

- **Atomic Operations**: All operations use temporary files and atomic renames
- **Automatic Rollback**: Failed operations automatically restore the original state
- **Backup Validation**: Settings backup is validated before empty mode activation
- **State Tracking**: Complete state information is preserved for reliable restoration

### Requirements

- Node.js 14.0.0 or higher
- Claude Code installed and configured

### Development

#### Building from Source

```bash
# Clone the repository
git clone https://github.com/hobee/cc-switch.git
cd cc-switch

# Install dependencies
go mod tidy

# Build for all platforms
npm run build

# Test locally
go run . --help
```

#### Project Structure

```
cc-switch/
├── main.go                 # Entry point
├── cmd/                    # CLI commands
│   ├── root.go
│   ├── list.go
│   ├── new.go
│   ├── use.go
│   ├── delete.go
│   ├── current.go
│   ├── view.go
│   └── edit.go
├── internal/config/        # Configuration management
│   └── manager.go
├── internal/handler/       # Business logic handlers
│   ├── config_handler.go
│   └── types.go
├── internal/ui/           # User interface layer
│   ├── cli.go
│   ├── interactive.go
│   └── interfaces.go
├── scripts/               # Build and install scripts
│   ├── build.sh
│   └── install.js
└── package.json           # npm configuration
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/hobee/cc-switch/issues) page
2. Create a new issue if your problem isn't already reported
3. Include your OS, Node.js version, and error messages

---

## 中文

### 安装

通过 npm 全局安装：

```bash
npm install -g cc-switch
```

### 使用方法

#### 列出配置
```bash
cc-switch list
```
显示所有可用配置，当前配置会高亮显示。

#### 创建新配置
```bash
cc-switch new <名称>
```
通过复制当前 Claude Code 设置来创建新配置。

#### 切换配置
```bash
cc-switch use <名称>
```
切换到指定的配置。

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
临时移除所有 Claude Code 配置（空配置模式）。这在您想要暂时禁用 Claude Code 而不丢失已保存配置时很有用。

#### 从空配置模式恢复
```bash
cc-switch use --restore
```
从空配置模式恢复到进入空配置模式之前活动的配置。

#### 交互模式
```bash
cc-switch use
# 或
cc-switch use -i
```
进入交互模式，您可以使用箭头键选择配置。在交互模式中，您还可以选择特殊选项，如"空配置模式"或"恢复上一个"。

#### 删除配置
```bash
cc-switch delete <名称>
```
删除指定配置（无法删除当前正在使用的配置）。

#### 显示当前配置
```bash
cc-switch current
```
显示当前正在使用的配置名称。

### 使用示例

```bash
# 列出所有配置
cc-switch list

# 创建工作配置
cc-switch new work

# 切换到工作配置
cc-switch use work

# 切换到上一个配置
cc-switch use --previous

# 进入空配置模式（临时禁用所有配置）
cc-switch use --empty

# 在空配置模式下检查状态
cc-switch current

# 从空配置模式恢复
cc-switch use --restore

# 使用交互模式进行选择
cc-switch use

# 显示当前配置
cc-switch current

# 查看配置详情
cc-switch view work

# 编辑配置
cc-switch edit work

# 删除不用的配置
cc-switch delete old-config
```

### 工作原理

工具在 `~/.claude/profiles/` 目录中管理配置：

```
~/.claude/
├── settings.json          # 当前活动配置（空配置模式下被移除）
├── profiles/              # 存储的配置
│   ├── default.json       # 默认配置
│   ├── work.json          # 工作配置
│   ├── personal.json      # 个人配置
│   └── .empty_backup_settings.json  # 空配置模式下的备份
├── .current              # 当前配置标记文件
└── .empty_mode           # 空配置模式状态文件（空配置模式下存在）
```

#### 初始化

首次运行时：
- 如果 `~/.claude/settings.json` 存在，会创建一个 `default` 配置
- 如果没有配置存在，会引导用户先创建配置

#### 安全性

- 配置文件使用 600 权限存储（仅所有者可读写）
- 使用临时文件的原子操作确保配置完整性
- 切换前自动备份当前配置
- 空配置模式在移除 settings.json 前创建安全备份
- 回滚机制防止操作过程中的数据丢失

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
| `list` | 列出所有可用配置 |
| `new <名称>` | 创建新配置 |
| `use <名称>` | 切换到配置 |
| `use -p, --previous` | 切换到上一个配置 |
| `use -e, --empty` | 进入空配置模式（禁用配置） |
| `use --restore` | 从空配置模式恢复到上一个配置 |
| `use -i, --interactive` | 进入交互选择模式 |
| `delete <名称>` | 删除配置 |
| `current` | 显示当前配置或空配置模式状态 |
| `view <名称>` | 查看配置详情 |
| `edit <名称>` | 在文本编辑器中编辑配置 |

### 空配置模式功能

空配置模式是一种特殊状态，可以临时禁用所有 Claude Code 配置。这在以下场景中很有用：

- 您想要暂时禁用 Claude Code 而不丢失已保存的配置
- 您需要测试 Claude Code 的默认行为而不使用任何自定义设置
- 您想要通过临时移除所有设置来排查配置问题

#### 空配置模式的工作原理

1. **进入空配置模式**：使用 `cc-switch use --empty` 进入空配置模式
   - 当前的 `settings.json` 被安全备份到 `.empty_backup_settings.json`
   - `settings.json` 文件被移除，禁用 Claude Code
   - 创建 `.empty_mode` 文件来跟踪状态和之前的配置

2. **空配置模式状态**：所有命令正常工作并显示空配置模式指示器
   - `cc-switch current` 显示 "空配置模式（无配置激活）"
   - `cc-switch list` 显示 "空配置模式激活" 警告和有用提示

3. **退出空配置模式**：多种方式恢复配置
   - `cc-switch use --restore` 恢复到之前的配置
   - `cc-switch use <名称>` 切换到任何指定配置
   - 两种方法都会自动恢复备份并清理空配置模式状态

#### 安全特性

- **原子操作**：所有操作使用临时文件和原子重命名
- **自动回滚**：失败的操作自动恢复原始状态
- **备份验证**：在激活空配置模式前验证设置备份
- **状态跟踪**：保存完整状态信息以确保可靠恢复

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
│   └── edit.go
├── internal/config/        # 配置管理
│   └── manager.go
├── internal/handler/       # 业务逻辑处理器
│   ├── config_handler.go
│   └── types.go
├── internal/ui/           # 用户界面层
│   ├── cli.go
│   ├── interactive.go
│   └── interfaces.go
├── scripts/               # 构建和安装脚本
│   ├── build.sh
│   └── install.js
└── package.json           # npm 配置
```

### 贡献

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

### 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

### 支持

如果遇到问题或有疑问：

1. 查看 [Issues](https://github.com/hobee/cc-switch/issues) 页面
2. 如果问题尚未报告，请创建新 issue
3. 请包含您的操作系统、Node.js 版本和错误信息