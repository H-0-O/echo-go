package channel

// Channel defines the interface for a broadcasting channel.
type Channel interface {
	Listen(event string, callback func(data interface{})) Channel
	ListenForWhisper(event string, callback func(data interface{})) Channel
	Whisper(event string, data interface{}) Channel
	StopListening(event string) Channel
	StopListeningForWhisper(event string) Channel
}

// PresenceChannel defines the interface for a presence channel.
type PresenceChannel interface {
	Channel
	Here(callback func(members interface{})) PresenceChannel
	Joining(callback func(member interface{})) PresenceChannel
	Leaving(callback func(member interface{})) PresenceChannel
}
