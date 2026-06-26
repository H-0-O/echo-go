package channel

import (
	"testing"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/pusher-go/src/types"
)

func newTestPresenceChannel(t *testing.T) *PusherPresenceChannel {
	t.Helper()
	client, err := pusher.NewPusher("test-key", pusher.Options{
		Options: types.Options{Cluster: "mt1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewPusherPresenceChannel(client, "room", "App.Events")
}

func TestPresenceHerePayload(t *testing.T) {
	pc := newTestPresenceChannel(t)

	var members []any
	pc.Here(func(m interface{}) {
		members = m.([]any)
	})

	pc.emitEvent("pusher:subscription_succeeded", map[string]any{
		"members": map[string]any{
			"1": map[string]any{"name": "Alice"},
			"2": map[string]any{"name": "Bob"},
		},
	})

	if len(members) != 2 {
		t.Fatalf("len(members) = %d, want 2", len(members))
	}
	names := make(map[string]bool)
	for _, m := range members {
		info, ok := m.(map[string]any)
		if !ok {
			t.Fatalf("member info = %T, want map[string]any", m)
		}
		names[info["name"].(string)] = true
	}
	if !names["Alice"] || !names["Bob"] {
		t.Fatalf("members = %v, want Alice and Bob", members)
	}
}

func TestPresenceJoiningLeavingPayload(t *testing.T) {
	pc := newTestPresenceChannel(t)

	var joined, left any
	pc.Joining(func(m interface{}) { joined = m })
	pc.Leaving(func(m interface{}) { left = m })

	info := map[string]any{"name": "Carol"}
	pc.emitEvent("pusher:member_added", map[string]any{"id": "3", "info": info})
	pc.emitEvent("pusher:member_removed", map[string]any{"id": "3", "info": info})

	joinedInfo, ok := joined.(map[string]any)
	if !ok || joinedInfo["name"] != "Carol" {
		t.Fatalf("joining payload = %v, want info only", joined)
	}
	leftInfo, ok := left.(map[string]any)
	if !ok || leftInfo["name"] != "Carol" {
		t.Fatalf("leaving payload = %v, want info only", left)
	}
}

func TestPresenceMembersHelper(t *testing.T) {
	got := presenceMembers(map[string]any{
		"members": map[string]any{
			"1": map[string]any{"name": "Dave"},
		},
	})
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	info, ok := got[0].(map[string]any)
	if !ok || info["name"] != "Dave" {
		t.Fatalf("member = %v, want Dave", got[0])
	}

	info = memberInfo(map[string]any{"id": "1", "info": map[string]any{"name": "Eve"}}).(map[string]any)
	if info["name"] != "Eve" {
		t.Fatalf("memberInfo = %v, want Eve", info)
	}
}
