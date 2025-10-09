# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **Claude Code Configuration Switcher** tool that allows users to manage and switch between different Claude Code configurations. The project uses a **hybrid architecture** combining Go for the core CLI implementation and Node.js for npm distribution and web interface assets.

## Architecture

### Core Structure
- **Go Backend**: Main CLI application using Cobra framework
- **Node.js Wrapper**: npm package distribution and installation scripts
- **Web Interface**: Browser-based configuration management with static assets
- **Configuration Management**: Manages Claude Code profiles in `~/.claude/` directory

### Key Packages
- `cmd/`: CLI command implementations (list, new, use, rm, export, import, test, web, etc.)
- `internal/config/`: Configuration management logic and data structures
- `internal/handler/`: Business logic handlers including API testing
- `internal/ui/`: User interface layer including interactive prompts
- `internal/web/`: Web interface server and handlers
- `internal/export/` & `internal/import/`: Backup/restore functionality
- `internal/common/`: Shared utilities (compression, crypto)

## Development Commands

### Build Commands
```bash
# Build for all platforms (creates binaries for darwin, linux, windows)
npm run build

# Build Go binaries only
npm run build:go

# Clean build artifacts
npm run clean
```

### Testing
```bash
# Run Go tests
npm test
# or
go test ./...

# Test locally during development
go run . --help
```

### Development Workflow
```bash
# Install dependencies
go mod tidy

# Run locally without building
go run . list
go run . --help

# Test specific functionality
go run . test --current
go run . web --port 8080
```

## Key Implementation Details

### Configuration System
- **Profile Storage**: `~/.claude/profiles/` - individual configuration files
- **Template System**: `~/.claude/profiles/templates/` - configuration templates
- **State Management**: Uses atomic operations with temporary files for safety
- **Empty Mode**: Special state for temporarily disabling all configurations

### Cross-Platform Distribution
- **Build Targets**: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- **npm Integration**: Post-install script selects appropriate binary for platform
- **Binary Location**: Platform-specific binaries stored in `bin/` directory

### Web Interface
- **Server**: Go HTTP server serving static assets and API endpoints
- **Assets**: CSS/JS files in `internal/web/assets/`
- **Features**: Profile management, template CRUD, API testing, live configuration editing

### Security Features
- **File Permissions**: Configuration files stored with 600 permissions
- **Atomic Operations**: Temporary files and atomic renames prevent corruption
- **Backup Systems**: Automatic backup before configuration changes
- **Input Validation**: Path traversal protection and secure operations

## Common Development Scenarios

### Adding New CLI Commands
1. Create new command file in `cmd/` (e.g., `cmd/newfeature.go`)
2. Implement command using Cobra patterns from existing commands
3. Register command in `cmd/root.go` init function
4. Add business logic in appropriate `internal/` package

### Modifying Configuration Management
- Core logic in `internal/config/manager.go`
- Configuration structures in `internal/handler/types.go`
- Interactive UI components in `internal/ui/`

### Web Interface Changes
- Server logic: `internal/web/server.go` and `internal/web/handlers.go`
- Frontend assets: `internal/web/assets/css/main.css` and `internal/web/assets/js/main.js`

### Testing Configuration Changes
```bash
# Test specific configuration
go run . test myconfig

# Test current configuration  
go run . test --current

# Test all configurations
go run . test --all

# Launch web interface for testing
go run . web --open
```

## Build and Release Process

### Local Development
- Use `go run .` for quick testing
- Use `npm run build` to test cross-platform builds
- Web interface available at `http://localhost:13501` by default

### Release Process
- GitHub Actions workflow in `.github/workflows/release.yml`
- Automated building for all supported platforms
- npm package publishing with platform-specific binary selection

## Project Dependencies

### Go Dependencies
- `github.com/spf13/cobra`: CLI framework
- `github.com/manifoldco/promptui`: Interactive prompts
- `github.com/fatih/color`: Terminal colors
- `golang.org/x/crypto`: Encryption for export/import
- `golang.org/x/term`: Terminal utilities

### Node.js Components
- Platform detection and binary selection in `scripts/install.js`
- npm package configuration in `package.json`
- Cross-platform build script in `scripts/build.sh`