package channel

import (
	"strings"

	pusher "github.com/H-0-O/pusher-go"
)

type PusherPrivateChannel struct {
	*PusherChannel
}

func NewPusherPrivateChannel(client *pusher.Pusher, name string, namespace string) *PusherPrivateChannel {
	if !strings.HasPrefix(name, "private-") {
		name = "private-" + name
	}
	return &PusherPrivateChannel{
		PusherChannel: NewPusherChannel(client, name, namespace),
	}
}
