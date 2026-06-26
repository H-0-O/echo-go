package channel

import (
	"sync"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/echo-go/internal/util"
)

// PusherChannel represents a Pusher channel.
type PusherChannel struct {
	Name      string
	Pusher    *pusher.Pusher
	Channel   *pusher.Channel
	Formatter *util.EventFormatter
	callbacks sync.Map // event name -> pusher.Callback
}

// NewPusherChannel creates a new PusherChannel.
func NewPusherChannel(client *pusher.Pusher, name string, namespace string) *PusherChannel {
	ch, _ := client.Subscribe(name)
	return &PusherChannel{
		Name:      name,
		Pusher:    client,
		Channel:   ch,
		Formatter: util.NewEventFormatter(namespace),
	}
}

// Listen for an event on the channel instance.
func (c *PusherChannel) Listen(event string, callback func(data interface{})) Channel {
	eventName := c.Formatter.Format(event)
	wrapped := func(data any) {
		callback(data)
	}
	c.callbacks.Store(eventName, wrapped)
	c.Channel.Bind(eventName, wrapped)
	return c
}

// ListenForWhisper listens for a whisper event on the channel instance.
func (c *PusherChannel) ListenForWhisper(event string, callback func(data interface{})) Channel {
	return c.Listen("client-"+event, callback)
}

// Whisper sends a whisper event to the channel instance.
func (c *PusherChannel) Whisper(event string, data interface{}) Channel {
	_, _ = c.Channel.Trigger("client-"+event, data)
	return c
}

// StopListening stops listening for an event on the channel instance.
func (c *PusherChannel) StopListening(event string) Channel {
	eventName := c.Formatter.Format(event)
	if v, ok := c.callbacks.LoadAndDelete(eventName); ok {
		c.Channel.Unbind(eventName, v.(pusher.Callback))
	}
	return c
}

// StopListeningForWhisper stops listening for a whisper event on the channel instance.
func (c *PusherChannel) StopListeningForWhisper(event string) Channel {
	return c.StopListening("client-" + event)
}
