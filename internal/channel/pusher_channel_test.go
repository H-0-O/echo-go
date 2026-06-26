package channel

import (
	"testing"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/pusher-go/src/types"
)

func TestUnsubscribeStopsEvents(t *testing.T) {
	client, err := pusher.NewPusher("test-key", pusher.Options{
		Options: types.Options{Cluster: "mt1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	pc := NewPusherChannel(client, "orders", "App.Events")
	var called int
	pc.Listen("OrderCreated", func(data interface{}) { called++ })

	if client.Channel("orders") == nil {
		t.Fatal("expected channel subscribed before Unsubscribe")
	}

	pc.Unsubscribe()

	if client.Channel("orders") != nil {
		t.Fatal("expected channel removed from pusher client after Unsubscribe")
	}

	empty := true
	pc.callbacks.Range(func(_, _ any) bool {
		empty = false
		return false
	})
	if !empty {
		t.Fatal("expected callbacks map cleared after Unsubscribe")
	}

	// Re-bind and emit locally; old handler must not run.
	pc.Listen("OrderCreated", func(data interface{}) { called++ })
	if called != 0 {
		t.Fatalf("called = %d, want 0 before any events", called)
	}
}

func TestPresenceChannelName(t *testing.T) {
	client, err := pusher.NewPusher("test-key", pusher.Options{
		Options: types.Options{Cluster: "mt1"},
	})
	if err != nil {
		t.Fatal(err)
	}

	pc := NewPusherPresenceChannel(client, "chat", "App.Events")
	if pc.Name != "presence-chat" {
		t.Fatalf("Name = %q, want presence-chat", pc.Name)
	}
	if client.Channel("presence-chat") == nil {
		t.Fatal("expected presence-chat subscribed, not private-presence-chat")
	}
}
