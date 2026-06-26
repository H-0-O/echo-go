# Phase 2 — Channel Lifecycle

**Goal:** Subscribe/unsubscribe and registry behavior match Laravel Echo.

**Priority:** Second — fixes a real bug (cache key collision) and unlocks leave/join ergonomics.

**Depends on:** Phase 1 (stable connector interface). pusher-go `Unsubscribe` / channel `Unsubscribe()`.

---

## Background

Laravel Echo’s `PusherConnector` stores channels in a registry keyed by the **full prefixed name** (`private-orders`, `presence-chat`, `private-encrypted-secrets`). echo-go currently keys private and presence channels by the **unprefixed** name passed to `Private()` / `Presence()`, so:

```go
echo.Private("orders")   // caches as "orders"
echo.Channel("orders")   // returns same cached entry — wrong
```

JS `leave(channel)` unsubscribes all variants (public, private, encrypted, presence) for a logical channel name. echo-go has no leave APIs and channels cannot be unsubscribed publicly.

---

## Current state

| Issue | Location | Detail |
|-------|----------|--------|
| Cache key bug | `pusher_connector.go` `PrivateChannel`, `PresenceChannel` | Uses `name` not `private-{name}` / `presence-{name}` |
| No leave API | `echo.go`, `connector.go` | Missing `Leave`, `LeaveChannel`, `LeaveAllChannels` |
| No `Join` alias | `echo.go` | JS `join(name)` === `presence(name)` |
| No shorthand listen | `echo.go` | JS `listen(channel, event, cb)` |
| No `Unsubscribe` | `channel.go`, `pusher_channel.go` | Internal only in JS via connector leave |

---

## Tasks

### 2.1 Fix channel cache keys

**File:** `internal/connector/pusher_connector.go`

Use prefixed keys consistently:

| Method | Registry key |
|--------|--------------|
| `Channel("orders")` | `orders` |
| `PrivateChannel("orders")` | `private-orders` |
| `PresenceChannel("chat")` | `presence-chat` |
| `EncryptedPrivateChannel("x")` (Phase 5) | `private-encrypted-x` |

Apply prefix inside the connector before map lookup/store (channel types may also prefix — ensure single prefix, not double).

**Acceptance criteria:**

- [ ] `Private("a")` and `Channel("a")` return distinct channel instances
- [ ] Second call to `Private("a")` returns same instance
- [ ] `NullConnector` uses same key scheme for consistency

---

### 2.2 `Leave(channel string)`

**Files:** `echo.go`, `internal/connector/connector.go`, `pusher_connector.go`, `null_connector.go`

Laravel Echo leaves all subscription types for a logical name:

```go
func (e *Echo) Leave(channel string)
```

Implementation leaves, if subscribed:

- `channel`
- `private-{channel}`
- `private-encrypted-{channel}` (no-op until Phase 5)
- `presence-{channel}`

For each: call channel `Unsubscribe()`, remove from registry.

**Acceptance criteria:**

- [ ] Idempotent — leaving a non-subscribed name is safe
- [ ] Does not disconnect the WebSocket
- [ ] Matches JS `leave()` behavior for all four prefixes

---

### 2.3 `LeaveChannel(name string)`

Leaves a **single** channel by exact name (already prefixed if private/presence).

```go
func (e *Echo) LeaveChannel(name string)
```

**Acceptance criteria:**

- [ ] `LeaveChannel("private-orders")` only removes that entry
- [ ] Unknown name is no-op

---

### 2.4 `LeaveAllChannels()`

```go
func (e *Echo) LeaveAllChannels()
```

Iterate connector registry, unsubscribe each, clear map.

**Acceptance criteria:**

- [ ] After call, registry empty
- [ ] Connection remains open
- [ ] Re-subscribing same names creates fresh channel instances

---

### 2.5 `Join(name)` alias

```go
func (e *Echo) Join(name string) PresenceChannel {
    return e.Presence(name)
}
```

**Acceptance criteria:**

- [ ] Documented as alias; identical behavior to `Presence`

---

### 2.6 `Listen(channel, event, cb)` shorthand

```go
func (e *Echo) Listen(channelName, event string, callback func(data any)) Channel {
    return e.Channel(channelName).Listen(event, callback)
}
```

JS accepts channel name string; resolves public channel only (not private/presence).

**Acceptance criteria:**

- [ ] Returns `Channel` for chaining
- [ ] Uses `Channel()`, not `Private()` / `Presence()`

---

### 2.7 Channel `Unsubscribe()`

**Files:** `internal/channel/channel.go`, `pusher_channel.go`, `null_channel.go`

Add to `Channel` interface:

```go
Unsubscribe()
```

**Pusher implementation:**

1. `c.Channel.Unsubscribe()` via pusher-go
2. Clear local `callbacks` sync.Map
3. Connector removes from registry when invoked via `Leave*` (connector calls `Unsubscribe` then deletes map entry)

**Presence / private:** inherit from `PusherChannel` or override if pusher-go requires channel-name-specific unsubscribe.

**Acceptance criteria:**

- [ ] Public, private, presence channels unsubscribe correctly
- [ ] `NullChannel.Unsubscribe()` is no-op
- [ ] Listeners after unsubscribe do not fire

---

### 2.8 Extend `Connector` interface

**File:** `internal/connector/connector.go`

```go
Leave(channel string)
LeaveChannel(name string)
LeaveAllChannels()
```

`Echo` methods delegate to connector.

---

## Files touched

| File | Changes |
|------|---------|
| `echo.go` | `Leave`, `LeaveChannel`, `LeaveAllChannels`, `Join`, `Listen` |
| `internal/connector/connector.go` | Leave methods on interface |
| `internal/connector/pusher_connector.go` | Registry fix, leave impl, `Unsubscribe` delegation |
| `internal/connector/null_connector.go` | Leave stubs, registry fix |
| `internal/channel/channel.go` | `Unsubscribe()` on interface |
| `internal/channel/pusher_channel.go` | `Unsubscribe()` impl |
| `internal/channel/null_channel.go` | `Unsubscribe()` no-op |

---

## Tests

| Test | Purpose |
|------|---------|
| `TestChannelRegistryKeys` | Public vs private vs presence distinct |
| `TestLeaveAllVariants` | `Leave("x")` removes public, private, presence for `x` |
| `TestLeaveChannelExact` | Only one prefixed name removed |
| `TestLeaveAllChannels` | Registry cleared, connection alive |
| `TestUnsubscribeStopsEvents` | No callbacks after unsubscribe |

Prefer table-driven tests; mock pusher-go channel if needed.

---

## Definition of done

- [ ] Cache key bug fixed and covered by test
- [ ] All leave/join/listen APIs on `Echo`
- [ ] `Unsubscribe()` on `Channel` interface
- [ ] `go test ./...` passes
- [ ] No regression in Phase 1 connection APIs

---

## Out of scope

- `ListenToAll`, `Subscribed`, `Error` (Phase 3)
- `EncryptedPrivate` registry entries (Phase 5 — but key scheme documented here)
