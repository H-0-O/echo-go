// HTTP example: set X-Socket-Id manually on outbound API requests.
//
// Laravel Echo injects this header via browser interceptors. In Go, callers add it themselves.
// See ROADMAP.md — "Intentional Go differences".
//
// Environment:
//
//	REVERB_APP_KEY     Reverb app key
//	API_TOKEN          Bearer token for auth and outbound API
//	REVERB_HOST        WebSocket host (default localhost)
//	REVERB_PORT        WebSocket port (default 8080)
//	AUTH_URL           Channel auth URL (default http://localhost:8000/broadcasting/auth)
//	API_URL            Outbound POST target (default http://localhost:8000/api/example)
//
// Run: go run ./examples/http_socket_id
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/H-0-O/echo-go"
)

func main() {
	key := os.Getenv("REVERB_APP_KEY")
	if key == "" {
		log.Fatal("REVERB_APP_KEY is required")
	}
	token := os.Getenv("API_TOKEN")
	if token == "" {
		log.Fatal("API_TOKEN is required")
	}

	host := envOr("REVERB_HOST", "localhost")
	port := envIntOr("REVERB_PORT", 8080)
	authURL := envOr("AUTH_URL", "http://localhost:8000/broadcasting/auth")
	apiURL := envOr("API_URL", "http://localhost:8000/api/example")

	falseVal := false
	client, err := echo.New(echo.Config{
		Broadcaster: "reverb",
		Key:         key,
		Host:        host,
		Port:        port,
		BearerToken: token,
		Auth: echo.AuthConfig{
			Endpoint: authURL,
		},
		AutoConnect: &falseVal,
	})
	if err != nil {
		log.Fatal(err)
	}

	ready := make(chan struct{})
	client.OnConnectionChange(func(status echo.ConnectionStatus) {
		if status == echo.StatusConnected && client.SocketID() != "" {
			select {
			case <-ready:
			default:
				close(ready)
			}
		}
	})

	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	select {
	case <-ready:
	case <-time.After(20 * time.Second):
		log.Fatal("timed out waiting for connected socket id")
	}

	if err := apiRequest(client, apiURL, token); err != nil {
		log.Fatal(err)
	}
	log.Printf("request sent with X-Socket-Id: %s", client.SocketID())
	_ = client.Disconnect()
}

func apiRequest(e *echo.Echo, url, token string) error {
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	if id := e.SocketID(); id != "" {
		req.Header.Set("X-Socket-Id", id)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("api status %d", resp.StatusCode)
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("%s must be an integer: %v", key, err)
	}
	return n
}
