# Kassie

Modern database explorer for Cassandra & ScyllaDB with TUI and Web interfaces.

## Features

- **Terminal Interface (TUI)** - Stable, production-ready
- **Web UI** - 🚧 Under active development (Phase 5)
- Single binary with embedded server
- Browse keyspaces, tables, and data
- Filter with WHERE clauses
- Keyboard-driven navigation

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/kassie/main/install.sh | sh
```

### Manual Installation

Download the latest release for your platform from [GitHub Releases](https://github.com/KashifKhn/kassie/releases).

### Go Install

```bash
go install github.com/KashifKhn/kassie@latest
```

### Homebrew (Coming Soon)

```bash
# Homebrew tap planned for future release
brew tap KashifKhn/kassie
brew install kassie
```

See [installation docs](https://kashifkhn.github.io/kassie/guide/installation) for more options.

### Upgrade

Keep Kassie up-to-date with the built-in upgrade command:

```bash
kassie upgrade                    # Upgrade to latest version
kassie upgrade --check            # Check for updates only
kassie upgrade --version v0.2.0   # Upgrade to specific version
```

## Usage

```bash
kassie tui              # Launch terminal interface (recommended)
kassie web              # Launch web interface (under development)
kassie server           # Run standalone server
```

## Configuration

Create `~/.config/kassie/config.json`:

```json
{
  "profiles": [
    {
      "name": "local",
      "hosts": ["127.0.0.1"],
      "port": 9042
    }
  ]
}
```

## Development

```bash
make setup              # Install tools
make proto              # Generate gRPC code
make build              # Build binary
make test               # Run tests
```

## License

MIT
