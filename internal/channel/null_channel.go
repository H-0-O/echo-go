package channel

var (
	_ Channel         = (*NullChannel)(nil)
	_ Channel         = (*NullPrivateChannel)(nil)
	_ Channel         = (*NullEncryptedPrivateChannel)(nil)
	_ PresenceChannel = (*NullPresenceChannel)(nil)
)

type NullChannel struct{}

func (n *NullChannel) Listen(event string, callback func(data interface{})) Channel { return n }
func (n *NullChannel) ListenForWhisper(event string, callback func(data interface{})) Channel {
	return n
}
func (n *NullChannel) Whisper(event string, data interface{}) Channel { return n }
func (n *NullChannel) StopListening(event string, callback ...func(data interface{})) Channel {
	return n
}
func (n *NullChannel) StopListeningForWhisper(event string, callback ...func(data interface{})) Channel {
	return n
}
func (n *NullChannel) Subscribed(callback func()) Channel { return n }
func (n *NullChannel) Error(callback func(error)) Channel { return n }
func (n *NullChannel) Notification(callback func(data interface{})) Channel { return n }
func (n *NullChannel) StopListeningForNotification(callback ...func(data interface{})) Channel {
	return n
}
func (n *NullChannel) ListenToAll(callback func(event string, data interface{})) Channel { return n }
func (n *NullChannel) StopListeningToAll(callback ...func(event string, data interface{})) Channel {
	return n
}
func (n *NullChannel) Unsubscribe() {}

type NullPrivateChannel struct{ NullChannel }

type NullEncryptedPrivateChannel struct{ NullChannel }

type NullPresenceChannel struct{ NullChannel }

func (n *NullPresenceChannel) Here(callback func(members interface{})) PresenceChannel { return n }
func (n *NullPresenceChannel) Joining(callback func(member interface{})) PresenceChannel { return n }
func (n *NullPresenceChannel) Leaving(callback func(member interface{})) PresenceChannel { return n }
