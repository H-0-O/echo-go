// Worker example: subscribe to a Laravel private user channel from a background process.
//
// Environment:
//
//	REVERB_APP_KEY     Reverb app key
//	API_TOKEN          Bearer token for /broadcasting/auth
//	REVERB_HOST        WebSocket host (default localhost)
//	REVERB_PORT        WebSocket port (default 8080)
//	AUTH_URL           Channel auth URL (default http://localhost:8000/broadcasting/auth)
//	USER_ID            User id for private-App.Models.User.{id} (default 1)
//
// Run: go run ./examples/worker
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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
	userID := envIntOr("USER_ID", 1)

	// AutoConnect: true (default) connects in New. Alternative:
	//   AutoConnect: &false, then client.Connect() after wiring listeners.
	client, err := echo.New(echo.Config{
		Broadcaster: "reverb",
		Key:         key,
		Host:        host,
		Port:        port,
		BearerToken: token,
		Auth: echo.AuthConfig{
			Endpoint: authURL,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	channelName := fmt.Sprintf("App.Models.User.%d", userID)
	client.Private(channelName).Listen("NotificationSent", handleNotification)

	log.Printf("listening on private-%s for NotificationSent", channelName)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
	_ = client.Disconnect()
}

func handleNotification(data any) {
	log.Printf("notification: %v", data)
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
