# echo-go

A Go **library** that mirrors the [Laravel Echo](https://laravel.com/docs/broadcasting#client-side-installation) API for subscribing to broadcast channels over WebSockets — aimed at [Laravel Reverb](https://laravel.com/docs/reverb) and other Pusher-protocol servers.

Built on [github.com/H-0-O/pusher-go](https://github.com/H-0-O/pusher-go).

> **Vibe coded.** This library was built iteratively with AI-assisted development — shaped toward Laravel Echo parity, not transcribed line-by-line from the JS source. It is tested and documented, but the process was fast-and-loose by design. Check the parity tables below and [ROADMAP.md](ROADMAP.md) for what is done vs. planned; open an issue if the vibes lie.

## Installation

```bash
go get github.com/H-0-O/echo-go
```

Import only the root package:

```go
import "github.com/H-0-O/echo-go"
```

Implementation packages live under `internal/` and are not part of the public API.

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/H-0-O/echo-go"
)

func main() {
	client, err := echo.New(echo.Config{
		Broadcaster: "reverb",
		Key:         "your-reverb-key",
		Host:        "localhost",
		Port:        8080,
		TLS:         false,
		BearerToken: "your-token",
		Auth: echo.AuthConfig{
			Endpoint: "http://localhost:8000/broadcasting/auth",
		},
		UserAuthentication: echo.UserAuthConfig{
			Endpoint: "http://localhost:8000/broadcasting/user-auth",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	client.On("connected", func(_ any) {
		fmt.Println("connected:", client.SocketID())
		client.Signin()
	})

	client.Connect()

	client.Channel("orders").Listen("OrderPlaced", func(data any) {
		fmt.Printf("Order placed: %v\n", data)
	})

	client.Private("user.1").Listen("MessageSent", func(data any) {
		fmt.Printf("New message: %v\n", data)
	})

	select {}
}
```

## API

| Method | Description |
|--------|-------------|
| `New(Config)` | Create a client (`reverb`, `pusher`, `ably`, or `null` broadcaster) |
| `Connect` / `Disconnect` | WebSocket lifecycle |
| `On(event, cb)` | Connection events (`connecting`, `connected`, `disconnected`, `error`, …) |
| `Signin()` | Pusher user authentication (fire-and-forget; call when connected) |
| `SocketID()` | Current socket ID |
| `Channel(name)` | Public channel |
| `Private(name)` | Private channel (no `private-` prefix in the name you pass) |
| `Presence(name)` | Presence channel + `Here` / `Joining` / `Leaving` |

Channel methods follow Echo JS: `Listen`, `ListenForWhisper`, `Whisper`, `StopListening`, `StopListeningForWhisper`, `Subscribed`, `Error`, `Notification`, `ListenToAll`, and matching `Stop*` variants.

Echo client also exposes `EncryptedPrivate`, `Join`, `Leave`, `LeaveChannel`, `LeaveAllChannels`, `Listen` (shorthand), `ConnectionStatus`, and `OnConnectionChange`.

## Examples

| Example | Description |
|---------|-------------|
| [`examples/worker`](examples/worker/main.go) | Background worker on `private-App.Models.User.{id}` |
| [`examples/http_socket_id`](examples/http_socket_id/main.go) | Manual `X-Socket-Id` on outbound HTTP |

Worker env vars: `REVERB_APP_KEY`, `API_TOKEN`, optional `REVERB_HOST`, `REVERB_PORT`, `AUTH_URL`, `USER_ID`.

HTTP example env vars: same plus `API_URL` for the outbound POST target.

```bash
go run ./examples/worker
go run ./examples/http_socket_id
```

## Parity with Laravel Echo

Current status vs [Laravel Echo 2.x](https://github.com/laravel/echo/tree/2.x/packages/laravel-echo/src). See [ROADMAP.md](ROADMAP.md) and [phases/](phases/) for implementation history.

### Echo client

| API | Laravel Echo | echo-go | Notes |
|-----|--------------|---------|-------|
| `channel(name)` | ✅ | ✅ | |
| `private(name)` | ✅ | ✅ | Prefix applied automatically |
| `encryptedPrivate(name)` | ✅ | ✅ | `EncryptedPrivate` |
| `join(name)` | ✅ | ✅ | Alias for `Presence` |
| `leave(name)` | ✅ | ✅ | Leaves all variants of a logical name |
| `leaveChannel(name)` | ✅ | ✅ | Exact registry name |
| `leaveAllChannels()` | ✅ | ✅ | |
| `listen(channel, event, cb)` | ✅ | ✅ | Shorthand on `Echo` |
| `socketId()` | ✅ | ✅ | `SocketID()` |
| `connectionStatus()` | ✅ | ✅ | `ConnectionStatus()` |
| `onConnectionChange(cb)` | ✅ | ✅ | Returns unsubscribe func |
| Auto-connect on construct | ✅ | ✅ | Default `AutoConnect: true`; set `false` for explicit `Connect()` |
| `ably` broadcaster | ✅ | ✅ | Same Pusher connector, empty cluster |
| Custom `broadcaster: function` | ✅ | ✅ | `Config.Connector` injection |
| Invalid broadcaster → error | ✅ | ✅ | `New` returns error |
| HTTP interceptors | ✅ | N/A | Browser-only; see intentional differences |
| Socket.IO broadcaster | ✅ | ❌ | Out of scope |

### Connector

| API | Laravel Echo | echo-go | Notes |
|-----|--------------|---------|-------|
| Default auth endpoints | ✅ | ✅ | `/broadcasting/auth`, `/broadcasting/user-auth` |
| `bearerToken` header merge | ✅ | ✅ | `Config.BearerToken` |
| `csrfToken` header merge | ✅ | ✅ | `Config.CSRFToken` |
| `userAuthentication` | ✅ | ✅ | `Config.UserAuthentication` |
| `signin()` | ✅ | ✅ | `Signin()` |
| Channel registry (prefixed keys) | ✅ | ✅ | |
| `leave` / `leaveChannel` | ✅ | ✅ | |

### Channel

| API | Laravel Echo | echo-go | Notes |
|-----|--------------|---------|-------|
| `listen(event, cb)` | ✅ | ✅ | |
| `listenForWhisper` | ✅ | ✅ | `.client-{event}` via formatter |
| `whisper` | ✅ | ✅ | |
| `stopListening(event, cb?)` | ✅ | ✅ | Optional callback unbind |
| `notification(cb)` | ✅ | ✅ | Laravel broadcast notifications event |
| `subscribed(cb)` | ✅ | ✅ | |
| `error(cb)` | ✅ | ✅ | |
| `listenToAll(cb)` | ✅ | ✅ | |
| `stopListeningToAll(cb?)` | ✅ | ✅ | |
| `unsubscribe()` | ✅ | ⚠️ | Internal; use `Leave` / `LeaveChannel` on `Echo` |

### Presence

| Behavior | Laravel Echo | echo-go | Notes |
|----------|--------------|---------|-------|
| `here` payload | `Object.values(members)` | ✅ | `[]any` of member info |
| `joining` / `leaving` payload | `member.info` | ✅ | Info object only |

### Encrypted private

| API | Laravel Echo | echo-go | Notes |
|-----|--------------|---------|-------|
| `private-encrypted-` prefix | ✅ | ✅ | |
| Subscribe + decrypt | ✅ | ✅ | Via pusher-go encryption |

## Intentional Go differences

| Topic | Laravel Echo | echo-go |
|-------|--------------|---------|
| HTTP interceptors | Auto `X-Socket-Id` | Caller sets header on their `http.Client` |
| CSRF / bearer | DOM or options | `Config` fields only |
| Connect timing | Constructor | `Connect()` or default `AutoConnect` |
| Callback types | `CallableFunction` | `func(data any)` |
| Socket.IO | Supported | Out of scope |

## Configuration reference

| Field | Default | Description |
|-------|---------|-------------|
| `Broadcaster` | — | `"reverb"`, `"pusher"`, `"ably"`, or `"null"` |
| `Key` | — | Reverb/Pusher app key |
| `Cluster` | `"mt1"` (pusher-go) | Pusher cluster; ignored when `Host` set for Reverb |
| `Host` | — | WebSocket host for Reverb/self-hosted |
| `Port` | — | WebSocket port |
| `TLS` | `false` | Use WSS |
| `Namespace` | `"App.Events"` | Event name prefix; `&""` disables |
| `Auth.Endpoint` | `"/broadcasting/auth"` | Channel authorization URL |
| `Auth.Headers` | merged | Extra headers; Bearer/CSRF merged on `New` |
| `UserAuthentication.Endpoint` | `"/broadcasting/user-auth"` | Pusher user auth URL |
| `UserAuthentication.Headers` | merged | Same merge rules as `Auth.Headers` |
| `BearerToken` | — | Merged as `Authorization: Bearer …` |
| `CSRFToken` | — | Merged as `X-CSRF-TOKEN` |
| `AutoConnect` | `true` | Connect in `New` when true/nil |
| `Connector` | `nil` | Custom backend; ignores `Broadcaster` |
| `AuthEndpoint` | — | Deprecated; use `Auth.Endpoint` |
| `AuthHeaders` | — | Deprecated; use `Auth.Headers` |

## Running tests

Unit tests (no WebSocket server required):

```bash
go test ./...
```

Integration tests against Reverb + Laravel auth (skipped when env is unset):

```bash
export ECHO_TEST_REVERB_HOST=localhost
export ECHO_TEST_REVERB_PORT=8080
export ECHO_TEST_REVERB_KEY=your-key
export ECHO_TEST_AUTH_URL=http://localhost:8000/broadcasting/auth
export ECHO_TEST_AUTH_TOKEN=your-token
# Optional — for event trigger tests:
export ECHO_TEST_REVERB_APP_ID=your-app-id
export ECHO_TEST_REVERB_SECRET=your-secret

go test -tags=integration ./internal/integration/...
```

CI runs unit tests and `go build ./examples/...` on push. Integration tests are manual unless repo secrets are configured.

## Authentication

Laravel Echo configures two HTTP auth endpoints. echo-go mirrors both via `Config.Auth` and `Config.UserAuthentication`:

| Endpoint config | Default | Purpose |
|-----------------|---------|---------|
| `Auth.Endpoint` | `/broadcasting/auth` | Private/presence/encrypted channel subscribe |
| `UserAuthentication.Endpoint` | `/broadcasting/user-auth` | Pusher user authentication (`Signin`) |

Go has no browser CSRF or axios interceptors — pass headers explicitly.

### Bearer token

Use the convenience field or set the header directly:

```go
echo.Config{BearerToken: "your-token"}
// merged as Authorization: Bearer your-token on both auth endpoints

echo.Config{
	Auth: echo.AuthConfig{
		Headers: map[string]string{"Authorization": "Bearer your-token"},
	},
}
```

Caller values in `Auth.Headers` / `UserAuthentication.Headers` override merged defaults.

### CSRF (session-based Laravel)

```go
echo.Config{CSRFToken: "your-csrf-token"}
// merged as X-CSRF-TOKEN on both auth endpoints
```

### Cookies (Sanctum / session)

Pass the session cookie on both auth endpoints (Bearer/CSRF merge does not include cookies):

```go
cookie := "laravel_session=..."
echo.Config{
	Auth: echo.AuthConfig{
		Headers: map[string]string{"Cookie": cookie},
	},
	UserAuthentication: echo.UserAuthConfig{
		Headers: map[string]string{"Cookie": cookie},
	},
}
```

### Channel auth wire format

Private/presence subscribe auth is an HTTP POST to `Auth.Endpoint`. pusher-go sends `application/x-www-form-urlencoded`:

```
socket_id=123.456&channel_name=private-App.Models.User.1
```

`Content-Type` is set by pusher-go. Laravel typically expects `Accept: application/json` — add it to `Auth.Headers` if your app requires it.

### User auth (`Signin`)

`Signin()` POSTs to `UserAuthentication.Endpoint` with `socket_id` only (form-urlencoded). Call it after connect — typically inside a `"connected"` callback. It is fire-and-forget; pusher-go handles `pusher:signin` / `pusher:signin_success` internally. `Echo.On` does not expose signin events yet.

### Outbound API requests (no interceptors)

echo-go does not wrap `http.Client` or inject headers automatically. Set `X-Socket-Id` on your own HTTP requests so Laravel can exclude the current socket from broadcast fan-out:

```go
req.Header.Set("X-Socket-Id", client.SocketID())
req.Header.Set("Authorization", "Bearer "+token)
```

## How it works

1. **Protocol** — Reverb speaks the Pusher WebSocket protocol; echo-go delegates transport to pusher-go.
2. **Authentication** — Private/presence channels use pusher-go `ChannelAuthorization`, wired from `Config.Auth`. User auth uses `Config.UserAuthentication` and `Signin()`.
3. **Namespaces** — `Config.Namespace` prefixes event names (Laravel-style `.Event` syntax supported via the formatter).

## Further reading

- [ROADMAP.md](ROADMAP.md) — scope, gap analysis, success criteria
- [phases/](phases/) — phased implementation notes
- [CHANGELOG.md](CHANGELOG.md) — release history

## License

MIT
