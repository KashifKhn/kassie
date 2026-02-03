# Kassie

Modern database explorer for Cassandra & ScyllaDB with TUI and Web interfaces.

## Features

- Dual interface: Terminal (TUI) and Web UI
- Single binary with embedded server
- Browse keyspaces, tables, and data
- Filter with WHERE clauses
- Keyboard-driven navigation

## Installation

```bash
go install github.com/KashifKhn/kassie@latest
```

Or with Homebrew:

```bash
brew tap KashifKhn/kassie
brew install kassie
```

## Usage

```bash
kassie tui              # Launch terminal interface
kassie web              # Launch web interface at localhost:8080
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
