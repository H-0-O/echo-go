package channel

import (
	"strings"

	pusher "github.com/H-0-O/pusher-go"
)

type PusherPresenceChannel struct {
	*PusherChannel
}

func NewPusherPresenceChannel(client *pusher.Pusher, name string, namespace string) *PusherPresenceChannel {
	if !strings.HasPrefix(name, "presence-") {
		name = "presence-" + name
	}
	return &PusherPresenceChannel{
		PusherChannel: NewPusherChannel(client, name, namespace),
	}
}

func (c *PusherPresenceChannel) Here(callback func(members interface{})) PresenceChannel {
	c.Channel.Bind("pusher:subscription_succeeded", func(data any) {
		callback(data)
	})
	return c
}

func (c *PusherPresenceChannel) Joining(callback func(member interface{})) PresenceChannel {
	c.Channel.Bind("pusher:member_added", func(data any) {
		callback(data)
	})
	return c
}

func (c *PusherPresenceChannel) Leaving(callback func(member interface{})) PresenceChannel {
	c.Channel.Bind("pusher:member_removed", func(data any) {
		callback(data)
	})
	return c
}
