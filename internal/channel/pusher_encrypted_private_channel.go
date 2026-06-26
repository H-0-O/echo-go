package channel

import (
	"strings"

	pusher "github.com/H-0-O/pusher-go"
)

type PusherEncryptedPrivateChannel struct {
	*PusherChannel
}

func NewPusherEncryptedPrivateChannel(client *pusher.Pusher, name string, namespace string) *PusherEncryptedPrivateChannel {
	name = strings.TrimPrefix(name, "private-encrypted-")
	name = "private-encrypted-" + name
	return &PusherEncryptedPrivateChannel{
		PusherChannel: NewPusherChannel(client, name, namespace),
	}
}

// Whisper sends an unencrypted client event (same as JS encryptedPrivate whisper).
func (c *PusherEncryptedPrivateChannel) Whisper(event string, data interface{}) Channel {
	return c.PusherChannel.Whisper(event, data)
}
