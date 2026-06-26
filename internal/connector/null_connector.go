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

func (c *NullConnector) ConnectionStatus() ConnectionStatus {
	return StatusConnected
}

func (c *NullConnector) OnConnectionChange(_ func(ConnectionStatus)) func() {
	return func() {}
}

func (c *NullConnector) Channel(name string) channel.Channel {
	if ch, ok := c.channels[name]; ok {
		return ch
	}
	ch := &channel.NullChannel{}
	c.channels[name] = ch
	return ch
}

func (c *NullConnector) PrivateChannel(name string) channel.Channel {
	key := "private-" + name
	if ch, ok := c.channels[key]; ok {
		return ch
	}
	ch := &channel.NullChannel{}
	c.channels[key] = ch
	return ch
}

func (c *NullConnector) PresenceChannel(name string) channel.PresenceChannel {
	key := "presence-" + name
	if ch, ok := c.channels[key]; ok {
		return ch.(channel.PresenceChannel)
	}
	ch := &channel.NullPresenceChannel{}
	c.channels[key] = ch
	return ch
}

// Leave unsubscribes all subscription types for a logical channel name.
func (c *NullConnector) Leave(channel string) {
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
func (c *NullConnector) LeaveChannel(name string) {
	ch, ok := c.channels[name]
	if !ok {
		return
	}
	delete(c.channels, name)
	ch.Unsubscribe()
}

// LeaveAllChannels unsubscribes every channel and clears the registry.
func (c *NullConnector) LeaveAllChannels() {
	snapshot := c.channels
	c.channels = make(map[string]channel.Channel)
	for _, ch := range snapshot {
		ch.Unsubscribe()
	}
}

func (c *NullConnector) SocketID() string {
	return "null-socket-id"
}

func (c *NullConnector) On(event string, callback func(data interface{})) {
	// Do nothing
}
