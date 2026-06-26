package connector

import (
	"sync"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/pusher-go/src/types"
	"github.com/H-0-O/echo-go/internal/channel"
)

// PusherConfig holds configuration specifically for Pusher/Reverb.
type PusherConfig struct {
	Key              string
	Host             string
	Port             int
	Cluster          string
	Namespace        string
	TLS              bool
	AuthEndpoint     string
	AuthHeaders      map[string]string
	UserAuthEndpoint string
	UserAuthHeaders  map[string]string
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
	cluster := config.Cluster
	if cluster == "" {
		cluster = "mt1" // ponytail: pusher-go ValidateOptions requires cluster; ignored when WsHost set
	}

	client, err := pusher.NewPusher(config.Key, pusher.Options{
		Options: types.Options{
			Cluster:  cluster,
			WsHost:   config.Host,
			WsPort:   config.Port,
			ForceTLS: types.BoolPtr(config.TLS),
			ChannelAuthorization: types.ChannelAuthorizationConfig{
				Endpoint:  config.AuthEndpoint,
				Transport: "ajax",
				Headers:   config.AuthHeaders,
			},
			UserAuthentication: types.UserAuthenticationConfig{
				Endpoint:  config.UserAuthEndpoint,
				Transport: "ajax",
				Headers:   config.UserAuthHeaders,
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

// ConnectionStatus returns the current mapped connection status.
func (c *PusherConnector) ConnectionStatus() ConnectionStatus {
	return MapConnectionStatus(c.client.ConnectionState())
}

// OnConnectionChange registers a callback for connection status transitions.
func (c *PusherConnector) OnConnectionChange(cb func(ConnectionStatus)) func() {
	updateStatus := func(_ any) {
		cb(c.ConnectionStatus())
	}
	events := []string{"state_change", "connected", "disconnected"}
	for _, event := range events {
		c.client.Connection.Bind(event, updateStatus)
	}
	return func() {
		for _, event := range events {
			c.client.Connection.Unbind(event, updateStatus)
		}
	}
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
	key := "private-" + name
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.channels[key]; ok {
		return ch
	}

	ch := channel.NewPusherPrivateChannel(c.client, name, c.options.Namespace)
	c.channels[key] = ch
	return ch
}

// PresenceChannel returns a presence channel.
func (c *PusherConnector) PresenceChannel(name string) channel.PresenceChannel {
	key := "presence-" + name
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, ok := c.channels[key]; ok {
		return ch.(channel.PresenceChannel)
	}

	ch := channel.NewPusherPresenceChannel(c.client, name, c.options.Namespace)
	c.channels[key] = ch
	return ch
}

// Leave unsubscribes all subscription types for a logical channel name.
func (c *PusherConnector) Leave(channel string) {
	for _, name := range []string{
		channel,
		"private-" + channel,
		"private-encrypted-" + channel,
		"presence-" + channel,
	} {
		c.LeaveChannel(name)
	}
}

// LeaveChannel unsubscribes a single channel by exact registry name.
func (c *PusherConnector) LeaveChannel(name string) {
	c.mu.Lock()
	ch, ok := c.channels[name]
	if ok {
		delete(c.channels, name)
	}
	c.mu.Unlock()
	if !ok {
		return
	}
	ch.Unsubscribe()
}

// LeaveAllChannels unsubscribes every channel and clears the registry.
func (c *PusherConnector) LeaveAllChannels() {
	c.mu.Lock()
	snapshot := c.channels
	c.channels = make(map[string]channel.Channel)
	c.mu.Unlock()
	for _, ch := range snapshot {
		ch.Unsubscribe()
	}
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
