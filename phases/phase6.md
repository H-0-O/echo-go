# Phase 6 — Null Broadcaster & Custom Connectors

**Goal:** Test/dev ergonomics and extensibility.

**Priority:** Sixth — polish for CI and custom backends; not blocking core Reverb/Pusher parity.

**Depends on:** Phases 1–5 (full channel surface and encrypted null type from Phase 5).

---

## Background

Laravel Echo supports a `null` broadcaster for tests and local dev: channel methods return chainable no-op objects so application code can run without a WebSocket server. JS also allows `broadcaster: function` — a factory that returns a custom connector.

echo-go has a partial `NullConnector` and `NullChannel`, but:

- `PrivateChannel` reuses the public `NullChannel` (no separate `NullPrivateChannel`)
- `PresenceChannel` returns a fresh `NullPresenceChannel` without registry caching
- Phase 3+ methods (`Subscribed`, `ListenToAll`, etc.) may be missing from null types
- `EncryptedPrivateChannel` null type comes from Phase 5 — this phase completes parity
- There is no way to inject a custom `Connector` without forking the package

---

## Current state

| Area | File | Gap |
|------|------|-----|
| `NullChannel` | `null_channel.go` | Partial — missing Phase 3 API stubs |
| `NullPrivateChannel` | — | Does not exist; `PrivateChannel` aliases public null |
| `NullPresenceChannel` | `null_channel.go` | Exists but not registry-cached |
| `NullEncryptedPrivateChannel` | Phase 5 | May be incomplete |
| Registry key scheme | `null_connector.go` | Does not match Phase 2 prefixed keys |
| Custom connector | `echo.go` `New` | Only string broadcaster switch |
| Exported `Connector` | — | Not exported; may be needed for injection |

---

## Tasks

### 6.1 Complete null channel types

**File:** `internal/channel/null_channel.go` (split into `null_*.go` only if file grows unwieldy)

Mirror JS no-op chaining — every method returns `self` (or typed self for presence):

| Type | Embeds | Notes |
|------|--------|-------|
| `NullChannel` | — | Public channel no-op |
| `NullPrivateChannel` | `NullChannel` | Distinct type for clarity/tests |
| `NullPresenceChannel` | `NullChannel` | Adds `Here`, `Joining`, `Leaving` no-ops |
| `NullEncryptedPrivateChannel` | `NullChannel` | From Phase 5 — verify all methods |

Implement **all** methods from extended `Channel` interface (Phase 3):

```go
Subscribed(cb func()) Channel
Error(cb func(error)) Channel
Notification(cb func(data any)) Channel
StopListeningForNotification(cb ...func(data any)) Channel
ListenToAll(cb func(event string, data any)) Channel
StopListeningToAll(cb ...func(event string, data any)) Channel
Unsubscribe()
```

`StopListening` / `StopListeningForWhisper` with optional callback variadic (Phase 3 signature).

**Acceptance criteria:**

- [ ] Each null type satisfies its interface at compile time (`var _ Channel = (*NullChannel)(nil)`)
- [ ] Method chaining works: `n.Listen("x", fn).ListenForWhisper("y", fn).Whisper("z", nil)`
- [ ] No panics, no goroutines, no side effects

---

### 6.2 Complete `NullConnector`

**File:** `internal/connector/null_connector.go`

Align with Phase 2 registry and Phase 5 encrypted channel:

| Method | Returns | Registry key |
|--------|---------|--------------|
| `Channel(name)` | `*NullChannel` | `name` |
| `PrivateChannel(name)` | `*NullPrivateChannel` | `private-{name}` |
| `PresenceChannel(name)` | `*NullPresenceChannel` | `presence-{name}` |
| `EncryptedPrivateChannel(name)` | `*NullEncryptedPrivateChannel` | `private-encrypted-{name}` |

Connection behavior (unchanged):

- `Connect` / `Disconnect` → `nil` error
- `SocketID()` → `"null-socket-id"` (stable test fixture)
- `On` → no-op

Add Phase 1–2 connector methods:

```go
ConnectionStatus() ConnectionStatus  // always "connected" or "disconnected" — pick one and document
OnConnectionChange(cb func(ConnectionStatus)) func()
Leave(channel string)
LeaveChannel(name string)
LeaveAllChannels()
Signin()  // Phase 4 no-op
```

**Decision:** `ConnectionStatus` for null connector — recommend `connected` after `Connect()` and `disconnected` otherwise, so tests asserting status behave predictably.

**Acceptance criteria:**

- [ ] Registry keys match `PusherConnector` scheme
- [ ] `Private("a")` and `Channel("a")` return different null instances
- [ ] `Leave` / `LeaveAllChannels` clear registry
- [ ] `New(Config{Broadcaster: "null"})` works without `Key`/`Host`

---

### 6.3 Custom connector injection

**Goal:** Go equivalent of Laravel Echo `broadcaster: function`.

**File:** `echo.go`

Extend `Config`:

```go
type Config struct {
    // ... existing fields ...

    // Connector, when non-nil, is used instead of Broadcaster string resolution.
    // Broadcaster is ignored when Connector is set.
    Connector connector.Connector
}
```

`New` logic:

```go
func New(config Config) (*Echo, error) {
    if config.Connector != nil {
        return &Echo{connector: config.Connector, config: config}, nil
    }
    // existing switch on config.Broadcaster ...
}
```

**Export decision:**

| Option | Pros | Cons |
|--------|------|------|
| Keep `Connector` in `internal/connector` only | Smaller public API | Custom connectors must live inside module or use awkward workarounds |
| Export `Connector` type alias in root `echo.go` | Matches extensibility goal | Commits to interface stability |

**Recommendation:** Export from root package:

```go
// Connector is the broadcasting backend interface.
// Use Config.Connector to inject a custom implementation.
type Connector = connector.Connector
```

Only export the interface — keep `PusherConnector`, `NullConnector` constructors internal unless a concrete use case demands export.

**Acceptance criteria:**

- [ ] `New(Config{Connector: myConn})` skips broadcaster validation
- [ ] Custom connector receives all `Echo` delegations (`Channel`, `Connect`, `Signin`, leave methods, etc.)
- [ ] Godoc example shows minimal stub connector for unit tests

---

### 6.4 Example custom connector (in tests or `examples/`)

Minimal stub for documentation — **not** a new package:

```go
type stubConnector struct {
    connector.Connector // embed NullConnector or implement minimal subset
}
```

Use in `echo_test.go` or `examples/custom_connector/main.go` (Phase 7 may promote to examples/).

**Acceptance criteria:**

- [ ] Compiles and demonstrates injection pattern
- [ ] Does not add production dependency

---

### 6.5 `New` with `Broadcaster: "null"` vs injected null

Clarify in godoc:

- String `"null"` → built-in `NullConnector` (zero config)
- `Config.Connector = connector.NewNullConnector()` — only needed if caller wraps null with metrics/logging

Avoid exporting `NewNullConnector` unless requested; custom wrappers can compose internal types via exported `Connector` interface + their own struct.

---

## Files touched

| File | Changes |
|------|---------|
| `echo.go` | `Config.Connector`, optional `Connector` type export, `New` branch |
| `internal/connector/null_connector.go` | Full registry, leave/status/signin stubs |
| `internal/channel/null_channel.go` | All channel methods, split types |
| `internal/channel/null_encrypted_private_channel.go` | Verify complete (Phase 5) |
| `echo_test.go` or `examples/custom_connector/main.go` | Injection demo |

---

## Tests

| Test | Purpose |
|------|---------|
| `TestNullChannelChaining` | All methods return receiver |
| `TestNullRegistryKeys` | Same key scheme as pusher connector |
| `TestNullLeave` | Leave removes cached instances |
| `TestCustomConnectorInjection` | `Config.Connector` used by `Echo` |
| `TestNullBroadcasterNoKey` | `New` succeeds with only `Broadcaster: "null"` |

---

## Definition of done

- [ ] All four null channel types implement full interfaces
- [ ] `NullConnector` registry and leave behavior match `PusherConnector`
- [ ] Custom connector injectable via `Config.Connector`
- [ ] `Connector` interface exported from root package (if following recommendation above)
- [ ] `go test ./...` passes

---

## Out of scope

- Socket.IO or non-Pusher custom connectors (implementable by user via `Connector` interface, not shipped)
- Exporting `NewPusherConnector` / `NewNullConnector` constructors
- Mock HTTP auth in null connector

---

## References

- Laravel Echo: `null-connector.ts`, `connector.ts` (`broadcaster` factory)
- ROADMAP.md — intentional Go differences (explicit connect, no DOM)
