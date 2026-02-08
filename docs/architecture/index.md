# Architecture

Understanding Kassie's design and implementation.

## Overview

Kassie follows a client-server architecture with these key components:

- **Server Core**: gRPC server with HTTP gateway
- **TUI Client**: Bubbletea-based terminal interface
- **Web Client**: React-based browser interface
- **Shared Client SDK**: Common gRPC client wrapper

## Key Design Principles

1. **Dual-Client Equality**: TUI and Web as first-class citizens
2. **Single Binary**: All-in-one executable with embedded assets
3. **Type-Safe Communication**: gRPC with auto-generated clients
4. **Read-Safety First**: Optimized for browsing and observing data
5. **Embedded Server**: Server runs within client process or standalone

## Deployment Modes

### Embedded Mode
Server starts as background goroutine within TUI/Web client process.

### Standalone Mode
Server runs independently, accepting remote client connections.

## Learn More

- [Client-Server Model](/architecture/client-server) - Detailed architecture
- [Authentication](/architecture/authentication) - Auth system design
- [State Management](/architecture/state-management) - State handling
- [Protocol Design](/architecture/protocol) - gRPC protocol details

For the complete architecture document, see [n_docs/ARCHITECTURE.md](https://github.com/KashifKhn/kassie/blob/main/n_docs/ARCHITECTURE.md) in the repository.
