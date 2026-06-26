package channel

import (
	"fmt"
	"strings"
	"sync"

	pusher "github.com/H-0-O/pusher-go"
	"github.com/H-0-O/echo-go/internal/util"
)

const broadcastNotificationCreated = ".Illuminate\\Notifications\\Events\\BroadcastNotificationCreated"

type bindEntry struct {
	user any
	call func(any)
}

type eventBinding struct {
	mu      sync.Mutex
	entries []bindEntry
}

func (b *eventBinding) emit(data any) {
	b.mu.Lock()
	entries := append([]bindEntry(nil), b.entries...)
	b.mu.Unlock()
	for _, e := range entries {
		e.call(data)
	}
}

type globalBindEntry struct {
	user any
	call func(event string, data any)
}

type globalBinding struct {
	mu      sync.Mutex
	entries []globalBindEntry
}

func (g *globalBinding) emit(eventName string, data any) {
	g.mu.Lock()
	entries := append([]globalBindEntry(nil), g.entries...)
	g.mu.Unlock()
	for _, e := range entries {
		e.call(eventName, data)
	}
}

// PusherChannel represents a Pusher channel.
type PusherChannel struct {
	Name      string
	Pusher    *pusher.Pusher
	Channel   *pusher.Channel
	Formatter *util.EventFormatter

	mu      sync.Mutex
	events  map[string]*eventBinding
	global  *globalBinding
	globalBound bool
}

// NewPusherChannel creates a new PusherChannel.
func NewPusherChannel(client *pusher.Pusher, name string, namespace string) *PusherChannel {
	ch, _ := client.Subscribe(name)
	return &PusherChannel{
		Name:      name,
		Pusher:    client,
		Channel:   ch,
		Formatter: util.NewEventFormatter(namespace),
		events:    make(map[string]*eventBinding),
	}
}

func sameListener(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Func values in interfaces are not ==-comparable; compare underlying code pointers.
	av, bv := fmt.Sprintf("%p", a), fmt.Sprintf("%p", b)
	return av == bv
}

func (c *PusherChannel) addEventListener(eventName string, user any, call func(any)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, ok := c.events[eventName]
	if !ok {
		b = &eventBinding{}
		c.events[eventName] = b
		c.Channel.Bind(eventName, b.emit)
	}
	b.mu.Lock()
	b.entries = append(b.entries, bindEntry{user: user, call: call})
	b.mu.Unlock()
}

func (c *PusherChannel) removeEventListener(eventName string, user any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, ok := c.events[eventName]
	if !ok {
		return
	}
	b.mu.Lock()
	if user == nil {
		b.entries = nil
	} else {
		remaining := make([]bindEntry, 0, len(b.entries))
		for _, e := range b.entries {
			if !sameListener(e.user, user) {
				remaining = append(remaining, e)
			}
		}
		b.entries = remaining
	}
	empty := len(b.entries) == 0
	b.mu.Unlock()
	if empty {
		c.Channel.Unbind(eventName, b.emit)
		delete(c.events, eventName)
	}
}

func (c *PusherChannel) addGlobalListener(user any, call func(event string, data any)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.global == nil {
		c.global = &globalBinding{}
	}
	if !c.globalBound {
		c.Channel.BindGlobal(c.global.emit)
		c.globalBound = true
	}
	c.global.mu.Lock()
	c.global.entries = append(c.global.entries, globalBindEntry{user: user, call: call})
	c.global.mu.Unlock()
}

func (c *PusherChannel) removeGlobalListener(user any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.global == nil {
		return
	}
	c.global.mu.Lock()
	if user == nil {
		c.global.entries = nil
	} else {
		remaining := make([]globalBindEntry, 0, len(c.global.entries))
		for _, e := range c.global.entries {
			if !sameListener(e.user, user) {
				remaining = append(remaining, e)
			}
		}
		c.global.entries = remaining
	}
	empty := len(c.global.entries) == 0
	c.global.mu.Unlock()
	if empty && c.globalBound {
		c.Channel.UnbindGlobal(c.global.emit)
		c.globalBound = false
		c.global = nil
	}
}

func (c *PusherChannel) clearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for eventName, b := range c.events {
		c.Channel.Unbind(eventName, b.emit)
		delete(c.events, eventName)
	}
	if c.globalBound && c.global != nil {
		c.Channel.UnbindGlobal(c.global.emit)
	}
	c.global = nil
	c.globalBound = false
}

// Listen for an event on the channel instance.
func (c *PusherChannel) Listen(event string, callback func(data interface{})) Channel {
	eventName := c.Formatter.Format(event)
	c.addEventListener(eventName, callback, func(data any) { callback(data) })
	return c
}

// ListenForWhisper listens for a whisper event on the channel instance.
// Equivalent to Laravel Echo listenForWhisper — leading "." skips namespace.
func (c *PusherChannel) ListenForWhisper(event string, callback func(data interface{})) Channel {
	return c.Listen(".client-"+event, callback)
}

// Whisper sends a whisper event to the channel instance.
func (c *PusherChannel) Whisper(event string, data interface{}) Channel {
	_, _ = c.Channel.Trigger("client-"+event, data)
	return c
}

// StopListening stops listening for an event on the channel instance.
func (c *PusherChannel) StopListening(event string, callback ...func(data interface{})) Channel {
	eventName := c.Formatter.Format(event)
	if len(callback) == 0 {
		c.removeEventListener(eventName, nil)
	} else {
		c.removeEventListener(eventName, callback[0])
	}
	return c
}

// StopListeningForWhisper stops listening for a whisper event on the channel instance.
func (c *PusherChannel) StopListeningForWhisper(event string, callback ...func(data interface{})) Channel {
	return c.StopListening(".client-"+event, callback...)
}

// Subscribed registers a callback for successful subscription (Laravel Echo subscribed).
func (c *PusherChannel) Subscribed(callback func()) Channel {
	c.addEventListener("pusher:subscription_succeeded", callback, func(_ any) { callback() })
	return c
}

// Error registers a callback for subscription errors (Laravel Echo error).
func (c *PusherChannel) Error(callback func(error)) Channel {
	c.addEventListener("pusher:subscription_error", callback, func(data any) {
		callback(subscriptionError(data))
	})
	return c
}

// Notification listens for Laravel broadcast notifications.
func (c *PusherChannel) Notification(callback func(data interface{})) Channel {
	return c.Listen(broadcastNotificationCreated, callback)
}

// StopListeningForNotification stops listening for broadcast notifications.
func (c *PusherChannel) StopListeningForNotification(callback ...func(data interface{})) Channel {
	return c.StopListening(broadcastNotificationCreated, callback...)
}

// ListenToAll listens for all non-pusher events on the channel (Laravel Echo listenToAll).
func (c *PusherChannel) ListenToAll(callback func(event string, data interface{})) Channel {
	c.addGlobalListener(callback, func(eventName string, data any) {
		if strings.HasPrefix(eventName, "pusher:") {
			return
		}
		callback(c.Formatter.StripNamespace(eventName), data)
	})
	return c
}

// StopListeningToAll stops global event listeners registered via ListenToAll.
func (c *PusherChannel) StopListeningToAll(callback ...func(event string, data interface{})) Channel {
	if len(callback) == 0 {
		c.removeGlobalListener(nil)
	} else {
		c.removeGlobalListener(callback[0])
	}
	return c
}

// Unsubscribe removes all listeners and unsubscribes from the channel.
func (c *PusherChannel) Unsubscribe() {
	c.clearAll()
	c.Pusher.Unsubscribe(c.Name)
}

// ponytail: pusher-go sends map[string]any with type/error/status on subscription failure.
func subscriptionError(data any) error {
	m, ok := data.(map[string]any)
	if !ok {
		return fmt.Errorf("subscription error: %v", data)
	}
	errStr, _ := m["error"].(string)
	typ, _ := m["type"].(string)
	if status, ok := m["status"].(float64); ok && errStr != "" {
		return fmt.Errorf("%s: %s (status %d)", typ, errStr, int(status))
	}
	if errStr != "" {
		return fmt.Errorf("%s: %s", typ, errStr)
	}
	return fmt.Errorf("subscription error: %v", data)
}

// emitEvent drives channel events in tests (Channel embeds events.Dispatcher).
func (c *PusherChannel) emitEvent(event string, data any) {
	c.Channel.Emit(event, data, pusher.Metadata{})
}
