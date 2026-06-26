package channel

import "testing"

func TestNullChannelChaining(t *testing.T) {
	n := &NullChannel{}
	if got := n.Listen("x", func(data interface{}) {}).
		ListenForWhisper("y", func(data interface{}) {}).
		Whisper("z", nil).
		StopListening("x").
		StopListeningForWhisper("y").
		Subscribed(func() {}).
		Error(func(error) {}).
		Notification(func(data interface{}) {}).
		StopListeningForNotification().
		ListenToAll(func(event string, data interface{}) {}).
		StopListeningToAll(); got != n {
		t.Fatal("NullChannel methods must return receiver for chaining")
	}
	n.Unsubscribe()
}

func TestNullPrivateChannelChaining(t *testing.T) {
	n := &NullPrivateChannel{}
	if got := n.Listen("x", func(data interface{}) {}); got != &n.NullChannel {
		t.Fatal("NullPrivateChannel Listen must return embedded NullChannel")
	}
	if got := n.Whisper("z", nil); got != &n.NullChannel {
		t.Fatal("NullPrivateChannel Whisper must return embedded NullChannel")
	}
}

func TestNullPresenceChannelChaining(t *testing.T) {
	n := &NullPresenceChannel{}
	if got := n.Here(func(members interface{}) {}).
		Joining(func(member interface{}) {}).
		Leaving(func(member interface{}) {}); got != n {
		t.Fatal("NullPresenceChannel presence methods must return receiver for chaining")
	}
	if got := n.Listen("x", func(data interface{}) {}); got != &n.NullChannel {
		t.Fatal("NullPresenceChannel Listen must return embedded NullChannel")
	}
}

func TestNullEncryptedPrivateChannelChaining(t *testing.T) {
	n := &NullEncryptedPrivateChannel{}
	if got := n.Listen("x", func(data interface{}) {}); got != &n.NullChannel {
		t.Fatal("NullEncryptedPrivateChannel Listen must return embedded NullChannel")
	}
	if got := n.Whisper("z", nil); got != &n.NullChannel {
		t.Fatal("NullEncryptedPrivateChannel Whisper must return embedded NullChannel")
	}
}
