# Phase 4 — Authentication

**Goal:** Laravel broadcasting auth for channels and Pusher user authentication.

**Priority:** Fourth (can swap with Phase 5 if encrypted channels are not needed).

**Depends on:** Phase 1 (`UserAuthentication` config struct and header merging). pusher-go user auth (`Signin`, `UserAuthentication` options).

---

## Background

Laravel Echo configures two HTTP auth endpoints:

| Endpoint | Default | Purpose |
|----------|---------|---------|
| Channel auth | `/broadcasting/auth` | Private/presence/encrypted subscribe |
| User auth | `/broadcasting/user-auth` | Pusher user authentication (`signin`) |

Go services do not have DOM CSRF or axios interceptors. Callers pass `Authorization`, `X-CSRF-TOKEN`, and cookies explicitly via `Config.Auth.Headers` and `Config.UserAuthentication.Headers`, plus convenience fields `BearerToken` and `CSRFToken` (merged in Phase 1).

Channel auth POST body (`socket_id`, `channel_name`) is handled by pusher-go; this phase verifies Reverb compatibility and documents the Go pattern.

---

## Current state

| Area | Status |
|------|--------|
| Channel `AuthEndpoint` + `AuthHeaders` | ✅ via `PusherConnector` |
| `UserAuthentication` config | ❌ struct from Phase 1 not wired |
| `Signin()` on Echo | ❌ |
| Bearer/CSRF auto-merge | ❌ (Phase 1) |
| Docs for Go HTTP clients | ❌ |

---

## Tasks

### 4.1 Wire `UserAuthentication` into pusher-go

**File:** `internal/connector/pusher_connector.go`

Pass Phase 1 merged config:

```go
UserAuthentication: types.UserAuthenticationConfig{
    Endpoint:  config.UserAuthentication.Endpoint,
    Transport: "ajax",
    Headers:   mergedUserAuthHeaders,
},
```

**Acceptance criteria:**

- [ ] Default endpoint `/broadcasting/user-auth` when empty
- [ ] Custom headers reach pusher-go user authenticator
- [ ] Same transport as channel auth (`ajax` = HTTP POST)

---

### 4.2 `Signin()` on Echo and connector

**Files:** `echo.go`, `internal/connector/connector.go`, `pusher_connector.go`

```go
func (e *Echo) Signin()
```

Delegate to `pusher.Pusher.Signin()` (async in pusher-go — document whether Echo blocks or fires-and-forgets; match JS fire-and-forget).

**Connector interface:**

```go
Signin()
```

- `PusherConnector`: `c.client.Signin()`
- `NullConnector`: no-op

**Acceptance criteria:**

- [ ] Callable after `Connect`
- [ ] Uses user-auth endpoint and headers from config
- [ ] Document interaction with `pusher:signin_success` / `pusher:signin_failure` (optional: expose via `On` or dedicated callbacks later)

---

### 4.3 Verify channel auth POST against Reverb

**Task:** Manual or integration test (full harness in Phase 7).

Confirm pusher-go sends:

```json
{
  "socket_id": "123.456",
  "channel_name": "private-App.Models.User.1"
}
```

To `Auth.Endpoint` with configured headers.

**Acceptance criteria:**

- [ ] Private channel subscribe succeeds against Laravel Reverb with `routes/channels.php` auth
- [ ] Document any header requirements (`Accept`, `Content-Type`, cookies)

---

### 4.4 Document Go auth patterns

**File:** `README.md` (or `docs/auth.md` if README stays short — prefer README section per Phase 7)

Topics:

1. **Bearer token** — `Config.BearerToken` or `Auth.Headers["Authorization"]`
2. **CSRF** — `Config.CSRFToken` for session-based Laravel apps
3. **Cookies** — pass `Cookie` header on both auth endpoints for Sanctum/session
4. **Custom endpoints** — `Auth.Endpoint`, `UserAuthentication.Endpoint`
5. **No interceptors** — echo-go does not wrap `http.Client`; caller sets `X-Socket-Id` on outbound API requests using `echo.SocketID()`

Example snippet:

```go
req.Header.Set("X-Socket-Id", echo.SocketID())
req.Header.Set("Authorization", "Bearer "+token)
```

**Acceptance criteria:**

- [ ] README section committed (may complete fully in Phase 7)
- [ ] Covers channel + user auth

---

## Files touched

| File | Changes |
|------|---------|
| `echo.go` | `Signin()` |
| `internal/connector/connector.go` | `Signin()` on interface |
| `internal/connector/pusher_connector.go` | User auth options, `Signin` impl |
| `internal/connector/null_connector.go` | No-op `Signin` |
| `README.md` | Auth documentation |

---

## Tests

| Test | Type |
|------|------|
| `TestUserAuthConfigPassthrough` | Unit — options passed to pusher-go |
| `TestSigninNullConnector` | Unit — no panic |
| Reverb private + user auth | Integration — Phase 7 or minimal test here |

---

## Definition of done

- [ ] `Signin()` exposed on `Echo`
- [ ] User authentication endpoint configured like Laravel Echo
- [ ] Channel auth verified against Reverb (or documented blocker)
- [ ] Go auth header patterns documented
- [ ] `go test ./...` passes

---

## Out of scope

- HTTP client middleware / automatic `X-Socket-Id` injection (intentional Go difference)
- Encrypted channel shared secret handling (Phase 5 — uses same channel auth endpoint)

---

## References

- [Laravel broadcasting — authorizing channels](https://laravel.com/docs/broadcasting#authorizing-channels)
- [Pusher user authentication](https://pusher.com/docs/channels/using_channels/user-authentication/)
- pusher-go: `examples/connect/main.go` `UserAuthentication` option
