# Development

Contributing to Kassie and building from source.

## Quick Start

```bash
# Clone repository
git clone https://github.com/kashifKhn/kassie.git
cd kassie

# Setup dependencies
make setup

# Generate code
make proto

# Build
make build

# Run tests
make test
```

## Development Guides

- [Setup](/development/setup) — Development environment and prerequisites
- [Building](/development/building) — Build system and cross-compilation
- [Testing](/development/testing) — Running and writing tests
- [Architecture Decisions](/development/architecture-decisions) — ADRs and design rationale
- [Contributing](/development/contributing) — Contribution guidelines and PR process

## Prerequisites

- Go 1.24+ (tested with 1.24.5)
- Node.js 20+ (for web UI)
- protoc (Protocol Buffer compiler)
- Docker (for integration tests)
- Make

## Get Involved

- GitHub Issues: Report bugs and request features
- Pull Requests: Contribute code
- Discussions: Ask questions and share ideas

Visit the [Contributing Guide](/development/contributing) to learn how to get started.
