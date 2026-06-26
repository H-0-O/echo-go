package connector

import (
	"sync"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/pusher-go/src/types"
	"github.com/H-0-O/echo-go/internal/channel"
)

// PusherConfig holds configuration specifically for Pusher/Reverb.
type PusherConfig struct {
	Key          string
	Host         string
	Port         int
	Namespace    string
	AuthEndpoint string
	AuthHeaders  map[string]string
	TLS          bool
}

// PusherConnector handles connection to Pusher-compatible servers (like Reverb).
type PusherConnector struct {
	client   *pusher.Pusher
	options  PusherConfig
	channels map[string]channel.Channel
	mu       sync.RWMutex
}

// NewPusherConnector creates a new PusherConnector.
func NewPusherConnector(config PusherConfig) (*PusherConnector, error) {
	client, err := pusher.NewPusher(config.Key, pusher.Options{
		Options: types.Options{
			Cluster:  "mt1", // ponytail: required by ValidateOptions; unused when WsHost set
			WsHost:   config.Host,
			WsPort:   config.Port,
			ForceTLS: types.BoolPtr(config.TLS),
			ChannelAuthorization: types.ChannelAuthorizationConfig{
				Endpoint:  config.AuthEndpoint,
				Transport: "ajax",
				Headers:   config.AuthHeaders,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &PusherConnector{
		client:   client,
		options:  config,
		channels: make(map[string]channel.Channel),
	}, nil
}

// Connect to the service.
func (c *PusherConnector) Connect() error {
	c.client.Connect()
	return nil
}

// Disconnect from the service.
func (c *PusherConnector) Disconnect() error {
	c.client.Disconnect()
	return nil
}

// Channel returns a public channel.
func (c *PusherConnector) Channel(name string) channel.Channel {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.channels[name]; ok {
		return ch
	}

	ch := channel.NewPusherChannel(c.client, name, c.options.Namespace)
	c.channels[name] = ch
	return ch
}

// PrivateChannel returns a private channel.
func (c *PusherConnector) PrivateChannel(name string) channel.Channel {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.channels[name]; ok {
		return ch
	}

	ch := channel.NewPusherPrivateChannel(c.client, name, c.options.Namespace)
	c.channels[name] = ch
	return ch
}

// PresenceChannel returns a presence channel.
func (c *PusherConnector) PresenceChannel(name string) channel.PresenceChannel {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.channels[name]; ok {
		return ch.(channel.PresenceChannel)
	}

	ch := channel.NewPusherPresenceChannel(c.client, name, c.options.Namespace)
	c.channels[name] = ch
	return ch
}

// SocketID returns the connection socket ID.
func (c *PusherConnector) SocketID() string {
	return c.client.SocketID()
}

// On registers a callback for connection events.
func (c *PusherConnector) On(event string, callback func(data interface{})) {
	c.client.Connection.Bind(event, func(data any) {
		callback(data)
	})
}
