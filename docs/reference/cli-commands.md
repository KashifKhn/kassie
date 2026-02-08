# CLI Commands

Complete reference for all Kassie command-line interfaces.

## Global Options

These flags work with all commands:

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--config` | `-c` | string | Path to config file (default: `~/.config/kassie/config.json`) |
| `--profile` | `-p` | string | Database profile to use |
| `--log-level` | `-l` | string | Log level: `debug`, `info`, `warn`, `error` (default: `info`) |
| `--version` | `-v` | boolean | Print version information and exit |
| `--help` | `-h` | boolean | Show help message |

**Examples**:
```bash
# Use custom config
kassie tui --config /path/to/config.json

# Set log level
kassie tui --log-level debug

# Use specific profile
kassie tui --profile production
```

## Commands

### `kassie tui`

Launch the Terminal User Interface.

**Usage**:
```bash
kassie tui [options]
```

**Options**:

| Flag | Type | Description |
|------|------|-------------|
| `--server` | string | Connect to remote Kassie server (format: `host:port`) |

**Examples**:
```bash
# Launch TUI with default settings
kassie tui

# Connect to remote server
kassie tui --server remote.example.com:50051

# Use specific profile
kassie tui --profile production

# Debug mode
kassie tui --log-level debug
```

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Connection error

---

### `kassie web`

Launch the Web User Interface.

**Usage**:
```bash
kassie web [options]
```

**Options**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | integer | 8080 | HTTP port for web server |
| `--no-browser` | boolean | false | Don't auto-open browser |
| `--web-root` | string | - | Serve static files from directory (dev mode) |

**Examples**:
```bash
# Launch web UI on default port (8080)
kassie web

# Use custom port
kassie web --port 3000

# Don't open browser automatically
kassie web --no-browser

# Development mode with custom web root
kassie web --web-root ./web/dist
```

**Access**:
- Default URL: `http://localhost:8080`
- Custom port: `http://localhost:<port>`

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Configuration error
- `4`: Port already in use

---

### `kassie server`

Run Kassie in standalone server mode.

**Usage**:
```bash
kassie server [options]
```

**Options**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--grpc-port` | integer | 50051 | gRPC server port |
| `--http-port` | integer | 8080 | HTTP gateway port |
| `--host` | string | `0.0.0.0` | Bind address |

**Examples**:
```bash
# Run server with default ports
kassie server

# Custom ports
kassie server --grpc-port 9090 --http-port 8888

# Bind to specific interface
kassie server --host 192.168.1.100

# Production mode
kassie server --log-level warn
```

**Endpoints**:
- gRPC: `<host>:<grpc-port>`
- HTTP: `http://<host>:<http-port>`
- Health check: `http://<host>:<http-port>/health`

**Signals**:
- `SIGINT` / `SIGTERM`: Graceful shutdown
- `SIGKILL`: Force shutdown (not recommended)

---

### `kassie version`

Print version information.

**Usage**:
```bash
kassie version
```

**Output**:
```
Kassie v1.0.0
Commit: abc1234def5678
Built: 2024-01-15T10:30:00Z
```

**Programmatic access**:
```bash
# Get just version number
kassie version | head -1 | awk '{print $2}'
```

---

### `kassie help`

Show help information.

**Usage**:
```bash
kassie help [command]
```

**Examples**:
```bash
# General help
kassie help

# Help for specific command
kassie help tui
kassie help web
kassie help server
```

## Environment Variables

Kassie recognizes these environment variables:

| Variable | Description |
|----------|-------------|
| `KASSIE_CONFIG` | Config file path (overrides `--config`) |
| `KASSIE_PROFILE` | Default profile (overrides `--profile`) |
| `KASSIE_LOG_LEVEL` | Log level (overrides `--log-level`) |
| `KASSIE_JWT_SECRET` | JWT secret for server mode |

**Example**:
```bash
export KASSIE_CONFIG=~/.kassie.json
export KASSIE_PROFILE=production
export KASSIE_LOG_LEVEL=debug
kassie tui
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Connection error |
| 4 | Port/resource unavailable |
| 5 | Permission denied |

## Common Patterns

### Quick connection
```bash
kassie tui --profile local
```

### Development with debug logging
```bash
kassie tui --log-level debug
```

### Production server
```bash
kassie server \
  --host 0.0.0.0 \
  --grpc-port 50051 \
  --http-port 8080 \
  --log-level info
```

### Team web access
```bash
kassie web --port 8080 --no-browser
```

### Remote TUI connection
```bash
kassie tui --server prod-kassie.example.com:50051
```

## Scripting

Kassie can be used in scripts:

```bash
#!/bin/bash

# Check version
VERSION=$(kassie version | head -1 | awk '{print $2}')
echo "Using Kassie $VERSION"

# Start server in background
kassie server --log-level warn > kassie.log 2>&1 &
PID=$!

# Wait for server to start
sleep 2

# Do work...

# Stop server
kill $PID
```

## Next Steps

- [Configuration Schema](/reference/configuration-schema) - Detailed config reference
- [Keyboard Shortcuts](/reference/keyboard-shortcuts) - All shortcuts
- [API Reference](/reference/api) - Programmatic access
