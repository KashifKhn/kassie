# Kassie - Agent Guidelines

Kassie is a dual-client (TUI + Web) database explorer for Cassandra/ScyllaDB with client-server architecture.

## Architecture

**Single Binary with Embedded Web UI**: The `make build` command creates a unified binary that embeds the web UI assets. When running `kassie web`, it starts three servers:
- gRPC server (internal port, auto-assigned)
- HTTP API Gateway (port 9090, serves `/api/v1/*`)
- Web UI server (port 9091, serves static files)

## Build Commands

```bash
# Setup and code generation
make setup              # Install protoc plugins and tools
make proto              # Generate gRPC code (Go + TypeScript)

# Building
make build              # Build full binary with embedded web UI
make build-server       # Build server only (no web assets)
make web                # Build web UI only

# Development
make dev-tui            # Run TUI in development mode
make dev-web            # Run web UI with hot reload (port 9091)
make dev-server         # Run server only (API on port 9090)
```

## Running Kassie

```bash
# Web mode (embedded servers + web UI)
./kassie web                    # Web UI on :9091, API on :9090
./kassie web --web-port 8080    # Custom web port
./kassie web --api-port 8090    # Custom API port
./kassie web --no-browser       # Don't auto-open browser

# TUI mode (terminal interface)
./kassie tui                    # Starts embedded server automatically
./kassie tui --server <addr>    # Connect to remote server

# Server mode (standalone API server)
./kassie server                 # gRPC on :50051, HTTP on :8080
./kassie server --http-port 9090 --grpc-port 50051
```

## Test Commands

```bash
# Run all tests
make test
go test ./...

# Run unit tests only
make test-unit
go test -short ./...

# Run single test file
go test ./internal/server/service/auth_test.go

# Run single test function
go test -run TestGenerateToken ./internal/server/service/...

# Run tests with verbose output
go test -v ./internal/server/...

# Run integration tests (requires Docker)
make test-int
go test -tags=integration ./...

# Web client tests
cd web && npm test
cd web && npm test -- --watch
```

## Lint & Format

```bash
# Go
make lint               # Run golangci-lint
make fmt                # Format with gofmt
go fmt ./...
golangci-lint run

# TypeScript (web/)
cd web && npm run lint
cd web && npm run format
```

## Code Style Guidelines

### General Principles

- No comments unless absolutely necessary for complex algorithms
- No doc comments (godoc/jsdoc) unless public API
- Self-documenting code through clear naming
- Small focused functions (max ~50 lines)
- Max 300 lines per file
- Test files adjacent to source files (e.g., `auth.go` and `auth_test.go`)

### Go Standards

**Imports**: Group in order: stdlib, external, internal. Blank line between groups.

**Naming**:

- Packages: lowercase single word (`service` not `services`)
- Interfaces: verb-based (`Reader` not `IReader`)
- Functions: verb-noun (`GetUser` not `UserGet`)
- Variables: camelCase, descriptive
- Exported: PascalCase
- Unexported: camelCase

**Error Handling**:

- Return errors, never panic for expected conditions
- Wrap errors with context: `fmt.Errorf("failed to connect: %w", err)`
- Use sentinel errors for known conditions
- Early returns for error cases

**Patterns**:

- Context as first parameter
- Dependency injection via constructors
- Interfaces defined where used (consumer side)
- Functional options for complex initialization

**Structure**:

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

### TypeScript Standards (web/)

**Naming**:

- Components: `PascalCase` (Sidebar.tsx)
- Hooks: `useCamelCase` (useSession.ts)
- Utilities: `camelCase`
- Types/Interfaces: `PascalCase`

**Structure**:

- One component per file
- Hooks in dedicated files under `hooks/`
- Shared types in `types.ts`
- Functional components only

**Patterns**:

- Custom hooks for logic extraction
- Discriminated unions for state
- Strict null checks enabled
- Handle promise rejections explicitly

## Testing Requirements

Unit tests required for all business logic. No UI tests (TUI/Web components).

**Required test coverage**:

- `internal/server/service/*` - All service methods
- `internal/server/db/*` - Query building, connection management
- `internal/server/state/*` - Session and cursor stores
- `internal/shared/config/*` - Config loading and merging
- `internal/client/*` - Token refresh, error handling

**Test style**:

- Table-driven tests for multiple cases
- Use interfaces for mockability
- Test file naming: `*_test.go`
- Integration tests tagged with `//go:build integration`

## Directory Structure

```
cmd/kassie/          # CLI entrypoint
api/proto/           # Protobuf definitions
api/gen/             # Generated code (gitignored)
internal/server/     # Server core (grpc, service, db, state)
internal/client/     # gRPC client SDK
internal/tui/        # Bubbletea TUI
internal/web/        # Web asset serving
internal/cli/        # Cobra commands
internal/shared/     # Config, logger, utils
web/src/             # React TypeScript frontend
```

## Key Dependencies

**Go**: gocql, grpc, grpc-gateway, cobra, viper, zerolog, bubbletea, lipgloss, golang-jwt

**TypeScript**: react, tanstack-query, zustand, zod,grpc-web, tailwindcss, shadcn/ui,

## Git Workflow

- Semantic commits: `feat():`, `fix():`, `refactor():`, `test():`, `chore():`
- Squash merge to main
- Main branch always deployable
