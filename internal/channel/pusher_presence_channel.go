package channel

import (
	"strings"

	pusher "github.com/H-0-O/pusher-go"
)

type PusherPresenceChannel struct {
	*PusherChannel
}

func NewPusherPresenceChannel(client *pusher.Pusher, name string, namespace string) *PusherPresenceChannel {
	if !strings.HasPrefix(name, "presence-") {
		name = "presence-" + name
	}
	return &PusherPresenceChannel{
		PusherChannel: NewPusherChannel(client, name, namespace),
	}
}

// Here registers a callback with the current member list after subscription succeeds.
// The callback receives a []any slice of member info objects (Laravel Echo here).
func (c *PusherPresenceChannel) Here(callback func(members interface{})) PresenceChannel {
	c.addEventListener("pusher:subscription_succeeded", callback, func(data any) {
		callback(presenceMembers(data))
	})
	return c
}

// Joining registers a callback when a member joins.
// The callback receives the member info object only (Laravel Echo joining).
func (c *PusherPresenceChannel) Joining(callback func(member interface{})) PresenceChannel {
	c.addEventListener("pusher:member_added", callback, func(data any) {
		callback(memberInfo(data))
	})
	return c
}

// Leaving registers a callback when a member leaves.
// The callback receives the member info object only (Laravel Echo leaving).
func (c *PusherPresenceChannel) Leaving(callback func(member interface{})) PresenceChannel {
	c.addEventListener("pusher:member_removed", callback, func(data any) {
		callback(memberInfo(data))
	})
	return c
}

func presenceMembers(data any) []any {
	if members, ok := data.(*pusher.Members); ok {
		out := make([]any, 0)
		members.Each(func(m *pusher.Member) {
			out = append(out, m.Info)
		})
		return out
	}
	// ponytail: fallback for raw map shape if pusher-go changes
	if m, ok := data.(map[string]any); ok {
		if raw, ok := m["members"].(map[string]any); ok {
			out := make([]any, 0, len(raw))
			for _, v := range raw {
				out = append(out, v)
			}
			return out
		}
	}
	return nil
}

func memberInfo(data any) any {
	if m, ok := data.(map[string]any); ok {
		if info, ok := m["info"]; ok {
			return info
		}
	}
	return data
}
