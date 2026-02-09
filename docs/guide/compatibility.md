# Compatibility

Database and platform compatibility information for Kassie.

## Database Compatibility

Kassie is designed to work with Apache Cassandra and ScyllaDB clusters.

### Apache Cassandra

| Version | Status | Notes |
|---------|--------|-------|
| 4.1.x | ✅ Supported | Recommended for production |
| 4.0.x | ✅ Supported | Tested and stable |
| 3.11.x | ✅ Supported | Legacy support |
| 3.0.x | ⚠️ Limited | May work but untested |
| 2.x | ❌ Not Supported | Too old |

**CQL Protocol**: Kassie uses CQL protocol version 4 (via gocql driver)

### ScyllaDB

| Version | Status | Notes |
|---------|--------|-------|
| 5.x | ✅ Supported | Latest features |
| 4.x | ✅ Supported | Tested and stable |
| 3.x | ⚠️ Limited | May work but untested |

**Compatibility Notes**:
- ScyllaDB is Cassandra-compatible and works seamlessly with Kassie
- All core features supported (schema introspection, data browsing, filtering)
- Performance optimizations in ScyllaDB (e.g., tablets) work transparently

### Datastax Astra

| Offering | Status | Notes |
|----------|--------|-------|
| Astra DB | ⚠️ Untested | Cassandra-compatible, should work with SSL config |
| Astra Serverless | ⚠️ Untested | Requires specific SSL/TLS setup |

**Note:** While Astra is Cassandra-compatible, Kassie has not been explicitly tested against Astra deployments. Community feedback welcome!

---

## Operating Systems

Kassie binaries are available for:

| OS | Architecture | Status | Notes |
|----|-------------|--------|-------|
| **Linux** | amd64 | ✅ Tested | Ubuntu 20.04+, Debian 11+, Fedora 35+ |
| Linux | arm64 | ✅ Tested | Raspberry Pi 4, ARM servers |
| **macOS** | amd64 (Intel) | ✅ Tested | macOS 11+ |
| **macOS** | arm64 (Apple Silicon) | ✅ Tested | macOS 11+ (M1, M2, M3) |
| **Windows** | amd64 | ⚠️ Limited | Windows 10+, Windows Terminal recommended |
| Windows | arm64 | ❌ Untested | May work but not officially tested |
| FreeBSD | amd64 | ⚠️ Community | May work (Go supports it) |

---

## Terminal Compatibility

The TUI requires a modern terminal emulator with:
- **UTF-8** encoding support
- **256 colors** or better
- **Mouse support** (optional, keyboard works without it)

### Tested Terminals

#### macOS
| Terminal | Status | Notes |
|----------|--------|-------|
| iTerm2 | ✅ Excellent | Best experience, full mouse support |
| Terminal.app | ✅ Good | Default macOS terminal, works well |
| Alacritty | ✅ Excellent | Fast, GPU-accelerated |
| Kitty | ✅ Excellent | Full feature support |
| Warp | ✅ Good | Modern features, some quirks |

#### Linux
| Terminal | Status | Notes |
|----------|--------|-------|
| GNOME Terminal | ✅ Excellent | Default for Ubuntu/Debian |
| Konsole | ✅ Excellent | KDE default |
| Alacritty | ✅ Excellent | Recommended for performance |
| Kitty | ✅ Excellent | Full feature support |
| xterm | ⚠️ Basic | Works but limited styling |
| st (suckless) | ⚠️ Basic | Minimal, may have rendering issues |
| Terminator | ✅ Good | Supports all features |
| Tilix | ✅ Good | Tiling terminal, works well |

#### Windows
| Terminal | Status | Notes |
|----------|--------|-------|
| Windows Terminal | ✅ Recommended | Modern, full feature support |
| PowerShell ISE | ⚠️ Limited | Basic functionality only |
| CMD | ❌ Poor | Not recommended, use Windows Terminal |
| ConEmu | ✅ Good | Works with proper configuration |
| Mintty (Git Bash) | ✅ Good | Solid compatibility |

#### Remote (SSH/tmux/screen)
| Environment | Status | Notes |
|-------------|--------|-------|
| tmux | ✅ Excellent | Set `TERM=screen-256color` |
| GNU Screen | ✅ Good | Set `TERM=screen-256color` |
| SSH | ✅ Excellent | Ensure UTF-8 and 256 colors |
| Mosh | ✅ Good | Works well, some latency considerations |

### Troubleshooting Terminal Issues

**Broken characters / boxes**:
```bash
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8
```

**No colors / wrong colors**:
```bash
export TERM=xterm-256color
```

**Mouse not working**:
Mouse support is optional. All functionality is accessible via keyboard.

---

## Go Version

Kassie is built with Go 1.24+ (tested with 1.24.5).

**Building from source requires**:
- Go 1.24 or later
- Modules enabled (default in Go 1.16+)

---

## Node.js / npm (For Web UI Development)

::: info Web UI Status
The web UI is currently under development. These requirements are for contributors only.
:::

| Tool | Version | Purpose |
|------|---------|---------|
| Node.js | 20.x LTS | JavaScript runtime |
| npm | 10.x | Package manager |

---

## Docker

Kassie Docker images are based on Alpine Linux and support:

| Platform | Status | Image Tag |
|----------|--------|-----------|
| linux/amd64 | ✅ Supported | `latest`, `vX.Y.Z` |
| linux/arm64 | ✅ Supported | `latest`, `vX.Y.Z` |
| linux/arm/v7 | ❌ Not Built | (32-bit ARM not supported) |

---

## Browser Compatibility (Web UI)

::: warning Development Status
Web UI is under development. Browser requirements are preliminary.
:::

Planned browser support:

| Browser | Version | Status |
|---------|---------|--------|
| Chrome | 90+ | 🚧 Planned |
| Firefox | 88+ | 🚧 Planned |
| Safari | 14+ | 🚧 Planned |
| Edge | 90+ | 🚧 Planned |

---

## Known Limitations

### Database Features

**Not Supported**:
- Materialized views (read-only, cannot create/drop)
- User-defined types (UDTs) display as text
- User-defined functions (UDFs) - no introspection
- Secondary indexes - not shown in schema view
- Triggers - not visible

**Partial Support**:
- Collections (list, set, map) - displayed as JSON strings
- Tuples - displayed as comma-separated values
- Complex WHERE clauses - basic CQL syntax only

### Performance

- Large tables (>100M rows) may have slow initial queries
- Wide rows (>1000 columns) may cause TUI rendering slowness
- Blob columns (>1MB) not optimized for display

---

## Testing

Kassie is tested against:
- **Cassandra**: 4.0.11, 4.1.3
- **ScyllaDB**: 5.2.0
- **Go**: 1.24.5
- **Platforms**: Linux (Ubuntu 22.04), macOS (13+), Windows 11

Test coverage:
- Unit tests: `make test-unit`
- Integration tests: `make test-int` (requires Docker)

---

## Community Support

**Untested Configurations**:
If you successfully run Kassie on an unlisted platform or database version, please [open an issue](https://github.com/kashifKhn/kassie/issues) or [discussion](https://github.com/kashifKhn/kassie/discussions) to share your experience!

**Compatibility Issues**:
Found a compatibility problem? [Report it on GitHub](https://github.com/kashifKhn/kassie/issues/new).

---

## Next Steps

- [Installation](/guide/installation) - Install Kassie
- [Getting Started](/guide/getting-started) - First steps
- [Troubleshooting](/guide/troubleshooting) - Fix common issues
