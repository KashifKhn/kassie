# Testing Guide

This guide covers Kassie's testing strategy, how to run tests, and how to write new ones.

## Testing Philosophy

- **Unit tests** for all business logic
- **Integration tests** for gRPC services with real ScyllaDB
- **No UI tests** — TUI and Web components are not tested
- **Table-driven tests** for comprehensive case coverage
- **Interfaces for mockability** — dependencies injected via constructors

## Running Tests

### All Tests

```bash
make test
# or
go test ./...
```

### Unit Tests Only

```bash
make test-unit
# or
go test -short ./...
```

### Integration Tests

Requires Docker for ScyllaDB:

```bash
make test-int
# or
go test -tags=integration ./...
```

### Specific Package

```bash
go test ./internal/server/service/...
```

### Specific Test

```bash
go test -run TestGenerateToken ./internal/server/service/
```

### With Coverage

```bash
go test -cover ./...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Verbose Output

```bash
go test -v ./internal/server/...
```

## Test Coverage Areas

| Package | Test Focus |
|---------|------------|
| `internal/server/service/auth` | Token generation, validation, expiry |
| `internal/server/service/session` | Session lifecycle, profile loading |
| `internal/server/service/schema` | Schema parsing, caching logic |
| `internal/server/service/data` | Query building, pagination, filter parsing |
| `internal/server/db` | Connection management, CQL generation |
| `internal/server/state/store` | Session store operations |
| `internal/server/state/cursor` | Cursor management, expiry |
| `internal/shared/config` | Config loading, override merging, validation |
| `internal/client` | Token refresh logic, error handling |
| `internal/tui/components` | DataGrid and FilterBar logic |

## Writing Tests

### Table-Driven Tests

Kassie uses Go's table-driven test pattern for comprehensive case coverage:

```go
func TestAuthService_GenerateToken(t *testing.T) {
    tests := []struct {
        name      string
        sessionID string
        profile   string
        wantErr   bool
    }{
        {
            name:      "valid token generation",
            sessionID: "session-123",
            profile:   "local",
            wantErr:   false,
        },
        {
            name:      "empty session ID",
            sessionID: "",
            profile:   "local",
            wantErr:   true,
        },
        {
            name:      "empty profile",
            sessionID: "session-123",
            profile:   "",
            wantErr:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewAuthService("test-secret")
            token, err := svc.GenerateToken(tt.sessionID, tt.profile)

            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if token == "" {
                t.Error("expected non-empty token")
            }
        })
    }
}
```

### Mock Dependencies

Use interfaces for dependency injection and mock in tests:

```go
// Define interface where it's used
type DatabaseQuerier interface {
    QueryRows(ctx context.Context, keyspace, table string, pageSize int) ([]Row, error)
}

// Production implementation
type ScyllaQuerier struct {
    session *gocql.Session
}

// Test mock
type MockQuerier struct {
    rows []Row
    err  error
}

func (m *MockQuerier) QueryRows(ctx context.Context, ks, tbl string, ps int) ([]Row, error) {
    return m.rows, m.err
}
```

### Test File Naming

- Unit tests: `*_test.go` adjacent to source file
- Integration tests: use `//go:build integration` build tag

```
internal/server/service/
├── auth.go
├── auth_test.go          # Unit tests
├── data.go
├── data_test.go          # Unit tests
└── session.go
```

## Integration Tests

Integration tests run against a real ScyllaDB instance:

### Setup

Start ScyllaDB with Docker:

```bash
docker-compose up -d
```

### Writing Integration Tests

```go
//go:build integration

package service_test

func TestSchemaService_ListKeyspaces(t *testing.T) {
    // Connect to local ScyllaDB
    cluster := gocql.NewCluster("127.0.0.1")
    session, err := cluster.CreateSession()
    if err != nil {
        t.Skip("ScyllaDB not available")
    }
    defer session.Close()

    svc := NewSchemaService(session)
    keyspaces, err := svc.ListKeyspaces(context.Background())

    if err != nil {
        t.Fatalf("failed to list keyspaces: %v", err)
    }

    // system keyspaces should always exist
    if len(keyspaces) == 0 {
        t.Error("expected at least one keyspace")
    }
}
```

### Running Integration Tests

```bash
# Start database first
docker-compose up -d

# Wait for ScyllaDB to be ready (~30 seconds)
sleep 30

# Run integration tests
make test-int
```

## Linting

Run the Go linter:

```bash
make lint
# or
golangci-lint run
```

## Code Formatting

Format all Go and TypeScript code:

```bash
make fmt
```

This runs:
- `go fmt ./...` for Go files
- `pnpm run format` for TypeScript files (if web UI exists)

## CI Pipeline

Tests run automatically in GitHub Actions on every push:

1. `make setup` — Install tools
2. `make proto` — Generate code
3. `make test` — Run all tests
4. `make lint` — Check code quality
5. `make build` — Verify compilation
