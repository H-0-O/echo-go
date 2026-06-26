// Package echo is a Laravel Echo–style client library for Laravel Reverb and
// Pusher-compatible WebSocket broadcasters.
package echo

import (
	"github.com/H-0-O/echo-go/internal/channel"
	"github.com/H-0-O/echo-go/internal/connector"
)

// Channel is a broadcast channel subscription.
type Channel = channel.Channel

// PresenceChannel is a presence broadcast channel subscription.
type PresenceChannel = channel.PresenceChannel

// Config configures a new Echo client.
type Config struct {
	Broadcaster  string
	Host         string
	Key          string
	Port         int
	Namespace    string
	AuthEndpoint string
	AuthHeaders  map[string]string
	TLS          bool
}

// Echo is the main client. Create one with [New].
type Echo struct {
	connector connector.Connector
	config    Config
}

// New creates an Echo client for the given broadcaster configuration.
func New(config Config) (*Echo, error) {
	var conn connector.Connector

	switch config.Broadcaster {
	case "reverb", "pusher":
		pusherConn, err := connector.NewPusherConnector(connector.PusherConfig{
			Key:          config.Key,
			Host:         config.Host,
			Port:         config.Port,
			Namespace:    config.Namespace,
			AuthEndpoint: config.AuthEndpoint,
			AuthHeaders:  config.AuthHeaders,
			TLS:          config.TLS,
		})
		if err != nil {
			return nil, err
		}
		conn = pusherConn
	case "null":
		conn = connector.NewNullConnector()
	default:
		conn = connector.NewNullConnector()
	}

	return &Echo{
		connector: conn,
		config:    config,
	}, nil
}

// SocketID returns the current connection socket ID, or empty if not connected.
func (e *Echo) SocketID() string {
	return e.connector.SocketID()
}

// Channel returns a public channel subscription.
func (e *Echo) Channel(name string) Channel {
	return e.connector.Channel(name)
}

// Private returns a private channel subscription (private- prefix applied automatically).
func (e *Echo) Private(name string) Channel {
	return e.connector.PrivateChannel(name)
}

// Presence returns a presence channel subscription (presence- prefix applied automatically).
func (e *Echo) Presence(name string) PresenceChannel {
	return e.connector.PresenceChannel(name)
}

// Connect opens the WebSocket connection.
func (e *Echo) Connect() error {
	return e.connector.Connect()
}

// Disconnect closes the WebSocket connection.
func (e *Echo) Disconnect() error {
	return e.connector.Disconnect()
}

// On registers a callback for connection lifecycle events
// (e.g. "connecting", "connected", "disconnected", "error").
func (e *Echo) On(event string, callback func(data any)) {
	e.connector.On(event, callback)
}
