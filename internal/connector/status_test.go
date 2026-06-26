package connector_test

import (
	"testing"

	"github.com/H-0-O/echo-go/internal/connector"
)

func TestConnectionStatusMapping(t *testing.T) {
	tests := []struct {
		state string
		want  connector.ConnectionStatus
	}{
		{"connected", connector.StatusConnected},
		{"disconnected", connector.StatusDisconnected},
		{"connecting", connector.StatusConnecting},
		{"unavailable", connector.StatusReconnecting},
		{"failed", connector.StatusFailed},
		{"initialized", connector.StatusDisconnected},
		{"unknown", connector.StatusDisconnected},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := connector.MapConnectionStatus(tt.state); got != tt.want {
				t.Errorf("MapConnectionStatus(%q) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}
