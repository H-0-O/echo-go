# Phase 3 — Channel API Completeness

**Goal:** Full `Channel` interface parity for Pusher/Reverb channels.

**Priority:** Third — builds on Phase 2 unsubscribe/listen infrastructure.

**Depends on:** Phase 2 (`Unsubscribe`, registry). pusher-go `Bind`, `Unbind`, `BindGlobal` (per channel or client).

---

## Background

Laravel Echo’s Pusher channel wrapper adds subscription lifecycle hooks, Laravel notification event binding, global event listening, and correct whisper event naming (`.client-{event}` through the formatter so the leading `.` skips namespace).

Presence channels normalize member payloads: `here` receives `Object.values(members)`, `joining`/`leaving` receive `member.info`.

echo-go gaps are documented in ROADMAP.md gap analysis.

---

## Current state

| API | File | Gap |
|-----|------|-----|
| `ListenForWhisper` | `pusher_channel.go` | Uses `client-{event}` not `.client-{event}` → namespace incorrectly applied |
| `StopListening` | `pusher_channel.go` | No optional callback — always removes all handlers for event |
| `Subscribed`, `Error` | — | Not implemented |
| `Notification` | — | Not implemented |
| `ListenToAll` | — | Not implemented |
| Presence payloads | `pusher_presence_channel.go` | Raw pusher payloads, not normalized |

---

## Tasks

### 3.1 Fix `ListenForWhisper` prefix

**File:** `internal/channel/pusher_channel.go`

Change:

```go
// Before
c.Listen("client-"+event, callback)

// After
c.Listen(".client-"+event, callback)
```

`EventFormatter.Format` strips leading `.` and skips namespace (already implemented in `event_formatter.go`).

**Whisper trigger:** verify `Whisper` still sends `client-{event}` (no leading dot on wire).

**Acceptance criteria:**

- [ ] With namespace `App.Events`, listen whisper `typing` binds `client-typing` not `App.Events.client-typing`
- [ ] Without namespace, binds `client-typing`

---

### 3.2 `StopListening(event, cb?)` optional callback

**Files:** `internal/channel/channel.go`, `pusher_channel.go`, `null_channel.go`

Extend signature:

```go
StopListening(event string, callback ...func(data interface{})) Channel
```

- **No callback:** remove all handlers for event (current behavior)
- **With callback:** unbind only that specific wrapped callback if pusher-go `Unbind(event, cb)` supports it

Track mapping: original callback → wrapped callback in `callbacks` sync.Map (may need `map[event][]entry` if multiple listeners per event).

**Acceptance criteria:**

- [ ] Multiple `Listen` on same event; `StopListening(event, cb2)` leaves `cb1` active
- [ ] `StopListening(event)` removes all
- [ ] Null channel: no-op, returns self

---

### 3.3 `Subscribed(cb)`

**File:** `internal/channel/pusher_channel.go`

```go
func (c *PusherChannel) Subscribed(callback func()) Channel
```

Bind `pusher:subscription_succeeded` once per registration (allow multiple callbacks like JS).

**Acceptance criteria:**

- [ ] Fires after successful subscription
- [ ] Callback receives no payload (or ignore payload)
- [ ] Chainable

---

### 3.4 `Error(cb)`

```go
func (c *PusherChannel) Error(callback func(error)) Channel
```

Bind `pusher:subscription_error`; map payload to `error` (type assert or fmt.Errorf with code/message).

**Acceptance criteria:**

- [ ] Invoked on auth failure / subscription error
- [ ] Chainable

---

### 3.5 `Notification(cb)` / `StopListeningForNotification(cb)`

Laravel broadcast notifications use a fixed event name:

```
.Illuminate\Notifications\Events\BroadcastNotificationCreated
```

Leading `.` skips namespace (same as whisper fix).

```go
func (c *PusherChannel) Notification(callback func(data interface{})) Channel
func (c *PusherChannel) StopListeningForNotification(callback ...func(data interface{})) Channel
```

`StopListeningForNotification` delegates to `StopListening` with the notification event constant.

**Acceptance criteria:**

- [ ] Binds correct event regardless of `Config.Namespace`
- [ ] Stop removes notification listener only when callback specified

---

### 3.6 `ListenToAll(cb)` / `StopListeningToAll(cb?)`

```go
func (c *PusherChannel) ListenToAll(callback func(event string, data interface{})) Channel
func (c *PusherChannel) StopListeningToAll(callback ...func(event string, data interface{})) Channel
```

Wire pusher-go channel-level global bind if available; otherwise bind internal dispatcher:

- Skip events starting with `pusher:`
- Strip namespace from event name before passing to callback (mirror JS `listenToAll`)

**Acceptance criteria:**

- [ ] `App.Events.OrderShipped` reported as `OrderShipped` when namespace is `App.Events`
- [ ] `pusher:subscription_succeeded` not forwarded
- [ ] Optional stop by callback reference

**Research:** confirm pusher-go `Channel.Bind_global` or use `Pusher.BindGlobal` filtered by channel name.

---

### 3.7 Presence payload normalization

**File:** `internal/channel/pusher_presence_channel.go`

| Callback | JS payload | echo-go target |
|----------|------------|----------------|
| `Here` | `Object.values(data.members)` | `[]map[string]any` or `[]MemberInfo` |
| `Joining` | `member.info` | `member["info"]` or struct field |
| `Leaving` | `member.info` | same |

Parse `subscription_succeeded` data:

```go
// ponytail: shape follows pusher-go; assert members map exists
members := data.(map[string]any)["members"]
// convert to slice of member values
```

**Acceptance criteria:**

- [ ] `Here` callback receives slice of member info objects, not raw pusher map
- [ ] `Joining`/`Leaving` receive info object only
- [ ] Document payload types in godoc

---

### 3.8 Update `Channel` interface

**File:** `internal/channel/channel.go`

Add all new methods to `Channel` interface. Update `NullChannel` with no-op chainable stubs.

---

## Files touched

| File | Changes |
|------|---------|
| `internal/channel/channel.go` | Extended interface |
| `internal/channel/pusher_channel.go` | Whisper fix, new methods |
| `internal/channel/pusher_presence_channel.go` | Payload normalization |
| `internal/channel/null_channel.go` | Stub new methods |
| `internal/util/event_formatter.go` | Tests only (if edge cases found) |

---

## Tests (required)

ROADMAP explicitly calls for table-driven tests here.

| Test file | Cases |
|-----------|-------|
| `internal/util/event_formatter_test.go` | `.client-*`, `\\`, namespace on/off |
| `internal/channel/pusher_channel_test.go` | Whisper bind name; notification event constant |
| `internal/channel/pusher_presence_channel_test.go` | `here`/`joining`/`leaving` payload shaping with fixture JSON |

Example formatter cases:

| Input event | Namespace | Expected bind name |
|-------------|-----------|-------------------|
| `.client-typing` | `App.Events` | `client-typing` |
| `OrderCreated` | `App.Events` | `App.Events.OrderCreated` |
| `.Illuminate\Notifications\...` | `App.Events` | `Illuminate\Notifications\...` |

---

## Definition of done

- [ ] `Channel` interface matches Laravel Echo Pusher channel surface (minus encrypted-specific behavior in Phase 5)
- [ ] All tests above pass
- [ ] `go test ./...` passes
- [ ] Godoc on new methods references Laravel Echo equivalents

---

## Out of scope

- Encrypted private channel listen/decrypt (Phase 5)
- User `Signin` (Phase 4)
