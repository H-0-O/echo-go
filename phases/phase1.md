# Phase 1 — Config & Connection Parity

**Goal:** Match Laravel Echo defaults and connection observability without changing call sites much.

**Priority:** First — all later phases assume stable config merging and connection status APIs.

**Depends on:** pusher-go connection state events (`Connection.State()`, `Connection.Bind`).

---

## Background

Laravel Echo’s `EchoOptions` (see `packages/laravel-echo/src/echo.ts` on the `2.x` branch) centralizes auth, namespace, and cluster defaults in `setOptions()`. echo-go currently exposes a flat `Config` with no defaults, no bearer/CSRF header merging, and no typed connection status.

Connection observability in JS uses `connectionStatus()` and `onConnectionChange(cb)` returning an unsubscribe function. echo-go only exposes lower-level `On(event, callback)` mapped to raw pusher-go connection events.

---

## Current state

| Area | File | Gap |
|------|------|-----|
| Flat config, no defaults | `echo.go` `Config` | Missing `Auth`, `UserAuthentication`, `BearerToken`, `CSRFToken`, `Cluster`, `AutoConnect` |
| Unknown broadcaster → null | `echo.go` `New` | Should return error |
| No `ably` case | `echo.go` `New` | JS routes `ably` to Pusher connector with empty cluster |
| Raw connection events only | `echo.go` `On` | No `ConnectionStatus()` / `OnConnectionChange` |
| Hardcoded cluster `mt1` | `pusher_connector.go` | Should use `Config.Cluster` for real Pusher |

---

## Tasks

### 1.1 Expand `Config` to mirror `EchoOptions` (non-browser subset)

**File:** `echo.go`

Add types and fields:

```go
type AuthConfig struct {
    Endpoint string            // default: "/broadcasting/auth"
    Headers  map[string]string
}

type UserAuthConfig struct {
    Endpoint string            // default: "/broadcasting/user-auth"
    Headers  map[string]string
}

type Config struct {
    Broadcaster         string
    Key                 string
    Cluster             string   // real Pusher only; empty for Reverb/Ably
    Host                string
    Port                int
    TLS                 bool
    Namespace           string   // default: "App.Events"; empty disables namespace
    Auth                AuthConfig
    UserAuthentication  UserAuthConfig
    BearerToken         string
    CSRFToken           string
    AutoConnect         bool     // default: true (match JS); see 1.5
}
```

**Acceptance criteria:**

- [ ] `New` applies defaults when fields are empty:
  - `Auth.Endpoint` → `/broadcasting/auth`
  - `UserAuthentication.Endpoint` → `/broadcasting/user-auth`
  - `Namespace` → `App.Events`
- [ ] `BearerToken` merged as `Authorization: Bearer {token}` into auth headers (both channel and user auth if applicable)
- [ ] `CSRFToken` merged as `X-CSRF-TOKEN` into auth headers
- [ ] `Auth.Headers` and `UserAuthentication.Headers` merged on top (caller overrides win)
- [ ] Empty `Namespace` disables prefixing (JS `namespace: false` equivalent)

**Laravel Echo reference:** `connector.ts` `defaultOptions`, `pusher-connector.ts` `setOptions`.

**Migration:** Keep deprecated top-level `AuthEndpoint` / `AuthHeaders` as aliases or remove in same phase with a breaking-change note in CHANGELOG (Phase 7).

---

### 1.2 `ConnectionStatus` type

**File:** `echo.go`

```go
type ConnectionStatus string

const (
    StatusConnected     ConnectionStatus = "connected"
    StatusDisconnected  ConnectionStatus = "disconnected"
    StatusConnecting    ConnectionStatus = "connecting"
    StatusReconnecting  ConnectionStatus = "reconnecting"
    StatusFailed        ConnectionStatus = "failed"
)
```

Map from pusher-go `Connection.State()` / `ConnectionState()` string values.

| pusher-go state | echo-go status |
|-----------------|----------------|
| `connected` | `connected` |
| `disconnected` | `disconnected` |
| `connecting` | `connecting` |
| `unavailable` / reconnect paths | `reconnecting` |
| `failed` | `failed` |

**Acceptance criteria:**

- [ ] `ConnectionStatus()` on `Echo` returns current mapped status
- [ ] `PusherConnector` implements mapping in one place (not duplicated in `echo.go`)

---

### 1.3 `OnConnectionChange(cb) func()`

**Files:** `echo.go`, `internal/connector/connector.go`, `internal/connector/pusher_connector.go`

```go
func (e *Echo) OnConnectionChange(cb func(ConnectionStatus)) func()
```

- Bind to pusher-go connection `state_change` (or equivalent) events
- Invoke `cb` with mapped `ConnectionStatus` on each transition
- Return unsubscribe function that removes the listener
- `NullConnector`: no-op subscribe, no-op unsubscribe

**Acceptance criteria:**

- [ ] Callback fires on connect, disconnect, reconnect, and failure
- [ ] Unsubscribe stops further callbacks
- [ ] Does not break existing `On(event, callback)` API

---

### 1.4 Broadcaster handling

**File:** `echo.go` `New`

| Broadcaster | Behavior |
|-------------|----------|
| `reverb`, `pusher` | `PusherConnector` with `Cluster` from config |
| `ably` | Same as `pusher` with empty cluster (JS behavior) |
| `null` | `NullConnector` |
| anything else | `return nil, fmt.Errorf("unknown broadcaster %q", ...)` |

**Acceptance criteria:**

- [ ] Invalid broadcaster returns error from `New` (no silent null fallback)
- [ ] `ably` works with Reverb-style `Host`/`Port` config

---

### 1.5 Optional `AutoConnect`

**Files:** `echo.go`, `internal/connector/connector.go`

When `AutoConnect` is true (default), `New` calls `Connect()` before returning.

**Decision to document:**

- JS: constructor always connects
- Go idiom: explicit `Connect()` is valid; default `AutoConnect: true` preserves JS parity for drop-in ports

**Acceptance criteria:**

- [ ] `AutoConnect: false` → caller must call `Connect()`
- [ ] Connect error from auto-connect propagates from `New`

---

### 1.6 Pass expanded config to `PusherConnector`

**Files:** `internal/connector/pusher_connector.go`, `internal/connector/connector.go`

Extend `PusherConfig` with `Cluster`, merged auth headers, and user-auth endpoint/headers (wired in Phase 4 for `Signin`, but struct fields belong here).

```go
pusher.NewPusher(config.Key, pusher.Options{
    Options: types.Options{
        Cluster:  config.Cluster, // or "mt1" only when Host empty and Cluster empty
        WsHost:   config.Host,
        // ...
        ChannelAuthorization: types.ChannelAuthorizationConfig{ ... },
        UserAuthentication:   types.UserAuthenticationConfig{ ... }, // struct only; Signin in Phase 4
    },
})
```

**Acceptance criteria:**

- [ ] Reverb: `WsHost` set, cluster ignored
- [ ] Real Pusher: `Cluster` used when `Host` empty

---

## Files touched

| File | Changes |
|------|---------|
| `echo.go` | `Config`, `ConnectionStatus`, `ConnectionStatus()`, `OnConnectionChange`, `New` defaults & broadcaster logic |
| `internal/connector/connector.go` | Extend `Connector` with status methods |
| `internal/connector/pusher_connector.go` | Status mapping, config passthrough |
| `internal/connector/null_connector.go` | Stub status methods |

---

## Tests

| Test | Purpose |
|------|---------|
| `TestConfigDefaults` | Empty config gets Laravel defaults |
| `TestBearerCSRFMerge` | Token fields become correct headers |
| `TestUnknownBroadcaster` | `New` returns error |
| `TestConnectionStatusMapping` | Table-driven state → status |
| `TestOnConnectionChangeUnsubscribe` | Unsubscribe stops callbacks |

Use pusher-go test helpers or mock connection if integration is heavy.

---

## Definition of done

- [ ] All tasks above checked off
- [ ] `go test ./...` passes
- [ ] README snippet updated only if Phase 7 is blocked — otherwise defer full README to Phase 7
- [ ] No breaking changes to `Channel` / `Listen` APIs

---

## Out of scope (later phases)

- `Signin()` / user auth wiring (Phase 4)
- Channel leave/registry (Phase 2)
- Encrypted channels (Phase 5)
