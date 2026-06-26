package channel

import (
	"testing"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/pusher-go/src/types"
)

func newTestPusherChannel(t *testing.T, namespace string) *PusherChannel {
	t.Helper()
	client, err := pusher.NewPusher("test-key", pusher.Options{
		Options: types.Options{Cluster: "mt1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewPusherChannel(client, "orders", namespace)
}

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

	pc.mu.Lock()
	empty := len(pc.events) == 0 && !pc.globalBound
	pc.mu.Unlock()
	if !empty {
		t.Fatal("expected listeners cleared after Unsubscribe")
	}

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

func TestListenForWhisperSkipsNamespace(t *testing.T) {
	pc := newTestPusherChannel(t, "App.Events")

	var called bool
	pc.ListenForWhisper("typing", func(data interface{}) { called = true })

	pc.emitEvent("client-typing", nil)
	if !called {
		t.Fatal("expected whisper handler on client-typing, not namespaced event")
	}

	called = false
	pc.emitEvent("App.Events.client-typing", nil)
	if called {
		t.Fatal("whisper must not bind namespaced client-typing")
	}
}

func TestNotificationEventConstant(t *testing.T) {
	pc := newTestPusherChannel(t, "App.Events")

	var got any
	pc.Notification(func(data interface{}) { got = data })

	eventName := pc.Formatter.Format(broadcastNotificationCreated)
	if eventName != "Illuminate\\Notifications\\Events\\BroadcastNotificationCreated" {
		t.Fatalf("notification bind name = %q, want Illuminate\\Notifications\\Events\\BroadcastNotificationCreated", eventName)
	}

	pc.emitEvent(eventName, "payload")
	if got != "payload" {
		t.Fatalf("notification payload = %v, want payload", got)
	}
}

func TestStopListeningSelective(t *testing.T) {
	pc := newTestPusherChannel(t, "App.Events")

	var cb1, cb2 int
	fn1 := func(data interface{}) { cb1++ }
	fn2 := func(data interface{}) { cb2++ }
	pc.Listen("OrderCreated", fn1)
	pc.Listen("OrderCreated", fn2)

	pc.emitEvent("App.Events.OrderCreated", nil)
	if cb1 != 1 || cb2 != 1 {
		t.Fatalf("cb1=%d cb2=%d, want both 1", cb1, cb2)
	}

	pc.StopListening("OrderCreated", fn2)

	pc.emitEvent("App.Events.OrderCreated", nil)
	if cb1 != 2 {
		t.Fatalf("cb1 = %d, want 2 after selective stop", cb1)
	}
	if cb2 != 1 {
		t.Fatalf("cb2 = %d, want 1 after selective stop", cb2)
	}

	pc.StopListening("OrderCreated")
	pc.emitEvent("App.Events.OrderCreated", nil)
	if cb1 != 2 {
		t.Fatalf("cb1 = %d, want unchanged after remove all", cb1)
	}
}

func TestListenToAll(t *testing.T) {
	pc := newTestPusherChannel(t, "App.Events")

	var event string
	var data any
	pc.ListenToAll(func(e string, d interface{}) {
		event = e
		data = d
	})

	pc.emitEvent("App.Events.OrderShipped", "order")
	if event != "OrderShipped" || data != "order" {
		t.Fatalf("got event=%q data=%v, want OrderShipped/order", event, data)
	}

	event = ""
	pc.emitEvent("pusher:subscription_succeeded", nil)
	if event != "" {
		t.Fatalf("pusher: events must be skipped, got %q", event)
	}
}

func TestSubscribedAndError(t *testing.T) {
	pc := newTestPusherChannel(t, "App.Events")

	var subscribed bool
	pc.Subscribed(func() { subscribed = true })
	pc.emitEvent("pusher:subscription_succeeded", nil)
	if !subscribed {
		t.Fatal("expected subscribed callback")
	}

	var errMsg string
	pc.Error(func(err error) { errMsg = err.Error() })
	pc.emitEvent("pusher:subscription_error", map[string]any{
		"type":  "AuthError",
		"error": "denied",
	})
	if errMsg != "AuthError: denied" {
		t.Fatalf("error = %q, want AuthError: denied", errMsg)
	}
}
