# echo-go

A Go **library** that mirrors the [Laravel Echo](https://laravel.com/docs/broadcasting#client-side-installation) API for subscribing to broadcast channels over WebSockets — aimed at [Laravel Reverb](https://laravel.com/docs/reverb) and other Pusher-protocol servers.

Built on [github.com/H-0-O/pusher-go](https://github.com/H-0-O/pusher-go).

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

Channel methods follow Echo JS: `Listen`, `ListenForWhisper`, `Whisper`, `StopListening`, `StopListeningForWhisper`.

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

## License

MIT
