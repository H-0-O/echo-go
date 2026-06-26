# echo-go Roadmap

Implement the **core** of [Laravel Echo](https://github.com/laravel/echo) (`packages/laravel-echo` on the `2.x` branch) as a Go library for backend services, workers, and CLIs.

**Reference:** [laravel/echo](https://github.com/laravel/echo) v2.3.x  
**Transport:** [pusher-go](https://github.com/H-0-O/pusher-go) (Pusher protocol — Reverb, Pusher, Ably-via-Pusher)

## Scope

### In scope

Everything in Laravel Echo that is **protocol, subscription, and auth logic** — the `Echo` class, `Connector` implementations, and `Channel` types under `src/connector/`, `src/channel/`, and `src/util/`.

### Out of scope (front-end / browser-only)

These exist in Laravel Echo but are not part of echo-go:

| Laravel Echo feature | Reason |
|----------------------|--------|
| `registerInterceptors()` and HTTP client hooks (axios, jQuery, Vue, Turbo) | Browser HTTP middleware; Go callers set `X-Socket-Id` on their own HTTP client |
| DOM CSRF lookup (`meta[name="csrf-token"]`, `window.Laravel`) | No DOM; accept `CSRFToken` in `Config` instead |
| `window.Pusher` / global script detection | Go always passes an explicit client |
| IIFE / NPM build pipeline | Go module distribution only |
| TypeScript typings / `dist/` | N/A for Go |
| Socket.IO broadcaster | Different protocol; only add if there is a concrete Go Socket.IO client to wrap |

---

## Current state

Rough parity with an early Laravel Echo Pusher connector. Already working:

| Area | Status |
|------|--------|
| `New`, `Connect`, `Disconnect`, `On` | ✅ |
| `Channel`, `Private`, `Presence` | ✅ |
| `Listen`, `ListenForWhisper`, `Whisper`, `StopListening`, `StopListeningForWhisper` | ✅ (with gaps — see below) |
| Presence `Here`, `Joining`, `Leaving` | ✅ (payload shape differs from JS) |
| `reverb` / `pusher` / `null` broadcasters | ✅ |
| Channel auth via `AuthEndpoint` + `AuthHeaders` | ✅ |
| Event namespace formatting (`Config.Namespace`) | ✅ |
| `NullConnector` + no-op channels | ✅ partial |

---

## Gap analysis vs Laravel Echo 2.x

### Echo client (`echo.ts`)

| API | Laravel Echo | echo-go |
|-----|--------------|---------|
| `channel(name)` | ✅ | ✅ |
| `private(name)` | ✅ | ✅ |
| `encryptedPrivate(name)` | ✅ | ❌ |
| `join(name)` (alias for presence) | ✅ | ❌ |
| `leave(name)` | ✅ | ❌ |
| `leaveChannel(name)` | ✅ | ❌ |
| `leaveAllChannels()` | ✅ | ❌ |
| `listen(channel, event, cb)` shorthand | ✅ | ❌ |
| `socketId()` | ✅ | ✅ (`SocketID`) |
| `connectionStatus()` | ✅ | ❌ |
| `onConnectionChange(cb)` → unsubscribe fn | ✅ | ❌ (`On` is lower-level) |
| Auto-connect on construct | ✅ | ❌ (explicit `Connect`) |
| `ably` broadcaster | ✅ (Pusher connector) | ❌ |
| Custom `broadcaster: function` | ✅ | ❌ |
| Invalid broadcaster → error | ✅ | ❌ (falls back to null) |

### Connector (`connector.ts`, `pusher-connector.ts`)

| API | Laravel Echo | echo-go |
|-----|--------------|---------|
| Default options (`authEndpoint`, `namespace`, …) | ✅ | ❌ partial |
| `bearerToken` → `Authorization` header merge | ✅ | ❌ (manual `AuthHeaders` only) |
| `csrfToken` → `X-CSRF-TOKEN` header merge | ✅ | ❌ |
| `userAuthentication` endpoint + headers | ✅ | ❌ |
| `signin()` (Pusher user auth) | ✅ | ❌ |
| Channel registry with prefixed keys | ✅ | ⚠️ bug: private/presence cached by unprefixed name |
| `leave` / `leaveChannel` | ✅ | ❌ |
| `connectionStatus` / `onConnectionChange` | ✅ | ❌ |

### Channel (`channel.ts`, `pusher-channel.ts`)

| API | Laravel Echo | echo-go |
|-----|--------------|---------|
| `listen(event, cb)` | ✅ | ✅ |
| `listenForWhisper` (`.client-` prefix) | ✅ | ⚠️ uses `client-` without leading `.` |
| `whisper` | ✅ | ✅ |
| `stopListening(event, cb?)` optional callback | ✅ | ❌ (always removes all handlers for event) |
| `notification(cb)` | ✅ | ❌ |
| `stopListeningForNotification(cb)` | ✅ | ❌ |
| `subscribed(cb)` | ✅ | ❌ |
| `error(cb)` | ✅ | ❌ |
| `listenToAll(cb)` | ✅ | ❌ |
| `stopListeningToAll(cb?)` | ✅ | ❌ |
| `unsubscribe()` | ✅ | ❌ (no public API) |

### Presence (`pusher-presence-channel.ts`)

| Behavior | Laravel Echo | echo-go |
|----------|--------------|---------|
| `here` callback payload | `Object.values(members)` | raw `subscription_succeeded` data |
| `joining` / `leaving` callback payload | `member.info` | raw member object |

### Encrypted private (`pusher-encrypted-private-channel.ts`)

| API | Laravel Echo | echo-go |
|-----|--------------|---------|
| `private-encrypted-` prefix | ✅ | ❌ |
| Subscribe + decrypt via shared secret from auth | ✅ | ❌ (pusher-go `with-encryption` exists) |

---

## Phased plan

### Phase 1 — Config & connection parity

**Goal:** Match Laravel Echo defaults and connection observability without changing call sites much.

- [ ] Expand `Config` to mirror `EchoOptions` (non-browser subset):
  - `Auth` struct with `Headers`
  - `AuthEndpoint` default `/broadcasting/auth`
  - `UserAuthentication` struct (`Endpoint` default `/broadcasting/user-auth`, `Headers`)
  - `BearerToken`, `CSRFToken` — merge into auth headers on `New` (like JS `setOptions`)
  - `Namespace` default `App.Events`; support disable via empty string (JS `false`)
  - `Cluster` for real Pusher (not Reverb)
- [ ] `ConnectionStatus` type: `connected`, `disconnected`, `connecting`, `reconnecting`, `failed`
- [ ] `ConnectionStatus()` on `Echo` — map from `pusher-go` `Connection.State()`
- [ ] `OnConnectionChange(cb func(ConnectionStatus)) func()` — unsubscribe handle
- [ ] Optional `AutoConnect bool` (default `true` to match JS, or document explicit `Connect` as intentional Go idiom)
- [ ] `ably` broadcaster → same `PusherConnector` with empty cluster (JS behavior)
- [ ] Unknown broadcaster → return error from `New` instead of silent null connector

**Depends on:** pusher-go connection state events (already available).

---

### Phase 2 — Channel lifecycle

**Goal:** Subscribe/unsubscribe and registry behavior match Laravel Echo.

- [ ] Fix channel cache keys: store `private-{name}`, `presence-{name}`, `private-encrypted-{name}` separately (JS uses prefixed keys)
- [ ] `Leave(channel string)` — leave public + private + encrypted + presence variants
- [ ] `LeaveChannel(name string)` — unsubscribe one channel, remove from registry
- [ ] `LeaveAllChannels()` — iterate registry
- [ ] `Join(name)` alias for `Presence(name)`
- [ ] `Listen(channel, event, cb)` shorthand on `Echo`
- [ ] Channel `Unsubscribe()` (internal, called by leave)

**Depends on:** pusher-go `Unsubscribe` / channel `Unsubscribe()`.

---

### Phase 3 — Channel API completeness

**Goal:** Full `Channel` interface parity for Pusher/Reverb channels.

- [ ] `ListenForWhisper` — use `.client-{event}` through formatter (leading `.` skips namespace, per JS)
- [ ] `StopListening(event, cb?)` — optional callback-specific unbind when pusher-go supports it
- [ ] `Subscribed(cb)` — bind `pusher:subscription_succeeded`
- [ ] `Error(cb)` — bind `pusher:subscription_error`
- [ ] `Notification(cb)` / `StopListeningForNotification(cb)` — Laravel broadcast notifications event:
  - `.Illuminate\Notifications\Events\BroadcastNotificationCreated`
- [ ] `ListenToAll(cb func(event string, data any))` — wire `pusher-go` `BindGlobal` per channel, skip `pusher:*`, strip namespace like JS
- [ ] `StopListeningToAll(cb?)`
- [ ] Presence payload normalization:
  - `Here` → `[]member` from `data.members`
  - `Joining` / `Leaving` → `member.info`

**Tests:** table-driven tests for event formatter + whisper prefix + presence payload shaping.

---

### Phase 4 — Authentication

**Goal:** Laravel broadcasting auth for channels and Pusher user authentication.

- [ ] Wire `UserAuthentication` config into pusher-go `UserAuthentication` options
- [ ] `Signin()` on `Echo` / `PusherConnector` — delegate to `pusher.Signin()`
- [ ] Ensure channel auth POST body matches Laravel (`socket_id`, `channel_name`) — already via pusher-go; verify against Reverb
- [ ] Document how Go services pass `Authorization`, `X-CSRF-TOKEN`, cookies via `Auth.Headers`

**Depends on:** pusher-go user auth (implemented).

---

### Phase 5 — Encrypted private channels

**Goal:** `EncryptedPrivate(name)` for `private-encrypted-` channels.

- [ ] `EncryptedPrivate(name string) Channel` on `Echo`
- [ ] `PusherEncryptedPrivateChannel` wrapping pusher-go encrypted channel support
- [ ] `private-encrypted-` prefix applied automatically
- [ ] `Whisper` on encrypted private (trigger only — same as JS)
- [ ] `NullEncryptedPrivateChannel` for null broadcaster

**Depends on:** pusher-go `with-encryption` import path / `Nacl` config if required.

---

### Phase 6 — Null broadcaster & custom connectors

**Goal:** Test/dev ergonomics and extensibility.

- [ ] Complete `NullChannel`, `NullPrivateChannel`, `NullPresenceChannel`, `NullEncryptedPrivateChannel` — mirror JS no-op chaining
- [ ] Custom connector: `Config.Broadcaster` as injectable `Connector` interface (Go equivalent of `broadcaster: function`)
- [ ] Export `Connector` interface from root package only if needed for custom backends (keep `internal/` otherwise)

---

### Phase 7 — Quality & docs

**Goal:** Confidence and adoption.

- [ ] Integration test against Reverb (or pusher-go test harness): public, private, presence subscribe
- [ ] Example: worker subscribing to `private-App.Models.User.{id}`
- [ ] Example: passing `X-Socket-Id` on outbound HTTP (manual interceptor pattern for Go)
- [ ] README: parity table, intentional differences (explicit `Connect`, no DOM auth)
- [ ] CHANGELOG

---

## Architecture (target)

Mirror Laravel Echo layering; keep `internal/` hidden.

```
echo-go/
├── echo.go                 # Echo, Config, ConnectionStatus — public API
├── internal/
│   ├── connector/
│   │   ├── connector.go    # Connector interface + default options merge
│   │   ├── pusher_connector.go
│   │   └── null_connector.go
│   ├── channel/
│   │   ├── channel.go      # Channel + PresenceChannel interfaces
│   │   ├── pusher_channel.go
│   │   ├── pusher_private_channel.go
│   │   ├── pusher_encrypted_private_channel.go
│   │   ├── pusher_presence_channel.go
│   │   └── null_*.go
│   └── util/
│       └── event_formatter.go
```

No new packages unless encrypted channels require a separate pusher-go import path.

---

## Priority order

```
Phase 1 (config + connection status)
  → Phase 2 (leave / registry fixes)
  → Phase 3 (channel API)
  → Phase 4 (user auth)
  → Phase 5 (encrypted)
  → Phase 6 (null + custom)
  → Phase 7 (tests + docs)
```

Phases 4 and 5 can swap if encrypted channels are not needed yet.

---

## Intentional Go differences (keep)

| Topic | Laravel Echo | echo-go |
|-------|--------------|---------|
| HTTP interceptors | Auto-inject `X-Socket-Id` | Caller adds header to their `http.Client` |
| CSRF / bearer | Auto from DOM or options | Explicit `Config` fields only |
| Connect timing | Constructor calls `connect()` | Explicit `Connect()` unless `AutoConnect` added |
| Callback types | `CallableFunction` | `func(data any)` |
| Socket.IO | First-class broadcaster | Out of scope unless requested |

---

## Success criteria

echo-go is **core-complete** when a Go program can:

1. Connect to Reverb/Pusher with Laravel-compatible channel and user auth.
2. Subscribe to public, private, presence, and encrypted private channels.
3. Listen, whisper, and leave channels with the same naming rules as Laravel Echo.
4. Observe connection status the same way as `connectionStatus()` / `onConnectionChange()`.
5. Run in CI with a null broadcaster and without a browser.

---

## References

- [Laravel Echo source (`2.x`)](https://github.com/laravel/echo/tree/2.x/packages/laravel-echo/src)
- [Laravel broadcasting — client installation](https://laravel.com/docs/broadcasting#client-side-installation)
- [Laravel Reverb](https://laravel.com/docs/reverb)
- [Pusher user authentication](https://pusher.com/docs/channels/using_channels/user-authentication/)
