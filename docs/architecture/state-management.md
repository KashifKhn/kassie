# State Management

Kassie manages state across three layers: server-side state, TUI client state, and Web client state. Each layer uses different strategies optimized for its context.

## Server-Side State

All server state is held in-memory. There is no external database or cache for state storage.

### Session Store

Maps `session_id` to session data:

| Field | Type | Description |
|-------|------|-------------|
| `session_id` | string | UUID, unique per login |
| `profile` | ProfileInfo | Active database profile |
| `connection` | *gocql.Session | Database connection handle |
| `created_at` | time.Time | Session creation timestamp |

**Lifecycle:**
1. Created on successful `Login`
2. Used for every authenticated request
3. Destroyed on `Logout` or token expiry
4. Connection pool returned on cleanup

### Cursor Store

Maps `cursor_id` to Scylla paging state for pagination:

| Field | Type | Description |
|-------|------|-------------|
| `cursor_id` | string | UUID, unique per query |
| `paging_state` | []byte | Scylla paging state token |
| `keyspace` | string | Query keyspace |
| `table` | string | Query table |
| `where_clause` | string | Active filter (if any) |
| `last_accessed` | time.Time | For expiration tracking |

**Lifecycle:**
1. Created on initial `QueryRows` or `FilterRows`
2. Updated on each `GetNextPage` call
3. Expires after 30 minutes of inactivity
4. Manually cleared on new query to same table

### Connection Pool

Per-profile database connection management:

```
Profile "local"  ──> Connection Pool (5 connections)
Profile "staging" ──> Connection Pool (5 connections)
Profile "prod"   ──> Connection Pool (5 connections)
```

- Pools are created lazily on first login
- Shared across sessions using the same profile
- Health-checked and auto-reconnected
- Closed when all sessions for a profile end

---

## TUI State (Bubbletea)

The TUI uses Bubbletea's Elm Architecture: Model → Update → View.

### Global State

```go
type AppState struct {
    CurrentView    View          // connection, explorer, help
    Profile        *ProfileInfo  // active profile after login
    SessionToken   string        // JWT access token
    RefreshToken   string        // JWT refresh token
    Error          string        // current error message
}
```

### Explorer State

```go
type ExplorerState struct {
    Keyspaces      []Keyspace    // loaded keyspaces
    Tables         []Table       // tables for selected keyspace
    SelectedKS     string        // active keyspace name
    SelectedTable  string        // active table name
    Schema         *TableSchema  // column definitions
    Rows           []Row         // current page of data
    CursorID       string        // pagination cursor
    HasMore        bool          // more pages available
    Filter         string        // active WHERE clause
    FocusedPane    Pane          // sidebar, grid, or inspector
}
```

### Component State

Each component manages its own local state:

| Component | Local State |
|-----------|-------------|
| Sidebar | Scroll position, expanded keyspaces, search text |
| DataGrid | Column widths, horizontal scroll offset, selected row index |
| Inspector | Scroll position, expanded JSON nodes |
| FilterBar | Input text, cursor position, validation error |
| StatusBar | Derived from global + explorer state (no own state) |

### State Flow

```
User Input ──> Bubbletea Update() ──> New Model ──> View() ──> Terminal
                    │
                    ├── Local state change (navigation, UI toggle)
                    │
                    └── gRPC call ──> Server ──> Response ──> Update Model
```

---

## Web State

The Web client uses a three-layer state model following React best practices.

### Server State (TanStack Query)

All data fetched from the API lives in the TanStack Query cache:

| Query Key | Data | Stale Time |
|-----------|------|-----------|
| `['profiles']` | Profile list | 5 minutes |
| `['keyspaces']` | Keyspace list | 5 minutes |
| `['tables', keyspace]` | Tables for a keyspace | 5 minutes |
| `['schema', keyspace, table]` | Table schema/columns | 5 minutes |
| `['rows', keyspace, table]` | Initial row query | On navigation |
| `['rows', keyspace, table, 'filter', where]` | Filtered rows | On navigation |
| `['rows', 'page', cursorId]` | Next page of data | On navigation |

**Benefits:**
- Automatic background refetching
- Cache invalidation
- Request deduplication
- Loading and error states built-in

### UI State (Zustand)

Ephemeral client-side state stored in lightweight Zustand stores:

**Auth Store:**

| Field | Type | Persisted |
|-------|------|-----------|
| `accessToken` | string | localStorage |
| `refreshToken` | string | localStorage |
| `expiresAt` | number | localStorage |
| `profile` | ProfileInfo | localStorage |
| `isAuthenticated` | boolean | Derived |

**UI Store:**

| Field | Type | Persisted |
|-------|------|-----------|
| `sidebarCollapsed` | boolean | localStorage |
| `inspectorOpen` | boolean | localStorage |
| `selectedRowIndex` | number | No |
| `theme` | 'light' \| 'dark' \| 'system' | localStorage |

### URL State (React Router)

Navigation state encoded in URL search parameters for shareability:

```
/?keyspace=app_data&table=users&filter=id%3D'abc'&page=2
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `keyspace` | string | Selected keyspace |
| `table` | string | Selected table |
| `filter` | string | WHERE clause (URL-encoded) |
| `page` | number | Current page number |

**Benefits:**
- Shareable links
- Browser back/forward navigation
- Bookmark support
- State survives page refresh

### Web State Flow

```
User Action
    │
    ├── UI change ──> Zustand store ──> React re-render
    │
    ├── Navigation ──> URL params ──> TanStack Query fetch ──> Re-render
    │
    └── Data fetch ──> TanStack Query ──> API call ──> Cache update ──> Re-render
```

---

## State Persistence Summary

| State | TUI | Web |
|-------|-----|-----|
| JWT tokens | In-memory | localStorage |
| UI preferences | Not persisted | localStorage |
| Navigation state | In-memory | URL params |
| Query cache | Not applicable | TanStack Query (memory) |
| Server sessions | In-memory (server) | In-memory (server) |
| Pagination cursors | In-memory (server) | In-memory (server) |
