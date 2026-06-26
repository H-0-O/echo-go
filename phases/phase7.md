# Phase 7 — Quality & Docs

**Goal:** Confidence and adoption.

**Priority:** Last — validates Phases 1–6 end-to-end and documents intentional differences from Laravel Echo.

**Depends on:** Phases 1–6 complete (or explicitly deferred items documented).

---

## Background

echo-go is **core-complete** when a Go program can connect to Reverb/Pusher with Laravel-compatible auth, subscribe to all channel types, listen/whisper/leave with JS naming rules, observe connection status, and run in CI with the null broadcaster.

This phase adds integration coverage, runnable examples, README parity documentation, and a CHANGELOG so adopters can trust behavior and track releases.

---

## Current state

| Area | Status |
|------|--------|
| Unit tests | Minimal — formatter/channel tests from Phase 3 only |
| Integration tests vs Reverb | ❌ |
| Worker / HTTP examples | ❌ |
| README parity table | ❌ (basic API table only) |
| Intentional differences doc | Partial in ROADMAP.md |
| CHANGELOG | ❌ |

---

## Tasks

### 7.1 Integration test harness

**Goal:** Subscribe to public, private, and presence channels against a real or test Pusher-protocol server.

**Options (pick smallest that works):**

| Approach | When to use |
|----------|-------------|
| pusher-go test env / helpers | Reuse `spec/golang/helpers/pusher_test_env.go` patterns if importable |
| Docker Reverb + Laravel app | Full fidelity; heavier CI |
| `testing.Short()` gate | Unit tests always run; integration skipped locally without env |

**Suggested layout:**

```
internal/integration/
  reverb_test.go   // build tag: //go:build integration
```

Or top-level:

```
spec/integration/echo_reverb_test.go
```

**Env vars:**

| Variable | Purpose |
|----------|---------|
| `ECHO_TEST_REVERB_HOST` | WebSocket host |
| `ECHO_TEST_REVERB_PORT` | Port |
| `ECHO_TEST_REVERB_KEY` | App key |
| `ECHO_TEST_AUTH_URL` | Channel auth endpoint |
| `ECHO_TEST_AUTH_TOKEN` | Bearer token for auth |

**Test cases:**

1. **Public** — connect, subscribe `test-channel`, trigger event from server or wait for fixture, assert callback
2. **Private** — subscribe `private-test`, verify auth POST succeeds
3. **Presence** — subscribe `presence-room`, `Here` receives member slice (Phase 3 shape), `Joining`/`Leaving` on second client optional
4. **Connection status** — `OnConnectionChange` fires `connecting` → `connected`
5. **Leave** — subscribe, `Leave`, assert no further events

Encrypted channel test optional if Reverb encryption setup is available (defer with `t.Skip`).

**Acceptance criteria:**

- [ ] Integration tests exist and pass with env configured
- [ ] `go test ./...` passes without env (integration skipped)
- [ ] CI job documented (GitHub Actions or README note) — optional if no CI yet

---

### 7.2 Example: worker on private channel

**File:** `examples/worker/main.go`

Scenario: background worker subscribes to `private-App.Models.User.{id}` (Laravel convention).

```go
client, err := echo.New(echo.Config{
    Broadcaster:  "reverb",
    Key:          os.Getenv("REVERB_APP_KEY"),
    Host:         "localhost",
    Port:         8080,
    BearerToken:  os.Getenv("API_TOKEN"),       // Phase 1
    AuthEndpoint: "http://localhost:8000/broadcasting/auth",
    AutoConnect:  true,
})
// ...
client.Private(fmt.Sprintf("App.Models.User.%d", userID)).
    Listen("NotificationSent", handleNotification)
```

Demonstrate:

- Explicit `Connect` vs `AutoConnect`
- `Private` without `private-` prefix in argument
- Blocking main (signal handling or `select {}`)

**Acceptance criteria:**

- [ ] `go run ./examples/worker` compiles
- [ ] README links to example
- [ ] Uses env vars for secrets, no hardcoded tokens

---

### 7.3 Example: `X-Socket-Id` on outbound HTTP

**File:** `examples/http_socket_id/main.go`

Laravel Echo browser interceptors auto-inject `X-Socket-Id`. Go callers do this manually:

```go
func apiRequest(echo *echo.Echo, url string) error {
    req, _ := http.NewRequest(http.MethodPost, url, nil)
    if id := echo.SocketID(); id != "" {
        req.Header.Set("X-Socket-Id", id)
    }
    req.Header.Set("Authorization", "Bearer "+token)
    // ...
}
```

Show pattern after `OnConnectionChange` or `connected` event so socket ID is non-empty.

**Acceptance criteria:**

- [ ] Documents intentional Go difference from ROADMAP
- [ ] Compiles and referenced from README

---

### 7.4 README: parity table & differences

**File:** `README.md`

Add sections:

#### Parity table

Mirror ROADMAP gap analysis as **current** status (post Phase 6). Columns: API, Laravel Echo, echo-go, Notes.

Group by:

- Echo client (`channel`, `private`, `encryptedPrivate`, `join`, `leave`, `socketId`, `connectionStatus`, …)
- Connector (auth defaults, `signin`, registry)
- Channel (`listen`, `listenToAll`, `notification`, `subscribed`, …)
- Presence payloads

Mark intentional N/A for browser-only features (interceptors, DOM CSRF).

#### Intentional Go differences

| Topic | Laravel Echo | echo-go |
|-------|--------------|---------|
| HTTP interceptors | Auto `X-Socket-Id` | Caller sets header |
| CSRF / bearer | DOM or options | `Config` fields only |
| Connect timing | Constructor | `Connect()` or `AutoConnect` |
| Callback types | `CallableFunction` | `func(data any)` |
| Socket.IO | Supported | Out of scope |

#### Configuration reference

Document `Config` fields after Phase 1 expansion: defaults, `Auth`, `UserAuthentication`, `BearerToken`, `CSRFToken`, `Cluster`, `AutoConnect`, `Connector`.

#### Running tests

```bash
go test ./...
go test -tags=integration ./...  # with Reverb env vars
```

**Acceptance criteria:**

- [ ] README is accurate for implemented API
- [ ] Parity table reflects actual code, not aspirational checkboxes
- [ ] Link to `ROADMAP.md` and `phases/` for implementation detail

---

### 7.5 CHANGELOG

**File:** `CHANGELOG.md`

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

## [Unreleased]

### Added
- Phase 1: ConnectionStatus, OnConnectionChange, expanded Config
- ...

### Fixed
- Channel registry cache key collision for private/presence

### Changed
- Unknown broadcaster now returns error from New
```

**Acceptance criteria:**

- [ ] `Unreleased` section summarizes Phases 1–7
- [ ] Breaking changes called out (e.g. `New` error on bad broadcaster, `Channel` interface additions)
- [ ] First tagged release can move `Unreleased` → `v0.1.0` when ready

---

### 7.6 Godoc pass

Quick review of exported symbols in `echo.go`:

- Every public method has a one-line description
- Laravel Echo equivalent named where helpful (`Join` → "alias for Presence")
- `Config` fields documented with defaults

**Acceptance criteria:**

- [ ] `go doc github.com/H-0-O/echo-go` reads cleanly
- [ ] No stutter (`Echo.Echo` avoided)

---

### 7.7 Optional: GitHub Actions CI

**File:** `.github/workflows/test.yml` (only if user wants CI in repo)

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test ./...
```

Integration job optional, `if: env.REVERB_APP_KEY != ''`.

**Acceptance criteria:**

- [ ] Unit tests run on push
- [ ] Document if integration is manual-only

---

## Files touched

| File | Changes |
|------|---------|
| `README.md` | Parity table, config reference, examples links, test instructions |
| `CHANGELOG.md` | New |
| `examples/worker/main.go` | New |
| `examples/http_socket_id/main.go` | New |
| `spec/integration/` or `internal/integration/` | Reverb tests |
| `.github/workflows/test.yml` | Optional CI |
| `echo.go` | Godoc polish |

---

## Tests

| Test | Type |
|------|------|
| Phase 3 unit tests | Required baseline |
| `TestNullBroadcasterCI` | Smoke — null broadcaster in default `go test` |
| Reverb integration suite | Env-gated |
| Example compile check | `go build ./examples/...` in CI or Makefile |

---

## Definition of done

- [ ] Success criteria from ROADMAP.md met and verified:
  1. Connect with Laravel-compatible channel and user auth
  2. Public, private, presence, encrypted private subscribe
  3. Listen, whisper, leave with JS naming rules
  4. `ConnectionStatus` / `OnConnectionChange`
  5. CI-safe null broadcaster tests
- [ ] README parity table and intentional differences published
- [ ] Two examples compile and are linked
- [ ] CHANGELOG started
- [ ] `go test ./...` passes

---

## Out of scope

- npm / TypeScript distribution
- Socket.IO broadcaster
- Hosted godoc badge / pkg.go.dev vanity URL setup (nice-to-have)
- Production deployment guide for Reverb

---

## References

- [Laravel Echo source (`2.x`)](https://github.com/laravel/echo/tree/2.x/packages/laravel-echo/src)
- [Laravel broadcasting](https://laravel.com/docs/broadcasting)
- [Laravel Reverb](https://laravel.com/docs/reverb)
- echo-go `ROADMAP.md` — success criteria
- pusher-go integration tests: `spec/golang/integration/`
