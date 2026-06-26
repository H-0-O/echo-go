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
		Broadcaster:  "reverb",
		Key:          "your-reverb-key",
		Host:         "localhost",
		Port:         8080,
		TLS:          false,
		AuthEndpoint: "http://localhost:8000/broadcasting/auth",
		AuthHeaders: map[string]string{
			"Authorization": "Bearer your-token",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	client.On("connected", func(_ any) {
		fmt.Println("connected:", client.SocketID())
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
| `New(Config)` | Create a client (`reverb`, `pusher`, or `null` broadcaster) |
| `Connect` / `Disconnect` | WebSocket lifecycle |
| `On(event, cb)` | Connection events (`connecting`, `connected`, `disconnected`, `error`, …) |
| `SocketID()` | Current socket ID |
| `Channel(name)` | Public channel |
| `Private(name)` | Private channel (no `private-` prefix in the name you pass) |
| `Presence(name)` | Presence channel + `Here` / `Joining` / `Leaving` |

Channel methods follow Echo JS: `Listen`, `ListenForWhisper`, `Whisper`, `StopListening`, `StopListeningForWhisper`.

## How it works

1. **Protocol** — Reverb speaks the Pusher WebSocket protocol; echo-go delegates transport to pusher-go.
2. **Authentication** — Private/presence channels use pusher-go `ChannelAuthorization`, wired from `Config.AuthEndpoint` and `Config.AuthHeaders` (form-urlencoded POST, matching Laravel broadcasting auth).
3. **Namespaces** — `Config.Namespace` prefixes event names (Laravel-style `.Event` syntax supported via the formatter).

## License

MIT
