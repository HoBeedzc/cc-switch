# Claude Code Configuration Switcher

[English](README.md) | [中文](README.zh.md)

A command-line tool for managing and switching between different Claude Code configurations.

---

### Installation

Install globally via npm:

```bash
npm install -g @hobeeliu/cc-switch
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

# Create and switch immediately after creation
cc-switch new <name> -u
cc-switch new <name> --use
```
Creates a new configuration using template structure. The default template provides a basic structure, and interactive mode allows you to fill in template fields with guided prompts. Use `--use` to automatically switch to the newly created configuration.

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

#### Refresh Current Configuration
```bash
cc-switch use --refresh
# or
cc-switch use -f
```
Refreshes the current configuration by re-applying it. This is useful after manually editing configuration files to ensure changes are properly applied.

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
Export configurations to encrypted backup files (.ccx format). Supports optional password protection.

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

# Provide decryption password via flag (or enter interactively)
cc-switch import backup.ccx -p <password>
```
Import configurations from encrypted backup files. Supports conflict resolution modes, dry-run, and encrypted archives.

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

# Retry on failure
cc-switch test -r 5                 # Retry up to 5 times
cc-switch test -r -1                # Retry infinitely until success
cc-switch test -r 3 --retry-interval 5s  # Retry 3 times with 5s interval
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

#### Update cc-switch
```bash
# Check for updates and prompt for confirmation
cc-switch update

# Automatically update without prompting
cc-switch update -y
cc-switch update --yes

# Only check for updates, don't update
cc-switch update -c
cc-switch update --check
```
Check for new versions and update cc-switch to the latest version. The tool automatically checks for updates once every 24 hours in the background and displays a notice if a new version is available.

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

# Create and switch immediately after creation
cc-switch new staging -u

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

# Refresh current configuration after manual edits
cc-switch use --refresh

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

# Check for updates
cc-switch update -c

# Update to latest version
cc-switch update -y
```

### How It Works

The tool manages configurations in `~/.claude/profiles/` directory:

```
~/.claude/
├── settings.json          # Current active configuration (removed in empty mode)
└── profiles/              # cc-switch data directory
    ├── default.json       # Default configuration
    ├── work.json          # Work configuration
    ├── personal.json      # Personal configuration
    ├── templates/         # Configuration templates
    │   ├── default.json   # Default template (cannot be deleted)
    │   └── company.json   # Custom company template
    ├── .current           # Current configuration marker
    ├── .history           # Configuration switch history
    ├── .update_check      # Update check cache
    ├── .empty_mode        # Empty mode state file (present in empty mode)
    └── .empty_backup_settings.json  # Backup when in empty mode
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
| `new <name> -u, --use` | Create configuration and switch to it immediately |
| `use <name>` | Switch to a configuration |
| `use <name> -l, --launch` | Switch to a configuration and launch Claude Code CLI |
| `use -p, --previous` | Switch to previous configuration |
| `use -e, --empty` | Enter empty mode (disable configurations) |
| `use --restore` | Restore from empty mode to previous configuration |
| `use -f, --refresh` | Refresh current configuration (re-apply) |
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
| `update` | Check for updates and prompt for confirmation |
| `update -y, --yes` | Automatically update without prompting |
| `update -c, --check` | Only check for updates, don't update |

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

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for detailed information on:

- Development setup and workflow
- Code style guidelines
- Testing requirements
- Pull request process
- Reporting issues

Quick start:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

For Chinese version, see [贡献指南 (中文)](CONTRIBUTING.zh.md).

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/hobee/cc-switch/issues) page
2. Create a new issue if your problem isn't already reported
3. Include your OS, Node.js version, and error messages
