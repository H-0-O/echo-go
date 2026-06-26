package channel

type NullChannel struct{}

func (n *NullChannel) Listen(event string, callback func(data interface{})) Channel { return n }
func (n *NullChannel) ListenForWhisper(event string, callback func(data interface{})) Channel { return n }
func (n *NullChannel) Whisper(event string, data interface{}) Channel { return n }
func (n *NullChannel) StopListening(event string) Channel { return n }
func (n *NullChannel) StopListeningForWhisper(event string) Channel { return n }
func (n *NullChannel) Unsubscribe() {}

type NullPresenceChannel struct{ NullChannel }

func (n *NullPresenceChannel) Here(callback func(members interface{})) PresenceChannel { return n }
func (n *NullPresenceChannel) Joining(callback func(member interface{})) PresenceChannel { return n }
func (n *NullPresenceChannel) Leaving(callback func(member interface{})) PresenceChannel { return n }
