//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/H-0-O/echo-go"
)

func TestPublicChannelListen(t *testing.T) {
	env := requireReverb(t)
	client := newEchoClient(t, env)

	channelName := randomName("public")
	done := make(chan any, 1)
	client.Channel(channelName).Listen("TestEvent", func(data any) {
		done <- data
	})

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	waitConnected(t, client)

	triggerReverbEvent(t, env, channelName, "TestEvent", map[string]string{"msg": "hi"})
	waitFor(t, 20*time.Second, "TestEvent", func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	})
}

func TestPrivateChannelSubscribe(t *testing.T) {
	env := requireReverb(t)
	client := newEchoClient(t, env)

	subscribed := make(chan struct{}, 1)
	client.Private("test").Subscribed(func() {
		select {
		case subscribed <- struct{}{}:
		default:
		}
	})

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	waitConnected(t, client)

	waitFor(t, 20*time.Second, "private subscription", func() bool {
		select {
		case <-subscribed:
			return true
		default:
			return false
		}
	})
}

func TestPresenceHere(t *testing.T) {
	env := requireReverb(t)
	client := newEchoClient(t, env)

	here := make(chan []any, 1)
	client.Presence("room").Here(func(members any) {
		if slice, ok := members.([]any); ok {
			select {
			case here <- slice:
			default:
			}
		}
	})

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	waitConnected(t, client)

	waitFor(t, 20*time.Second, "presence here", func() bool {
		select {
		case members := <-here:
			return len(members) >= 1
		default:
			return false
		}
	})
}

func TestConnectionStatusChange(t *testing.T) {
	env := requireReverb(t)
	client := newEchoClient(t, env)

	var statuses []echo.ConnectionStatus
	client.OnConnectionChange(func(s echo.ConnectionStatus) {
		statuses = append(statuses, s)
	})

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	waitConnected(t, client)

	hasConnecting := false
	hasConnected := false
	for _, s := range statuses {
		if s == echo.StatusConnecting {
			hasConnecting = true
		}
		if s == echo.StatusConnected {
			hasConnected = true
		}
	}
	if !hasConnecting {
		t.Error("expected connecting status in OnConnectionChange callbacks")
	}
	if !hasConnected {
		t.Error("expected connected status in OnConnectionChange callbacks")
	}
}

func TestLeaveStopsEvents(t *testing.T) {
	env := requireReverb(t)
	client := newEchoClient(t, env)

	channelName := randomName("leave")
	received := make(chan struct{}, 1)
	client.Channel(channelName).Listen("LeaveEvent", func(_ any) {
		select {
		case received <- struct{}{}:
		default:
		}
	})

	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	waitConnected(t, client)

	client.Leave(channelName)
	time.Sleep(200 * time.Millisecond)

	triggerReverbEvent(t, env, channelName, "LeaveEvent", "payload")
	time.Sleep(500 * time.Millisecond)

	select {
	case <-received:
		t.Fatal("expected no events after Leave")
	default:
	}
}

func TestEncryptedPrivateSubscribe(t *testing.T) {
	t.Skip("encrypted channel integration requires Reverb encryption setup")
}
