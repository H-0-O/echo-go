package channel

import (
	"testing"

	pusher "github.com/H-0-O/pusher-go"
	pusherenc "github.com/H-0-O/pusher-go/with-encryption"
	"github.com/H-0-O/pusher-go/src/types"
)

func newTestEncryptedClient(t *testing.T) *pusher.Pusher {
	t.Helper()
	enc, err := pusherenc.New("test-key", pusher.Options{
		Options: types.Options{Cluster: "mt1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return enc.Pusher
}

func TestEncryptedPrefix(t *testing.T) {
	client := newTestEncryptedClient(t)

	for _, name := range []string{"secrets", "private-encrypted-secrets"} {
		t.Run(name, func(t *testing.T) {
			pc := NewPusherEncryptedPrivateChannel(client, name, "App.Events")
			if pc.Name != "private-encrypted-secrets" {
				t.Fatalf("Name = %q, want private-encrypted-secrets", pc.Name)
			}
			if client.Channel("private-encrypted-secrets") == nil {
				t.Fatal("expected private-encrypted-secrets subscribed")
			}
		})
	}
}

func TestEncryptedListenForWhisperSkipsNamespace(t *testing.T) {
	client := newTestEncryptedClient(t)
	pc := NewPusherEncryptedPrivateChannel(client, "secrets", "App.Events")

	var called bool
	pc.ListenForWhisper("typing", func(data interface{}) { called = true })

	pc.emitEvent("client-typing", nil)
	if !called {
		t.Fatal("ListenForWhisper must bind client-typing without namespace prefix")
	}
}
