package channel

// Channel defines the interface for a broadcasting channel.
type Channel interface {
	Listen(event string, callback func(data interface{})) Channel
	ListenForWhisper(event string, callback func(data interface{})) Channel
	Whisper(event string, data interface{}) Channel
	StopListening(event string, callback ...func(data interface{})) Channel
	StopListeningForWhisper(event string, callback ...func(data interface{})) Channel
	Subscribed(callback func()) Channel
	Error(callback func(error)) Channel
	Notification(callback func(data interface{})) Channel
	StopListeningForNotification(callback ...func(data interface{})) Channel
	ListenToAll(callback func(event string, data interface{})) Channel
	StopListeningToAll(callback ...func(event string, data interface{})) Channel
	Unsubscribe()
}

// PresenceChannel defines the interface for a presence channel.
type PresenceChannel interface {
	Channel
	Here(callback func(members interface{})) PresenceChannel
	Joining(callback func(member interface{})) PresenceChannel
	Leaving(callback func(member interface{})) PresenceChannel
}
