# Authentication

Kassie uses JWT (JSON Web Tokens) to authenticate clients with the server. This page covers the auth system design, token lifecycle, and security considerations.

## Overview

Authentication serves two purposes:

1. **Session binding** — Associates a client with a specific database profile and connection
2. **Request authorization** — Validates every API request belongs to an active session

## JWT Token Structure

### Access Token

Used for API authorization on every request.

**Claims:**

| Claim | Type | Description |
|-------|------|-------------|
| `session_id` | string | Unique session identifier |
| `profile` | string | Database profile name |
| `iat` | int64 | Issued at (Unix timestamp) |
| `exp` | int64 | Expires at (Unix timestamp) |

**Lifetime:** 1 hour

### Refresh Token

Used to obtain a new access token without re-login.

**Claims:**

| Claim | Type | Description |
|-------|------|-------------|
| `session_id` | string | Unique session identifier |
| `iat` | int64 | Issued at (Unix timestamp) |
| `exp` | int64 | Expires at (Unix timestamp) |

**Lifetime:** 24 hours

**Signing algorithm:** HMAC-SHA256

## Authentication Flow

```
┌──────────┐                          ┌──────────┐                    ┌─────────┐
│  Client  │                          │  Server  │                    │   DB    │
└────┬─────┘                          └────┬─────┘                    └────┬────┘
     │                                     │                               │
     │  1. Login(profile: "local")         │                               │
     │────────────────────────────────────>│                               │
     │                                     │  2. Validate profile exists   │
     │                                     │  3. Connect to database       │
     │                                     │──────────────────────────────>│
     │                                     │  4. Connection OK             │
     │                                     │<──────────────────────────────│
     │                                     │  5. Create session            │
     │                                     │  6. Generate JWT pair         │
     │  7. {access_token, refresh_token}   │                               │
     │<────────────────────────────────────│                               │
     │                                     │                               │
     │  8. API Request + Authorization     │                               │
     │────────────────────────────────────>│                               │
     │                                     │  9. Validate token            │
     │                                     │  10. Execute query            │
     │                                     │──────────────────────────────>│
     │  11. Response                       │                               │
     │<────────────────────────────────────│<──────────────────────────────│
```

### Step-by-Step

1. Client calls `Login` with a profile name
2. Server validates the profile exists in configuration
3. Server connects to the database using profile credentials
4. On successful connection, server creates a session in the state store
5. Server generates an access token and refresh token
6. Client stores tokens and includes the access token in all subsequent requests
7. Server interceptor validates the token on every request
8. On `401 Unauthorized`, client attempts a token refresh
9. If refresh fails, client redirects to login

## Token Refresh

When the access token expires, clients automatically refresh:

```
Client                              Server
  │                                    │
  │  API Request (expired token)       │
  │───────────────────────────────────>│
  │  401 Unauthorized                  │
  │<───────────────────────────────────│
  │                                    │
  │  Refresh(refresh_token)            │
  │───────────────────────────────────>│
  │  {new_access_token, expires_at}    │
  │<───────────────────────────────────│
  │                                    │
  │  Retry original request            │
  │───────────────────────────────────>│
  │  200 OK                            │
  │<───────────────────────────────────│
```

If the refresh token is also expired, the client must re-login.

## Token Storage

### TUI Client

- **In-memory only** — tokens are never persisted to disk
- Embedded server mode means tokens are local to the process
- Session ends when the TUI exits

### Web Client

- **localStorage** — tokens stored in browser localStorage
- Acceptable security for a localhost tool
- Auth store rehydrates from localStorage on page load
- Logout clears all stored tokens

## Server-Side Validation

The gRPC auth interceptor runs on every request (except `Login` and `GetProfiles`):

1. Extract `Authorization: Bearer <token>` header
2. Parse and validate JWT signature
3. Check token expiration
4. Look up session in state store
5. Attach session context to the request
6. If any step fails, return appropriate error code

## JWT Secret Management

| Mode | Secret Source | Risk Level |
|------|-------------|------------|
| Embedded (TUI/Web) | Auto-generated at startup | Low — localhost only |
| Standalone Server | `KASSIE_JWT_SECRET` env var | Higher — network-exposed |

::: warning
For standalone server mode, always set a strong `KASSIE_JWT_SECRET` environment variable:

```bash
export KASSIE_JWT_SECRET=$(openssl rand -hex 32)
kassie server
```
:::

## Security Considerations

- **Database credentials never exposed to clients** — clients only send profile names, the server reads credentials from config
- **No sensitive data in URLs** — tokens passed via headers, not query parameters
- **XSS protection** — React's default escaping prevents token theft via XSS
- **HMAC-SHA256 signing** — prevents token tampering
- **Session state validation** — even valid tokens are rejected if the session was explicitly logged out

## Error Codes

| Code | Meaning | Client Action |
|------|---------|--------------|
| `AUTH_REQUIRED` | No token provided | Redirect to login |
| `AUTH_INVALID` | Token invalid or expired | Attempt refresh |
| `AUTH_FORBIDDEN` | Insufficient permissions | Show error message |
