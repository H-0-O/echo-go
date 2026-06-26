package connector

import (
	"github.com/H-0-O/echo-go/internal/channel"
)

// NullConnector is used when no broadcaster is configured.
type NullConnector struct {
	channels map[string]channel.Channel
}

// NewNullConnector creates a new NullConnector.
func NewNullConnector() *NullConnector {
	return &NullConnector{
		channels: make(map[string]channel.Channel),
	}
}

func (c *NullConnector) Connect() error { return nil }

func (c *NullConnector) Disconnect() error { return nil }

func (c *NullConnector) Channel(name string) channel.Channel {
	if ch, ok := c.channels[name]; ok {
		return ch
	}
	ch := &channel.NullChannel{}
	c.channels[name] = ch
	return ch
}

func (c *NullConnector) PrivateChannel(name string) channel.Channel {
	return c.Channel(name)
}

func (c *NullConnector) PresenceChannel(name string) channel.PresenceChannel {
	return &channel.NullPresenceChannel{}
}

func (c *NullConnector) SocketID() string {
	return "null-socket-id"
}

func (c *NullConnector) On(event string, callback func(data interface{})) {
	// Do nothing
}
