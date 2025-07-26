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

# Show current configuration
cc-switch current

# Delete unused configuration
cc-switch delete old-config
```

### How It Works

The tool manages configurations in `~/.claude/profiles/` directory:

```
~/.claude/
├── settings.json          # Current active configuration
├── profiles/              # Stored configurations
│   ├── default.json       # Default configuration
│   ├── work.json          # Work configuration
│   └── personal.json      # Personal configuration
└── .current              # Current configuration marker
```

#### Initialization

On first run:
- If `~/.claude/settings.json` exists, it creates a `default` profile
- If no configuration exists, it guides you to create one first

#### Security

- Configuration files are stored with 600 permissions (owner read/write only)
- Atomic operations using temporary files ensure configuration integrity
- Automatic backup of current configuration before switching

### Commands Reference

| Command | Description |
|---------|-------------|
| `list` | List all available configurations |
| `new <name>` | Create a new configuration |
| `use <name>` | Switch to a configuration |
| `delete <name>` | Delete a configuration |
| `current` | Show current configuration |

### Requirements

- Node.js 14.0.0 or higher
- Claude Code installed and configured

### Development

#### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/cc-switch.git
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
│   └── current.go
├── internal/config/        # Configuration management
│   └── manager.go
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

1. Check the [Issues](https://github.com/yourusername/cc-switch/issues) page
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

# 显示当前配置
cc-switch current

# 删除不用的配置
cc-switch delete old-config
```

### 工作原理

工具在 `~/.claude/profiles/` 目录中管理配置：

```
~/.claude/
├── settings.json          # 当前活动配置
├── profiles/              # 存储的配置
│   ├── default.json       # 默认配置
│   ├── work.json          # 工作配置
│   └── personal.json      # 个人配置
└── .current              # 当前配置标记文件
```

#### 初始化

首次运行时：
- 如果 `~/.claude/settings.json` 存在，会创建一个 `default` 配置
- 如果没有配置存在，会引导用户先创建配置

#### 安全性

- 配置文件使用 600 权限存储（仅所有者可读写）
- 使用临时文件的原子操作确保配置完整性
- 切换前自动备份当前配置

### 命令参考

| 命令 | 说明 |
|------|------|
| `list` | 列出所有可用配置 |
| `new <名称>` | 创建新配置 |
| `use <名称>` | 切换到配置 |
| `delete <名称>` | 删除配置 |
| `current` | 显示当前配置 |

### 系统要求

- Node.js 14.0.0 或更高版本
- 已安装并配置 Claude Code

### 开发

#### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/yourusername/cc-switch.git
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
│   └── current.go
├── internal/config/        # 配置管理
│   └── manager.go
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

1. 查看 [Issues](https://github.com/yourusername/cc-switch/issues) 页面
2. 如果问题尚未报告，请创建新 issue
3. 请包含您的操作系统、Node.js 版本和错误信息