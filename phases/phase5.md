# Phase 5 — Encrypted Private Channels

**Goal:** `EncryptedPrivate(name)` for `private-encrypted-` channels.

**Priority:** Fifth (can swap with Phase 4).

**Depends on:** Phase 2 (registry keys `private-encrypted-{name}`). pusher-go `with-encryption` subpackage.

---

## Background

Laravel Echo exposes `encryptedPrivate(name)` which prefixes `private-encrypted-` and decrypts events using a shared secret returned from the channel auth endpoint. pusher-go provides encrypted channels via a separate import:

```go
import pusherenc "github.com/H-0-O/pusher-go/with-encryption"
```

Whisper on encrypted private channels is trigger-only (same as JS) — no encrypted client events.

---

## Current state

| Area | Status |
|------|--------|
| `EncryptedPrivate` on Echo | ❌ |
| `PusherEncryptedPrivateChannel` | ❌ |
| Registry key for encrypted | ❌ (scheme defined in Phase 2) |
| `NullEncryptedPrivateChannel` | ❌ |
| pusher-go Nacl / encryption config | Available in `with-encryption` |

---

## Tasks

### 5.1 Dependency and build layout

**File:** `go.mod`

```bash
go get github.com/H-0-O/pusher-go/with-encryption
```

**Architecture note (from ROADMAP):** No new echo-go packages unless required. Prefer:

- `internal/channel/pusher_encrypted_private_channel.go`
- Optional: encrypted client type in connector if pusher-go encrypted `Pusher` differs from standard `Pusher`

**Research spike:**

- [ ] Determine if `with-encryption` wraps standard `Pusher` or replaces it
- [ ] Identify `Nacl` / master key config requirements for Reverb

---

### 5.2 `EncryptedPrivate(name string) Channel`

**File:** `echo.go`

```go
func (e *Echo) EncryptedPrivate(name string) Channel {
    return e.connector.EncryptedPrivateChannel(name)
}
```

Strip `private-encrypted-` prefix if caller passes it (match `Private` / `Presence` prefix behavior).

---

### 5.3 `PusherEncryptedPrivateChannel`

**File:** `internal/channel/pusher_encrypted_private_channel.go`

- Embed or wrap `PusherPrivateChannel`
- Apply `private-encrypted-` prefix on subscribe
- Use pusher-go encrypted channel subscribe API
- Decrypt incoming events transparently (pusher-go handles crypto)
- `Listen`, `ListenForWhisper`, `StopListening`, etc. inherit from base channel

**Whisper:**

```go
func (c *PusherEncryptedPrivateChannel) Whisper(event string, data interface{}) Channel {
    // trigger only — same as JS, no encryption on client events
}
```

**Acceptance criteria:**

- [ ] Subscribe hits `/broadcasting/auth` with `private-encrypted-{name}`
- [ ] Encrypted payloads decrypted before callback
- [ ] `ListenForWhisper` uses `.client-{event}` formatter (Phase 3)

---

### 5.4 Connector registry and interface

**Files:** `internal/connector/connector.go`, `pusher_connector.go`, `null_connector.go`

```go
EncryptedPrivateChannel(name string) channel.Channel
```

Registry key: `private-encrypted-{name}` (Phase 2).

`Leave(channel)` must include encrypted variant (implemented in Phase 2 — verify here).

**Acceptance criteria:**

- [ ] Distinct from `Private(name)` cache entry
- [ ] Second `EncryptedPrivate("secrets")` returns same instance

---

### 5.5 `NullEncryptedPrivateChannel`

**File:** `internal/channel/null_channel.go` (or `null_encrypted_private_channel.go` if cleaner)

No-op chainable type implementing full `Channel` interface including Phase 3 methods.

`NullConnector.EncryptedPrivateChannel` returns this type (not plain `NullChannel` if type distinction matters for tests).

---

### 5.6 Pusher connector client initialization

If encrypted channels require encrypted pusher client from construction:

- Extend `PusherConfig` with encryption flags
- Use `pusherenc.NewPusher` when any encrypted channel is used, **or** always use encrypt-capable client (simpler, ponytail: slightly heavier client)

Document chosen approach in code comment (`ponytail:` if lazy default).

---

## Files touched

| File | Changes |
|------|---------|
| `go.mod`, `go.sum` | `with-encryption` dependency |
| `echo.go` | `EncryptedPrivate` |
| `internal/connector/connector.go` | `EncryptedPrivateChannel` |
| `internal/connector/pusher_connector.go` | Registry + client setup |
| `internal/connector/null_connector.go` | Null encrypted channel |
| `internal/channel/pusher_encrypted_private_channel.go` | New |
| `internal/channel/null_channel.go` | `NullEncryptedPrivateChannel` |

---

## Tests

| Test | Purpose |
|------|---------|
| `TestEncryptedPrefix` | Name normalization |
| `TestRegistryDistinctEncrypted` | Encrypted vs private keys |
| Integration with Reverb + Laravel `encryptedPrivate` channel | Decrypt round-trip |

Integration may defer to Phase 7 if Reverb encryption setup is heavy.

---

## Definition of done

- [ ] `EncryptedPrivate` on `Echo`
- [ ] Subscribe/auth/decrypt works against Reverb or pusher-go test harness
- [ ] `Leave` cleans encrypted subscriptions
- [ ] Null broadcaster supports encrypted no-op
- [ ] `go test ./...` passes

---

## Out of scope

- Custom NaCl key management beyond pusher-go defaults
- Socket.IO encrypted channels

---

## References

- Laravel Echo: `pusher-encrypted-private-channel.ts`
- pusher-go: `with-encryption/`, `spec/golang/unit/core/pusher_with_encryption_test.go`
- [Pusher encrypted channels](https://pusher.com/docs/channels/using_channels/encrypted-channels/)
