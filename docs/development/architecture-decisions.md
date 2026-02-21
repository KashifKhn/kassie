# Architecture Decision Records

Key architectural decisions made during the development of Kassie, documented in ADR format.

## ADR-001: gRPC over REST

**Status:** Accepted

**Context:**
Kassie needs a communication protocol between its clients (TUI and Web) and the server. The protocol must support type-safe communication, efficient binary encoding, and code generation for both Go and TypeScript.

**Decision:**
Use gRPC with Protocol Buffers as the primary API protocol, with grpc-gateway providing auto-generated REST endpoints for web clients.

**Consequences:**
- ✅ Type-safe contracts defined in `.proto` files
- ✅ Efficient binary encoding (protobuf) for TUI client
- ✅ Auto-generated REST/JSON API for web client via grpc-gateway
- ✅ Single source of truth for API definitions
- ✅ Code generation for Go server stubs
- ⚠️ Requires protoc toolchain for development
- ⚠️ Additional build step for code generation

---

## ADR-002: Embedded Server Mode

**Status:** Accepted

**Context:**
Users should be able to run Kassie as a single command without managing a separate server process. However, the architecture should also support remote server access for team environments.

**Decision:**
Support both embedded and standalone server modes. In embedded mode, the server runs as a background goroutine within the client process.

**Consequences:**
- ✅ Single binary, single command to start
- ✅ No separate server management for personal use
- ✅ Standalone mode available for team/shared access
- ✅ Same server code runs in both modes
- ⚠️ Embedded mode binds to localhost only (by design)
- ⚠️ Server lifecycle tied to client in embedded mode

---

## ADR-003: JWT for Authentication

**Status:** Accepted

**Context:**
Even in embedded mode, clients and server communicate over gRPC. A lightweight authentication mechanism is needed to bind sessions to database connections.

**Decision:**
Use JWT (JSON Web Tokens) with HMAC-SHA256 signing. Access tokens expire in 1 hour, refresh tokens in 24 hours.

**Consequences:**
- ✅ Stateless token validation — no database lookup needed
- ✅ Consistent auth mechanism across embedded and standalone modes
- ✅ Auto-generated secret in embedded mode (zero config)
- ✅ Refresh token pattern reduces re-login frequency
- ⚠️ Tokens must be stored securely by clients
- ⚠️ Standalone mode requires explicit secret configuration

---

## ADR-004: Bubbletea for TUI

**Status:** Accepted

**Context:**
Kassie needs a responsive, keyboard-driven terminal interface with support for complex layouts (sidebar, data grid, inspector) and modern styling.

**Decision:**
Use the [Charm](https://charm.sh/) ecosystem: Bubbletea for the TUI framework, Lipgloss for styling, and Bubbles for reusable components.

**Consequences:**
- ✅ Elm Architecture (Model → Update → View) provides predictable state management
- ✅ Lipgloss enables modern terminal styling (borders, colors, padding)
- ✅ Active community and ecosystem
- ✅ Pure Go — no CGo dependencies
- ⚠️ Complex component composition requires careful state management
- ⚠️ Limited built-in layout primitives (custom layout logic needed)

---

## ADR-005: React for Web UI

**Status:** Accepted

**Context:**
Kassie needs a modern web interface that mirrors TUI functionality with a responsive, accessible design. The web UI ships embedded in the Go binary.

**Decision:**
Use React 18 with TypeScript (strict mode), Vite for builds, Tailwind CSS for styling, shadcn/ui for accessible component primitives, TanStack Query for server state, and Zustand for UI state.

**Consequences:**
- ✅ Large ecosystem and component library options
- ✅ shadcn/ui provides accessible, customizable primitives built on Radix UI
- ✅ TanStack Query handles caching, deduplication, and background refetching
- ✅ Zustand is minimal (~1KB) with zero boilerplate
- ✅ TypeScript strict mode catches type errors at compile time
- ✅ Vite provides fast builds and excellent HMR
- ⚠️ Bundle size must be monitored (target: <170KB gzipped)
- ⚠️ React is a heavy framework for what is essentially a data browser

---

## ADR-006: Single Binary Distribution

**Status:** Accepted

**Context:**
Kassie should be easy to install and run on any platform without runtime dependencies, package managers, or configuration steps.

**Decision:**
Distribute as a single statically-linked Go binary with web assets embedded via `embed.FS`.

**Consequences:**
- ✅ Zero runtime dependencies — just download and run
- ✅ Cross-platform support (Linux, macOS, Windows, ARM64)
- ✅ `go install` works out of the box
- ✅ Consistent behavior across environments
- ⚠️ Larger binary size (~25-50MB) due to embedded assets
- ⚠️ Web UI updates require rebuilding the entire binary

---

## ADR-007: REST via grpc-gateway over gRPC-Web

**Status:** Accepted

**Context:**
The web client needs to communicate with the server. Two options: use gRPC-Web (requires a proxy or Envoy) or use REST/JSON via grpc-gateway (already part of the server).

**Decision:**
Use REST endpoints generated by grpc-gateway instead of gRPC-Web.

**Consequences:**
- ✅ No gRPC-Web runtime needed (~50KB savings)
- ✅ No Envoy proxy or additional infrastructure
- ✅ Standard HTTP/JSON — easy to debug with browser DevTools and curl
- ✅ grpc-gateway already configured on the server
- ⚠️ Slightly higher latency vs native gRPC (~3-5ms vs ~1-2ms locally)
- ⚠️ Larger payload sizes (JSON vs protobuf binary)
- ⚠️ API types manually defined in TypeScript (not auto-generated)

---

## ADR-008: Cursor-Based Pagination

**Status:** Accepted

**Context:**
Cassandra/ScyllaDB doesn't support `OFFSET`-based pagination. A pagination strategy is needed that works with Scylla's native paging mechanism.

**Decision:**
Use cursor-based pagination built on Scylla's paging state tokens. The server stores paging state per cursor ID and clients request the next page via cursor reference.

**Consequences:**
- ✅ Leverages Scylla's native paging — efficient and consistent
- ✅ No `ALLOW FILTERING` needed
- ✅ No duplicate or missing rows between pages
- ✅ Works with any table size (even 100M+ rows)
- ⚠️ Cursors are stateful and expire after 30 minutes
- ⚠️ No random page access (forward-only pagination)
- ⚠️ Server must store cursor state in memory

---

## ADR-009: Read-Only by Default

**Status:** Accepted

**Context:**
In Cassandra, `INSERT` and `UPDATE` are identical (upserts). A typo in a primary key creates duplicate "ghost rows" instead of updating the intended row. This is a significant safety risk.

**Decision:**
Kassie is read-only by default. No write operations are exposed through the API. A future "Safe Mode Editor" may allow editing non-key columns with dry-run preview.

**Consequences:**
- ✅ Eliminates the ghost row risk entirely
- ✅ Safe to use on production databases
- ✅ Simpler server implementation (no mutation logic)
- ✅ Clear security boundary — Kassie can't damage data
- ⚠️ Users must use cqlsh or other tools for writes
- ⚠️ Limits Kassie's utility for some workflows
