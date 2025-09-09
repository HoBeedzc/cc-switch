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

#### Initialize Configuration (First Time Setup)
```bash
cc-switch init
```
Initialize Claude Code configuration with interactive setup. Prompts for API token and base URL.

#### Create New Configuration
```bash
# Create from default template
cc-switch new <name>

# Create from specific template
cc-switch new <name> -t <template>
cc-switch new <name> --template <template>

# Interactive template creation (fills template fields interactively)
cc-switch new <name> -i
cc-switch new <name> --interactive
```
Creates a new configuration using template structure. The default template provides a basic structure, and interactive mode allows you to fill in template fields with guided prompts.

#### Switch Configuration
```bash
# Switch to specific configuration
cc-switch use <name>

# Launch Claude Code CLI after switching
cc-switch use <name> -l
cc-switch use <name> --launch
```
Switches to the specified configuration. Use the `--launch` flag to automatically start Claude Code CLI after switching.

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
# Delete specific configuration
cc-switch rm <name>

# Interactive deletion mode
cc-switch rm
cc-switch rm -i

# Delete current configuration and enter empty mode
cc-switch rm -c
cc-switch rm --current

# Delete all configurations (requires manual confirmation)
cc-switch rm -a
cc-switch rm --all

# Skip confirmation prompts
cc-switch rm <name> -y
cc-switch rm <name> --yes
```
Deletes configurations with various options. Cannot delete currently active configuration unless using `--current` flag. The `--all` flag requires typing "DELETE ALL" for safety.

#### Copy Configuration
```bash
cc-switch cp <source> <destination>
```
Creates a copy of an existing configuration with a new name. Original configuration remains unchanged.

#### Move (Rename) Configuration
```bash
cc-switch mv <old-name> <new-name>
```
Renames an existing configuration. If currently active, the marker will be updated automatically.

#### Export Configurations
```bash
# Export specific profile
cc-switch export <profile-name> -o backup.ccx

# Export all profiles
cc-switch export --all -o all-configs.ccx

# Export current profile
cc-switch export --current -o current.ccx
```
Export configurations to encrypted backup files (.ccx format). Supports password protection.

#### Import Configurations
```bash
# Import from backup file
cc-switch import backup.ccx

# Import with overwrite existing profiles
cc-switch import backup.ccx --conflict=overwrite

# Import and skip conflicting profiles
cc-switch import backup.ccx --conflict=skip

# Import and rename conflicting profiles (default)
cc-switch import backup.ccx --conflict=both

# Preview import without making changes
cc-switch import backup.ccx --dry-run
```
Import configurations from encrypted backup files. Supports conflict resolution and dry-run mode.

#### Test Configuration Connectivity
```bash
# Test specific configuration
cc-switch test <profile-name>

# Test current configuration
cc-switch test --current

# Test all configurations
cc-switch test --all

# Quick connectivity test
cc-switch test --quick
```
Test Claude Code API connectivity and authentication for configurations.

#### Web Interface
```bash
# Launch web interface with default settings
cc-switch web

# Launch on custom port
cc-switch web --port 8080

# Launch on custom host and port
cc-switch web --host 0.0.0.0 --port 8080

# Open browser automatically after starting
cc-switch web --open

# Suppress startup messages
cc-switch web --quiet
```
Launch a modern browser-based interface for managing configurations at http://localhost:13501 (or custom host:port).

**Web Interface Features:**
- **Profile Management**: Create, edit, delete, and switch between configuration profiles
- **Template Management**: Full template CRUD operations with security validation
- **Live Configuration Editing**: Edit configurations directly in the browser with JSON validation
- **API Connectivity Testing**: Test Claude Code API connections for all or specific profiles
- **Real-time Status**: View current active configuration and system status
- **Responsive Design**: Modern, mobile-friendly interface with intuitive navigation
- **Security Features**: Path traversal protection, input validation, and secure operations

#### Template Management
```bash
# List available templates
cc-switch list -t
cc-switch list --template

# Delete a template (interactive mode)
cc-switch rm -t
cc-switch rm --template

# Delete specific template
cc-switch rm -t <template-name>

# Copy template
cc-switch cp -t <source-template> <dest-template>

# Move (rename) template
cc-switch mv -t <old-template> <new-template>

# View template content
cc-switch view -t <template-name>

# Edit template
cc-switch edit -t <template-name>
```
Templates provide pre-configured structures for creating new configurations. The default template cannot be deleted for system safety.

#### Show Current Configuration
```bash
cc-switch current
```
Displays the name of the currently active configuration.

### Examples

```bash
# Initialize configuration (first time setup)
cc-switch init

# List all configurations
cc-switch list

# List all templates
cc-switch list -t

# Create a work configuration from default template
cc-switch new work

# Create configuration from specific template
cc-switch new personal -t company-template

# Create configuration with interactive template filling
cc-switch new dev-env -i

# Switch to work configuration
cc-switch use work

# Switch and launch Claude Code CLI
cc-switch use work --launch

# Switch to previous configuration
cc-switch use --previous

# Copy configuration for backup
cc-switch cp work work-backup

# Rename configuration
cc-switch mv work-backup work-v2

# Test current configuration
cc-switch test --current

# Export all configurations
cc-switch export --all -o backup.ccx

# Import configurations
cc-switch import backup.ccx

# Launch web interface
cc-switch web

# Launch web interface with automatic browser opening
cc-switch web --open

# Launch web interface on custom port with quiet mode
cc-switch web --port 8080 --quiet

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
cc-switch rm old-config

# Delete current configuration and enter empty mode
cc-switch rm --current

# Interactive deletion
cc-switch rm -i

# Template management
cc-switch view -t default
cc-switch cp -t default my-template
cc-switch edit -t my-template
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
│   ├── templates/         # Configuration templates
│   │   ├── default.json   # Default template (cannot be deleted)
│   │   └── company.json   # Custom company template
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
| `init` | Initialize Claude Code configuration with interactive setup |
| `list` | List all available configurations |
| `list -t, --template` | List all available templates |
| `new <name>` | Create a new configuration from default template |
| `new <name> -t <template>` | Create a new configuration from specific template |
| `new <name> -i, --interactive` | Create configuration with interactive template filling |
| `use <name>` | Switch to a configuration |
| `use <name> -l, --launch` | Switch to a configuration and launch Claude Code CLI |
| `use -p, --previous` | Switch to previous configuration |
| `use -e, --empty` | Enter empty mode (disable configurations) |
| `use --restore` | Restore from empty mode to previous configuration |
| `use -i, --interactive` | Enter interactive selection mode |
| `cp <source> <dest>` | Copy a configuration |
| `cp -t <source> <dest>` | Copy a template |
| `mv <old> <new>` | Move (rename) a configuration |
| `mv -t <old> <new>` | Move (rename) a template |
| `rm <name>` | Delete a configuration |
| `rm -c, --current` | Delete current configuration and enter empty mode |
| `rm -a, --all` | Delete ALL configurations (requires manual confirmation) |
| `rm -t <template>` | Delete a template |
| `export [profile]` | Export configurations to backup file |
| `import <file>` | Import configurations from backup file |
| `test [profile]` | Test configuration API connectivity |
| `web` | Launch web interface with configuration management |
| `current` | Show current configuration or empty mode status |
| `view <name>` | View configuration details |
| `view -t <template>` | View template details |
| `edit <name>` | Edit configuration in text editor |
| `edit -t <template>` | Edit template in text editor |

### Template System

The template system allows you to create standardized configuration structures for consistent setups across different environments.

#### What are Templates?

Templates are pre-configured JSON structures that serve as blueprints for creating new configurations. They can include:

- **Standard Settings**: Common Claude Code settings that apply across environments
- **Placeholder Fields**: Fields that can be filled interactively during configuration creation
- **Environment-Specific Values**: Default values that can be customized for different use cases

#### Template Features

- **Default Template**: A built-in template that cannot be deleted, providing basic structure
- **Custom Templates**: Create your own templates tailored to specific needs (company, project, etc.)
- **Interactive Creation**: Fill template fields with guided prompts during configuration creation
- **Template Management**: Copy, edit, view, and organize templates like configurations

#### Using Templates

1. **List Templates**: `cc-switch list -t` to see available templates
2. **Create from Template**: `cc-switch new myconfig -t mytemplate` to use a specific template
3. **Interactive Mode**: `cc-switch new myconfig -i` for guided template field input
4. **Template Management**: Use standard operations (cp, mv, edit, view) with `-t` flag

#### Template Structure

Templates use the same JSON structure as configurations but may include special placeholder values that get filled during interactive creation.

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
│   ├── edit.go
│   ├── cp.go               # Copy command
│   ├── mv.go               # Move command
│   ├── export.go           # Export command
│   ├── import.go           # Import command
│   ├── test.go             # Test command
│   ├── init.go             # Init command
│   └── web.go              # Web interface command
├── internal/config/        # Configuration management
│   └── manager.go
├── internal/handler/       # Business logic handlers
│   ├── config_handler.go
│   ├── api_tester.go       # API testing functionality
│   └── types.go
├── internal/ui/           # User interface layer
│   ├── cli.go
│   ├── interactive.go
│   └── interfaces.go
├── internal/web/          # Web interface
│   ├── server.go
│   ├── handlers.go
│   └── assets/            # Web static resources
│       ├── css/
│       │   └── main.css   # Web interface styles
│       └── js/
│           └── main.js    # Web interface logic
├── internal/export/       # Export functionality
│   ├── exporter.go
│   └── format.go
├── internal/import/       # Import functionality
│   └── importer.go
├── internal/common/       # Common utilities
│   ├── compress.go
│   └── crypto.go
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

#### 初始化配置（首次设置）
```bash
cc-switch init
```
通过交互式设置初始化 Claude Code 配置。提示输入 API token 和基础 URL。

#### 列出配置
```bash
cc-switch list
```
显示所有可用配置，当前配置会高亮显示。

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
```
使用模板结构创建新配置。默认模板提供基本结构，交互模式允许通过引导提示填写模板字段。

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
使用各种选项删除配置。无法删除当前活动配置，除非使用 `--current` 标志。`--all` 标志为安全起见需要输入 "DELETE ALL"。

#### 复制配置
```bash
cc-switch cp <源名称> <目标名称>
```
创建现有配置的副本并使用新名称。原配置保持不变。

#### 移动（重命名）配置
```bash
cc-switch mv <旧名称> <新名称>
```
重命名现有配置。如果是当前激活的配置，标记会自动更新。

#### 导出配置
```bash
# 导出指定配置
cc-switch export <配置名称> -o backup.ccx

# 导出所有配置
cc-switch export --all -o all-configs.ccx

# 导出当前配置
cc-switch export --current -o current.ccx
```
将配置导出为加密备份文件（.ccx 格式）。支持密码保护。

#### 导入配置
```bash
# 从备份文件导入
cc-switch import backup.ccx

# 导入并覆盖现有配置
cc-switch import backup.ccx --overwrite

# 导入并添加前缀
cc-switch import backup.ccx --prefix team-
```
从加密备份文件导入配置。支持冲突解决和试运行模式。

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
```
测试 Claude Code API 连接性和配置认证。

#### Web 界面
```bash
# 使用默认设置启动 Web 界面
cc-switch web

# 在自定义端口启动
cc-switch web --port 8080

# 在自定义主机和端口启动
cc-switch web --host 0.0.0.0 --port 8080

# 启动后自动打开浏览器
cc-switch web --open

# 禁用启动消息
cc-switch web --quiet
```
在 http://localhost:13501（或自定义主机:端口）启动现代化的基于浏览器的配置管理界面。

**Web 界面功能特性：**
- **配置管理**：创建、编辑、删除和切换配置文件
- **模板管理**：完整的模板 CRUD 操作，具备安全验证
- **在线配置编辑**：在浏览器中直接编辑配置，支持 JSON 验证
- **API 连接测试**：测试所有或指定配置的 Claude Code API 连接
- **实时状态显示**：查看当前活动配置和系统状态
- **响应式设计**：现代化、移动设备友好的直观界面
- **安全功能**：路径遍历保护、输入验证和安全操作

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
模板为创建新配置提供预配置结构。出于系统安全考虑，默认模板不能被删除。

#### 显示当前配置
```bash
cc-switch current
```
显示当前正在使用的配置名称。

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

# 使用交互式模板填写创建配置
cc-switch new dev-env -i

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

# 启动 Web 界面并自动打开浏览器
cc-switch web --open

# 在自定义端口启动 Web 界面并启用静默模式
cc-switch web --port 8080 --quiet

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
├── settings.json          # 当前活动配置（空配置模式下被移除）
├── profiles/              # 存储的配置
│   ├── default.json       # 默认配置
│   ├── work.json          # 工作配置
│   ├── personal.json      # 个人配置
│   ├── templates/         # 配置模板
│   │   ├── default.json   # 默认模板（不能删除）
│   │   └── company.json   # 自定义公司模板
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
| `init` | 通过交互式设置初始化 Claude Code 配置 |
| `list` | 列出所有可用配置 |
| `list -t, --template` | 列出所有可用模板 |
| `new <名称>` | 从默认模板创建新配置 |
| `new <名称> -t <模板>` | 从指定模板创建新配置 |
| `new <名称> -i, --interactive` | 使用交互式模板填写创建配置 |
| `use <名称>` | 切换到配置 |
| `use <名称> -l, --launch` | 切换到配置并启动 Claude Code CLI |
| `use -p, --previous` | 切换到上一个配置 |
| `use -e, --empty` | 进入空配置模式（禁用配置） |
| `use --restore` | 从空配置模式恢复到上一个配置 |
| `use -i, --interactive` | 进入交互选择模式 |
| `cp <源> <目标>` | 复制配置 |
| `cp -t <源> <目标>` | 复制模板 |
| `mv <旧> <新>` | 移动（重命名）配置 |
| `mv -t <旧> <新>` | 移动（重命名）模板 |
| `rm <名称>` | 删除配置 |
| `rm -c, --current` | 删除当前配置并进入空配置模式 |
| `rm -a, --all` | 删除所有配置（需要手动确认） |
| `rm -t <模板>` | 删除模板 |
| `export [配置]` | 将配置导出到备份文件 |
| `import <文件>` | 从备份文件导入配置 |
| `test [配置]` | 测试配置 API 连接性 |
| `web` | 启动 Web 界面进行配置管理 |
| `current` | 显示当前配置或空配置模式状态 |
| `view <名称>` | 查看配置详情 |
| `view -t <模板>` | 查看模板详情 |
| `edit <名称>` | 在文本编辑器中编辑配置 |
| `edit -t <模板>` | 在文本编辑器中编辑模板 |

### 模板系统

模板系统允许您创建标准化的配置结构，确保不同环境之间的一致设置。

#### 什么是模板？

模板是预配置的 JSON 结构，用作创建新配置的蓝图。它们可以包括：

- **标准设置**：适用于各种环境的常见 Claude Code 设置
- **占位符字段**：可在配置创建过程中交互式填写的字段
- **环境特定值**：可为不同使用情况定制的默认值

#### 模板功能

- **默认模板**：内置模板，不能删除，提供基本结构
- **自定义模板**：创建适合特定需求的模板（公司、项目等）
- **交互式创建**：在配置创建过程中通过引导提示填写模板字段
- **模板管理**：像配置一样复制、编辑、查看和组织模板

#### 使用模板

1. **列出模板**：`cc-switch list -t` 查看可用模板
2. **从模板创建**：`cc-switch new myconfig -t mytemplate` 使用特定模板
3. **交互模式**：`cc-switch new myconfig -i` 进行引导式模板字段输入
4. **模板管理**：使用带 `-t` 标志的标准操作（cp、mv、edit、view）

#### 模板结构

模板使用与配置相同的 JSON 结构，但可能包含在交互式创建期间填写的特殊占位符值。

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