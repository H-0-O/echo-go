package util

import "testing"

func TestEventFormatterFormat(t *testing.T) {
	tests := []struct {
		event     string
		namespace string
		want      string
	}{
		{".client-typing", "App.Events", "client-typing"},
		{"OrderCreated", "App.Events", "App.Events.OrderCreated"},
		{".Illuminate\\Notifications\\Events\\BroadcastNotificationCreated", "App.Events", "Illuminate\\Notifications\\Events\\BroadcastNotificationCreated"},
		{"\\BackslashEvent", "App.Events", "BackslashEvent"},
		{"PlainEvent", "", "PlainEvent"},
	}

	for _, tt := range tests {
		f := NewEventFormatter(tt.namespace)
		got := f.Format(tt.event)
		if got != tt.want {
			t.Errorf("Format(%q, namespace=%q) = %q, want %q", tt.event, tt.namespace, got, tt.want)
		}
	}
}

func TestEventFormatterStripNamespace(t *testing.T) {
	tests := []struct {
		event     string
		namespace string
		want      string
	}{
		{"App.Events.OrderShipped", "App.Events", "OrderShipped"},
		{"OrderShipped", "App.Events", "OrderShipped"},
		{"App.Events.OrderShipped", "", "App.Events.OrderShipped"},
	}

	for _, tt := range tests {
		f := NewEventFormatter(tt.namespace)
		got := f.StripNamespace(tt.event)
		if got != tt.want {
			t.Errorf("StripNamespace(%q, namespace=%q) = %q, want %q", tt.event, tt.namespace, got, tt.want)
		}
	}
}
