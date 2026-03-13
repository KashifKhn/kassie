# Development Setup

This guide walks through setting up a complete development environment for Kassie.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/dl/) | 1.24+ | Server, TUI, CLI |
| [Node.js](https://nodejs.org/) | 20+ | Web UI |
| [pnpm](https://pnpm.io/) | Latest | Web UI package manager |
| [protoc](https://grpc.io/docs/protoc-installation/) | 3.x+ | Protocol Buffer compiler |
| [Docker](https://www.docker.com/) | Latest | Integration tests, local database |
| [Make](https://www.gnu.org/software/make/) | Any | Build automation |

## Getting the Source

```bash
git clone https://github.com/KashifKhn/kassie.git
cd kassie
```

## Initial Setup

Install protoc plugins and development tools:

```bash
make setup
```

This installs:
- `protoc-gen-go` — Go code generation
- `protoc-gen-go-grpc` — gRPC service stubs
- `protoc-gen-grpc-gateway` — REST gateway generation
- `protoc-gen-openapiv2` — OpenAPI spec generation
- `golangci-lint` — Go linter

## Generate Code

Generate Go code from protobuf definitions:

```bash
make proto
```

This runs `scripts/gen-proto.sh` which generates files into `api/gen/go/`.

::: warning
Always run `make proto` after modifying any `.proto` file in `api/proto/`.
:::

## Build the Project

```bash
# Full binary with embedded web UI
make build

# Server only (without web assets)
make build-server

# Web UI only
make web
```

## Local Database

Start a local ScyllaDB instance with Docker Compose:

```bash
docker-compose up -d
```

The `docker-compose.yml` provides a single-node ScyllaDB cluster for development. Connect with the `local` profile:

```json
{
  "profiles": [{
    "name": "local",
    "hosts": ["127.0.0.1"],
    "port": 9042
  }]
}
```

## Development Modes

### TUI Development

Run the TUI directly with the Go toolchain:

```bash
make dev-tui
# or
go run cmd/kassie/main.go tui
```

### Web UI Development

Requires **two terminals** for hot-reload:

**Terminal 1** — Go server on port 9090:
```bash
make dev-server
```

**Terminal 2** — Vite dev server with proxy:
```bash
make dev-web
# or
cd web && pnpm dev
```

Vite proxies all `/api/` requests to the Go server. Changes to React code hot-reload instantly.

### Server Development

Run the server standalone:

```bash
make dev-server
# or
go run cmd/kassie/main.go server --http-port 9090
```

## IDE Setup

### VS Code

Recommended extensions:

- **Go** (`golang.go`) — Go language support
- **vscode-proto3** — Protobuf syntax highlighting
- **ESLint** — TypeScript linting
- **Tailwind CSS IntelliSense** — CSS class autocomplete

Workspace settings (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true
}
```

### GoLand / IntelliJ

- Enable the Go plugin
- Set Go SDK to 1.24+
- Enable Protobuf support via plugin
- Use the built-in terminal for Make commands

## Project Structure

```
kassie/
├── api/proto/          # Protobuf definitions
├── api/gen/go/         # Generated Go code
├── cmd/kassie/         # CLI entry point
├── internal/
│   ├── cli/            # Cobra commands (root, tui, web, server)
│   ├── client/         # gRPC client wrapper
│   ├── server/         # Server core
│   │   ├── db/         # Database connection & queries
│   │   ├── gateway/    # HTTP gateway
│   │   ├── grpc/       # gRPC server & interceptors
│   │   ├── service/    # Business logic
│   │   ├── state/      # Session & cursor stores
│   │   └── web/        # Embedded web assets
│   ├── shared/         # Config, logger, utils
│   └── tui/            # Bubbletea TUI
│       ├── components/ # Sidebar, DataGrid, Inspector, etc.
│       ├── views/      # Connection, Explorer, Help
│       └── styles/     # Theme definitions
├── web/                # React web client
│   └── src/
│       ├── api/        # API client, types, schemas
│       ├── components/ # React components
│       ├── pages/      # Login, Explorer, NotFound
│       └── stores/     # Zustand stores
├── docs/               # VitePress documentation
├── scripts/            # Build & generation scripts
└── Makefile            # Build automation
```

## Common Development Tasks

| Task | Command |
|------|---------|
| Run all tests | `make test` |
| Run unit tests only | `make test-unit` |
| Run integration tests | `make test-int` |
| Lint code | `make lint` |
| Format code | `make fmt` |
| Clean build artifacts | `make clean` |
| Full build | `make build` |

## Next Steps

- [Building from Source](/development/building) — Detailed build process
- [Testing Guide](/development/testing) — Writing and running tests
- [Contributing](/development/contributing) — Contribution workflow
