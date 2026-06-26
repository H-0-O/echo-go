// Package echo is a Laravel Echo–style client library for Laravel Reverb and
// Pusher-compatible WebSocket broadcasters.
package echo

import (
	"fmt"

	"github.com/H-0-O/echo-go/internal/channel"
	"github.com/H-0-O/echo-go/internal/connector"
)

// Channel is a broadcast channel subscription.
type Channel = channel.Channel

// PresenceChannel is a presence broadcast channel subscription.
type PresenceChannel = channel.PresenceChannel

// Connector is the broadcasting backend interface.
// Use Config.Connector to inject a custom implementation.
type Connector = connector.Connector

// ConnectionStatus is the high-level WebSocket connection state.
type ConnectionStatus = connector.ConnectionStatus

// Connection status constants.
const (
	StatusConnected    = connector.StatusConnected
	StatusDisconnected = connector.StatusDisconnected
	StatusConnecting   = connector.StatusConnecting
	StatusReconnecting = connector.StatusReconnecting
	StatusFailed       = connector.StatusFailed
)

// AuthConfig configures channel authorization requests.
type AuthConfig struct {
	Endpoint string
	Headers  map[string]string
}

// UserAuthConfig configures Pusher user authentication requests (wired in Phase 4).
type UserAuthConfig struct {
	Endpoint string
	Headers  map[string]string
}

// Config configures a new Echo client.
type Config struct {
	Broadcaster        string
	Key                string
	Cluster            string
	Host               string
	Port               int
	TLS                bool
	Namespace          *string // nil → "App.Events"; &"" disables prefixing
	Auth               AuthConfig
	UserAuthentication UserAuthConfig
	BearerToken        string
	CSRFToken          string
	AutoConnect        *bool

	// Connector, when non-nil, is used instead of Broadcaster string resolution.
	// Broadcaster is ignored when Connector is set.
	Connector Connector

	// Deprecated: use Auth.Endpoint. Migrated in applyDefaults when Auth.Endpoint is empty.
	AuthEndpoint string
	// Deprecated: use Auth.Headers. Migrated in applyDefaults.
	AuthHeaders map[string]string
}

// Echo is the main client. Create one with [New].
type Echo struct {
	connector connector.Connector
	config    Config
}

// New creates an Echo client for the given broadcaster configuration.
//
// Use Broadcaster "null" for a built-in no-op connector (no Key or Host required).
// Set Config.Connector to inject a custom backend (Go equivalent of Laravel Echo
// broadcaster: function); Broadcaster is ignored when Connector is set.
func New(config Config) (*Echo, error) {
	config = applyDefaults(config)

	if config.Connector != nil {
		e := &Echo{
			connector: config.Connector,
			config:    config,
		}
		if autoConnect(config) {
			if err := e.Connect(); err != nil {
				return nil, err
			}
		}
		return e, nil
	}

	var conn connector.Connector

	switch config.Broadcaster {
	case "reverb", "pusher":
		pusherConn, err := newPusherConnector(config)
		if err != nil {
			return nil, err
		}
		conn = pusherConn
	case "ably":
		config.Cluster = ""
		pusherConn, err := newPusherConnector(config)
		if err != nil {
			return nil, err
		}
		conn = pusherConn
	case "null":
		conn = connector.NewNullConnector()
	default:
		return nil, fmt.Errorf("unknown broadcaster %q", config.Broadcaster)
	}

	e := &Echo{
		connector: conn,
		config:    config,
	}

	if autoConnect(config) {
		if err := e.Connect(); err != nil {
			return nil, err
		}
	}

	return e, nil
}

func newPusherConnector(config Config) (*connector.PusherConnector, error) {
	return connector.NewPusherConnector(connector.PusherConfig{
		Key:              config.Key,
		Host:             config.Host,
		Port:             config.Port,
		Cluster:          config.Cluster,
		Namespace:        resolveNamespace(config.Namespace),
		TLS:              config.TLS,
		AuthEndpoint:     config.Auth.Endpoint,
		AuthHeaders:      config.Auth.Headers,
		UserAuthEndpoint: config.UserAuthentication.Endpoint,
		UserAuthHeaders:  config.UserAuthentication.Headers,
	})
}

func resolveNamespace(ns *string) string {
	if ns == nil {
		return "App.Events"
	}
	return *ns
}

func autoConnect(c Config) bool {
	return c.AutoConnect == nil || *c.AutoConnect
}

func applyDefaults(c Config) Config {
	if c.Auth.Endpoint == "" && c.AuthEndpoint != "" {
		c.Auth.Endpoint = c.AuthEndpoint
	}
	if c.Auth.Endpoint == "" {
		c.Auth.Endpoint = "/broadcasting/auth"
	}
	if c.UserAuthentication.Endpoint == "" {
		c.UserAuthentication.Endpoint = "/broadcasting/user-auth"
	}

	if c.Namespace == nil {
		defaultNS := "App.Events"
		c.Namespace = &defaultNS
	}

	c.Auth.Headers = mergeAuthHeaders(c.BearerToken, c.CSRFToken, c.Auth.Headers, c.AuthHeaders)
	c.UserAuthentication.Headers = mergeAuthHeaders(c.BearerToken, c.CSRFToken, c.UserAuthentication.Headers, nil)

	return c
}

func mergeAuthHeaders(bearer, csrf string, headers, deprecated map[string]string) map[string]string {
	out := make(map[string]string)
	if bearer != "" {
		out["Authorization"] = "Bearer " + bearer
	}
	if csrf != "" {
		out["X-CSRF-TOKEN"] = csrf
	}
	for k, v := range headers {
		out[k] = v
	}
	for k, v := range deprecated {
		out[k] = v
	}
	return out
}

// SocketID returns the current connection socket ID, or empty if not connected.
func (e *Echo) SocketID() string {
	return e.connector.SocketID()
}

// ConnectionStatus returns the current mapped connection status.
func (e *Echo) ConnectionStatus() ConnectionStatus {
	return e.connector.ConnectionStatus()
}

// OnConnectionChange registers a callback for connection status transitions.
// The returned function unsubscribes the callback.
func (e *Echo) OnConnectionChange(cb func(ConnectionStatus)) func() {
	return e.connector.OnConnectionChange(cb)
}

// Channel returns a public channel subscription.
func (e *Echo) Channel(name string) Channel {
	return e.connector.Channel(name)
}

// Private returns a private channel subscription (private- prefix applied automatically).
func (e *Echo) Private(name string) Channel {
	return e.connector.PrivateChannel(name)
}

// EncryptedPrivate returns an encrypted private channel subscription (private-encrypted- prefix applied automatically).
func (e *Echo) EncryptedPrivate(name string) Channel {
	return e.connector.EncryptedPrivateChannel(name)
}

// Presence returns a presence channel subscription (presence- prefix applied automatically).
func (e *Echo) Presence(name string) PresenceChannel {
	return e.connector.PresenceChannel(name)
}

// Join is an alias for [Echo.Presence].
func (e *Echo) Join(name string) PresenceChannel {
	return e.Presence(name)
}

// Leave unsubscribes public, private, encrypted-private, and presence variants of a channel.
func (e *Echo) Leave(channel string) {
	e.connector.Leave(channel)
}

// LeaveChannel unsubscribes a single channel by exact name (already prefixed if private/presence).
func (e *Echo) LeaveChannel(name string) {
	e.connector.LeaveChannel(name)
}

// LeaveAllChannels unsubscribes every channel without disconnecting the WebSocket.
func (e *Echo) LeaveAllChannels() {
	e.connector.LeaveAllChannels()
}

// Listen subscribes to a public channel event (shorthand for Channel(name).Listen).
func (e *Echo) Listen(channelName, event string, callback func(data any)) Channel {
	return e.Channel(channelName).Listen(event, callback)
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

// Signin triggers Pusher user authentication (POST to UserAuthentication.Endpoint).
// Fire-and-forget; call after Connect when connected (or on a "connected" callback).
// Success/failure is handled inside pusher-go (pusher:signin / pusher:signin_success).
func (e *Echo) Signin() {
	e.connector.Signin()
}
