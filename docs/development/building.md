# Building from Source

This guide covers Kassie's build process, from protobuf generation to producing the final single binary with embedded web assets.

## Build Overview

The full build pipeline:

```
Proto Definitions ──> Generated Go Code ──> 
React Source ──> Web Build (dist/) ──> Embed into Go ──> Go Binary
```

## Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Full binary with embedded web UI |
| `make build-server` | Server-only binary (no web assets) |
| `make web` | Build web UI only → `web/dist/` |
| `make proto` | Generate Go code from protobuf |
| `make embed-web` | Copy web assets for embedding |
| `make clean` | Remove all build artifacts |

## Full Build

Build the complete Kassie binary:

```bash
make build
```

This executes three steps in sequence:

### 1. Build Web UI (`make web`)

```bash
cd web && pnpm install && pnpm run build
```

Produces optimized React bundle in `web/dist/`:
- Minified JavaScript with chunk splitting
- CSS with PostCSS processing
- Static assets with content hashing

### 2. Embed Web Assets (`make embed-web`)

```bash
# Copies web/dist/* → internal/server/web/dist/
cp -r web/dist/* internal/server/web/dist/
```

Go's `embed.FS` packages these files into the binary at compile time:

```go
//go:embed dist/*
var webAssets embed.FS
```

### 3. Compile Go Binary

```bash
go build -o kassie cmd/kassie/main.go
```

The output is a single `kassie` binary containing:
- Go server, TUI, and CLI code
- Embedded web UI assets
- All dependencies statically linked

After compilation, copied web assets are cleaned up from `internal/server/web/dist/`.

## Server-Only Build

Build without web UI (smaller binary):

```bash
make build-server
```

```bash
go build -tags=noui -o kassie cmd/kassie/main.go
```

The `noui` build tag excludes web asset embedding.

## Protobuf Code Generation

Generate Go code from `.proto` files:

```bash
make proto
```

**Input:** `api/proto/*.proto` (4 files)

**Output:** `api/gen/go/` containing:
- Message types (Go structs)
- gRPC service interfaces
- gRPC-gateway reverse proxy handlers

::: tip
Always regenerate after modifying proto files. The generated code is checked into the repository for CI builds.
:::

## Cross-Compilation

Build for different platforms using Go's cross-compilation:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o kassie-linux-amd64 cmd/kassie/main.go

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o kassie-darwin-arm64 cmd/kassie/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o kassie-windows-amd64.exe cmd/kassie/main.go
```

### Supported Platforms

| Platform | Architecture | Binary Name |
|----------|-------------|-------------|
| Linux | x86_64 | `kassie-linux-amd64` |
| Linux | ARM64 | `kassie-linux-arm64` |
| macOS | Intel | `kassie-darwin-amd64` |
| macOS | Apple Silicon | `kassie-darwin-arm64` |
| Windows | x86_64 | `kassie-windows-amd64.exe` |

## Version Information

Version data is injected at build time via Go linker flags:

```bash
go build -ldflags "-X main.version=v0.2.0 -X main.commit=$(git rev-parse HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" cmd/kassie/main.go
```

Check version:
```bash
kassie version
```

## Release Builds (GoReleaser)

Production releases use [GoReleaser](https://goreleaser.com/) via `.goreleaser.yml`:

```bash
# Create a release (requires a git tag)
git tag v0.2.0
goreleaser release
```

GoReleaser handles:
- Multi-platform binary compilation
- Checksum generation
- GitHub release creation with changelogs
- Archive packaging (tar.gz, zip)

## GitHub Actions CI

Automated builds run on every push via `.github/workflows/build.yml`:

1. Checkout code
2. Setup Go and Node.js
3. Generate protobuf code
4. Build web UI
5. Build Go binary
6. Run tests

Release workflow (`.github/workflows/release.yml`) triggers on new tags.

## Build Troubleshooting

### `make proto` fails

Ensure protoc plugins are installed:
```bash
make setup
```

### `make web` fails

Ensure Node.js 20+ and pnpm are installed:
```bash
node --version   # Should be 20+
pnpm --version   # Should be installed
cd web && pnpm install
```

### Binary too large

The full binary with embedded web assets is ~25-50MB. This is expected due to:
- Embedded web UI (~2-5MB)
- Go runtime
- gRPC libraries
- TUI framework (Bubbletea + Lipgloss)

### Web assets not found

If `kassie web` shows a blank page, ensure web assets were embedded:
```bash
make clean
make build  # Full build with web assets
```
