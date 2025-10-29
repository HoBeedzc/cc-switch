# Contributing to cc-switch

Thank you for your interest in contributing to cc-switch! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing](#testing)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Issues](#reporting-issues)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept feedback gracefully
- Prioritize the community's best interests

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go**: Version 1.23.0 or higher
- **Node.js**: Version 14.0.0 or higher
- **Git**: For version control
- **A text editor**: VS Code, GoLand, or your preferred editor

### First-Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/cc-switch.git
   cd cc-switch
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/HoBeedzc/cc-switch.git
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   go mod tidy
   ```

5. **Verify your setup**:
   ```bash
   go run . --help
   ```

## Development Setup

### Build the Project

```bash
# Build for all platforms
npm run build

# Build Go binaries only
npm run build:go

# Clean build artifacts
npm run clean
```

### Run Locally

```bash
# Run without building
go run . --help
go run . list
go run . web

# Run specific commands for testing
go run . test --current
```

### Development Scripts

Available npm scripts:

- `npm run dev`: Run the application in development mode
- `npm test`: Run Go tests
- `npm run fmt`: Format Go code
- `npm run vet`: Run Go vet for static analysis
- `npm run clean`: Clean build artifacts

## Project Structure

Understanding the project structure will help you navigate the codebase:

```
cc-switch/
├── cmd/                    # CLI command implementations
│   ├── root.go            # Root command and CLI setup
│   ├── list.go            # List configurations
│   ├── new.go             # Create new configuration
│   ├── use.go             # Switch configurations
│   ├── test.go            # Test API connectivity
│   ├── web.go             # Web interface server
│   └── ...                # Other commands
├── internal/
│   ├── config/            # Configuration management
│   ├── handler/           # Business logic handlers
│   ├── ui/                # User interface layer
│   ├── web/               # Web interface
│   ├── export/            # Export functionality
│   ├── import/            # Import functionality
│   └── common/            # Shared utilities
├── scripts/               # Build and installation scripts
├── .github/workflows/     # CI/CD workflows
└── npm/                   # npm package distribution
```

### Key Packages

- **cmd/**: Cobra-based CLI commands
- **internal/config/**: Core configuration management logic
- **internal/handler/**: Business logic and API testing
- **internal/ui/**: Interactive prompts and CLI UI
- **internal/web/**: Web interface server and handlers
- **internal/common/**: Utilities (compression, crypto, version)

## Development Workflow

### Creating a Feature Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
```

### Making Changes

1. **Write code** following the style guidelines below
2. **Test your changes** thoroughly
3. **Format code**: Run `npm run fmt`
4. **Run static analysis**: Run `npm run vet`
5. **Run tests**: Run `npm test`

### Keeping Your Fork Updated

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
```

## Code Style Guidelines

### Go Code Style

Follow standard Go conventions and best practices:

#### Formatting

- Use `gofmt` for automatic formatting (run `npm run fmt`)
- Use tabs for indentation (gofmt default)
- Line length should typically not exceed 100 characters

#### Naming Conventions

- **Packages**: Short, lowercase, single-word names
- **Files**: Lowercase with underscores (e.g., `config_handler.go`)
- **Functions**: CamelCase (exported) or camelCase (unexported)
- **Variables**: CamelCase or camelCase, descriptive names
- **Constants**: CamelCase or UPPER_CASE for exported constants

#### Code Organization

- Group related functionality in the same file
- Keep functions focused and single-purpose
- Use interfaces for abstraction where appropriate
- Document all exported functions, types, and constants

#### Error Handling

- Always check and handle errors
- Provide meaningful error messages with context
- Use `fmt.Errorf` for error wrapping
- Don't panic except in truly exceptional cases

Example:
```go
if err := someOperation(); err != nil {
    return fmt.Errorf("failed to perform operation: %w", err)
}
```

#### Comments

- Use complete sentences with proper punctuation
- Document all exported symbols
- Add package-level documentation in `doc.go` or package file
- Explain "why" rather than "what" in implementation comments

Example:
```go
// ConfigManager handles all configuration-related operations including
// loading, saving, and switching between different Claude Code profiles.
type ConfigManager struct {
    // ...
}
```

### Web Interface Guidelines

#### CSS

- Use semantic class names
- Follow BEM methodology when appropriate
- Maintain consistent spacing and indentation
- Group related styles together

#### JavaScript

- Use modern ES6+ features
- Handle errors gracefully with try-catch
- Provide user feedback for all operations
- Keep functions small and focused

## Testing

### Writing Tests

We aim for comprehensive test coverage. When adding new features:

1. **Create test files** named `*_test.go` in the same package
2. **Write table-driven tests** where applicable
3. **Test edge cases** and error conditions
4. **Mock external dependencies** when necessary

Example test structure:

```go
func TestConfigManager_LoadConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {
            name:    "valid config",
            input:   "testdata/valid.json",
            want:    &Config{...},
            wantErr: false,
        },
        {
            name:    "invalid config",
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

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/config/
```

### Test Coverage Goals

- **Critical packages** (config, handler): 80%+ coverage
- **Utility packages** (common): 70%+ coverage
- **CLI commands**: Basic integration tests

## Commit Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semi-colons, etc.)
- `refactor`: Code refactoring without feature changes
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build, etc.)
- `ci`: CI/CD changes

### Examples

```bash
# Feature addition
git commit -m "feat(web): add configuration comparison feature"

# Bug fix
git commit -m "fix(config): resolve race condition in config switching"

# Documentation
git commit -m "docs: update installation instructions for Windows"

# Refactoring
git commit -m "refactor(handler): extract common validation logic"

# Multiple changes (use body)
git commit -m "feat(export): add compression support

- Implement gzip compression for export files
- Add --compress flag to export command
- Update documentation for new flag"
```

### Commit Best Practices

- Keep commits atomic (one logical change per commit)
- Write clear, descriptive commit messages
- Reference issue numbers when applicable (e.g., "fixes #123")
- Avoid commits with too many unrelated changes

## Pull Request Process

### Before Submitting

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   npm run fmt
   npm run vet
   npm test
   ```

3. **Test manually** if applicable

4. **Update documentation** if you changed functionality

### Submitting a Pull Request

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Open a PR** on GitHub with:
   - Clear title following commit conventions
   - Detailed description of changes
   - Reference to related issues
   - Screenshots/demos for UI changes

3. **PR Template** (example):
   ```markdown
   ## Description
   Brief description of what this PR does

   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update

   ## Related Issues
   Fixes #123

   ## Testing
   - [ ] Tests added/updated
   - [ ] Manual testing completed

   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No new warnings generated
   ```

### Review Process

- Maintainers will review your PR
- Address feedback by pushing new commits
- Once approved, maintainers will merge your PR
- Squash commits may be used for cleaner history

### PR Guidelines

- Keep PRs focused and reasonably sized
- One feature/fix per PR
- Respond to review comments promptly
- Be open to feedback and suggestions

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- **Description**: Clear description of the bug
- **Steps to reproduce**: Detailed steps to reproduce the issue
- **Expected behavior**: What should happen
- **Actual behavior**: What actually happens
- **Environment**:
  - OS (macOS, Linux, Windows)
  - Go version (`go version`)
  - Node.js version (`node --version`)
  - cc-switch version (`cc-switch --version`)
- **Logs/Screenshots**: Any relevant error messages or screenshots

### Feature Requests

When requesting features, include:

- **Use case**: Describe the problem you're trying to solve
- **Proposed solution**: Your idea for solving it
- **Alternatives**: Other solutions you've considered
- **Additional context**: Any other relevant information

### Issue Labels

We use labels to categorize issues:

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Documentation improvements
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention needed
- `question`: Further information requested

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality additions
- **PATCH**: Backward-compatible bug fixes

### Release Steps (for maintainers)

1. Update version in `package.json`
2. Update `CHANGELOG.md`
3. Create a git tag: `git tag v1.x.x`
4. Push tag: `git push origin v1.x.x`
5. GitHub Actions will automatically build and publish

### Pre-release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Version bumped appropriately
- [ ] Breaking changes documented

## Getting Help

If you need help:

1. **Check documentation**: README.md and existing issues
2. **Ask questions**: Open a GitHub issue with the `question` label
3. **Join discussions**: Participate in GitHub Discussions (if enabled)

## Recognition

Contributors will be recognized in:

- GitHub contributors page
- Release notes for significant contributions
- Special mentions in CHANGELOG.md

## License

By contributing to cc-switch, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to cc-switch! Your efforts help make this tool better for everyone.
