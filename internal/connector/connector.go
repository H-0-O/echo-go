package connector

import (
	"github.com/H-0-O/echo-go/internal/channel"
)

// Connector defines the interface for different broadcasting backends.
type Connector interface {
	// Connect establishes the connection to the broadcasting service.
	Connect() error

	// Disconnect closes the connection to the broadcasting service.
	Disconnect() error

	// Channel returns a public channel instance.
	Channel(name string) channel.Channel

	// PrivateChannel returns a private channel instance.
	PrivateChannel(name string) channel.Channel

	// PresenceChannel returns a presence channel instance.
	PresenceChannel(name string) channel.PresenceChannel

	// SocketID returns the socket ID for the current connection.
	SocketID() string

	// ConnectionStatus returns the current mapped connection status.
	ConnectionStatus() ConnectionStatus

	// OnConnectionChange registers a callback for connection status transitions.
	// The returned function unsubscribes the callback.
	OnConnectionChange(cb func(ConnectionStatus)) func()

	// On registers a callback for a connection event.
	// Common events: "connecting", "connected", "disconnected", "error"
	On(event string, callback func(data interface{}))
}
