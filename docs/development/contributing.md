# Contributing to Kassie

Thank you for your interest in contributing to Kassie! This guide will help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Assume good intentions

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.24+** - [Download](https://go.dev/dl/)
- **Node.js 20+** - For web UI development
- **protoc** - Protocol Buffer compiler
- **Make** - Build automation
- **Git** - Version control

Optional but recommended:
- **Docker** - For integration tests
- **golangci-lint** - Go linter

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/kassie.git
cd kassie
```

3. Add upstream remote:

```bash
git remote add upstream https://github.com/kashifKhn/kassie.git
```

## Development Setup

### Initial Setup

Install dependencies and generate code:

```bash
# Install protoc plugins and tools
make setup

# Generate gRPC code (Go + TypeScript)
make proto
```

### Build the Project

```bash
# Build full binary with embedded web UI
make build

# Build server only (no web assets)
make build-server

# Build web UI only
make web
```

### Run in Development Mode

```bash
# Run TUI in development mode
make dev-tui

# Run web UI with hot reload
make dev-web

# Run server only
make dev-server
```

## Making Changes

### Branch Naming

Create a branch from `main`:

```bash
git checkout -b <type>/<description>
```

Branch types:
- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions/changes
- `chore/` - Maintenance tasks

Examples:
- `feat/add-pagination-controls`
- `fix/connection-timeout-issue`
- `docs/update-api-reference`

### Development Workflow

1. **Make your changes** in your feature branch
2. **Write tests** for new functionality
3. **Run tests** to ensure nothing breaks:

```bash
make test
```

4. **Format code**:

```bash
make fmt
```

5. **Run linter**:

```bash
make lint
```

6. **Build the project** to verify it compiles:

```bash
make build
```

## Commit Guidelines

We use semantic commit messages:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation only
- `style` - Formatting, missing semicolons, etc.
- `refactor` - Code restructuring without behavior change
- `test` - Adding or updating tests
- `chore` - Maintenance tasks, dependencies
- `ci` - CI/CD changes

### Scope

Optional, indicates what part of the codebase:
- `tui` - Terminal UI
- `web` - Web interface
- `server` - Backend server
- `client` - gRPC client
- `api` - API definitions
- `docs` - Documentation

### Examples

```bash
feat(tui): add row copy to clipboard

fix(server): handle connection timeout properly

docs(api): update authentication examples

test(client): add token refresh tests

chore(deps): update gocql to v1.6.0
```

### Commit Message Rules

- Use imperative mood ("add" not "added")
- No period at the end of subject
- Subject line max 72 characters
- Body wraps at 72 characters
- Reference issues in footer: `Fixes #123`

## Pull Request Process

### Before Submitting

1. **Sync with upstream**:

```bash
git fetch upstream
git rebase upstream/main
```

2. **Run full test suite**:

```bash
make test
make lint
```

3. **Update documentation** if needed

4. **Test manually** in both TUI and Web (if applicable)

### Submitting PR

1. Push your branch:

```bash
git push origin <branch-name>
```

2. Open PR on GitHub with:
   - Clear title following commit conventions
   - Description of changes
   - Screenshots (for UI changes)
   - Link to related issues

3. Fill out the PR template

### PR Requirements

- ✅ All tests pass
- ✅ Linter passes
- ✅ No merge conflicts
- ✅ Documentation updated
- ✅ Commit messages follow guidelines

### Review Process

1. Maintainers will review your PR
2. Address feedback in new commits
3. Once approved, PR will be squash merged
4. Your commits will be combined into one

## Code Style

### Go Code Style

**General Principles:**
- Self-documenting code through clear naming
- No comments unless absolutely necessary
- Small focused functions (max ~50 lines)
- Max 300 lines per file

**Naming:**
```go
// Packages: lowercase single word
package service  // not "services"

// Interfaces: verb-based
type Reader interface {}  // not "IReader"

// Functions: verb-noun
func GetUser() {}  // not "UserGet"

// Variables: camelCase
var userName string

// Exported: PascalCase
type User struct {}

// Unexported: camelCase
type session struct {}
```

**Error Handling:**
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to connect: %w", err)
}

// Early returns
if invalid {
    return ErrInvalidInput
}
```

**Structure Example:**
```go
type Service struct {
    db     Database
    logger Logger
}

func NewService(db Database, logger Logger) *Service {
    return &Service{db: db, logger: logger}
}

func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, ErrInvalidID
    }
    // implementation
}
```

### TypeScript Code Style

**Naming:**
```typescript
// Components: PascalCase
Sidebar.tsx

// Hooks: useCamelCase
useSession.ts

// Utilities: camelCase
formatDate()

// Types/Interfaces: PascalCase
interface UserProfile {}
```

**Patterns:**
- One component per file
- Custom hooks for logic extraction
- Functional components only
- Strict null checks enabled

## Testing

### Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests (requires Docker)
make test-int

# Specific package
go test ./internal/server/service/...

# Specific test
go test -run TestGenerateToken ./internal/server/service/

# With coverage
go test -cover ./...

# Verbose
go test -v ./internal/server/...
```

### Writing Tests

**Test Coverage Required:**
- `internal/server/service/*` - All service methods
- `internal/server/db/*` - Query building, connection management
- `internal/server/state/*` - Session and cursor stores
- `internal/shared/config/*` - Config loading and merging
- `internal/client/*` - Token refresh, error handling

**Test Style:**
```go
func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name    string
        userID  string
        want    *User
        wantErr bool
    }{
        {
            name:    "valid user",
            userID:  "123",
            want:    &User{ID: "123"},
            wantErr: false,
        },
        {
            name:    "empty id",
            userID:  "",
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

**Test File Naming:**
- `*_test.go` for unit tests
- `//go:build integration` tag for integration tests

## Documentation

### When to Update Docs

Update documentation when you:
- Add new features
- Change CLI commands
- Modify configuration options
- Update API endpoints
- Change keyboard shortcuts

### Documentation Location

- **User docs**: `docs/` directory (VitePress)
- **API docs**: `docs/reference/api.md`
- **Code comments**: Only for complex algorithms

### Building Docs Locally

```bash
cd docs
npm install
npm run dev
```

Visit `http://localhost:5173` to preview.

### Documentation Standards

- Use clear, concise language
- Include code examples
- Add screenshots for UI changes
- Keep navigation up to date
- Test all internal links

## Getting Help

### Resources

- **GitHub Issues**: Bug reports and feature requests
- **Discussions**: Questions and general discussion
- **Documentation**: https://kassie.kashifkhan.dev

### Asking Questions

When asking for help:
1. Search existing issues first
2. Provide context (OS, Go version, etc.)
3. Include error messages
4. Share minimal reproduction steps
5. Be patient and respectful

## Areas to Contribute

### Good First Issues

Look for issues labeled:
- `good first issue` - Beginner friendly
- `help wanted` - Extra attention needed
- `documentation` - Doc improvements

### High Priority

Current focus areas:
- Web UI completion (Phase 5)
- Additional themes for TUI
- Performance optimizations
- Test coverage improvements
- Documentation enhancements

### Feature Requests

Before implementing new features:
1. Check if issue exists
2. Create feature request issue
3. Discuss approach with maintainers
4. Wait for approval before coding

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors will be:
- Listed in release notes
- Credited in CONTRIBUTORS file
- Mentioned in documentation (for significant contributions)

Thank you for contributing to Kassie! 🚀
