package connector

import "testing"

func TestNullConnectionStatusLifecycle(t *testing.T) {
	c := NewNullConnector()
	if c.ConnectionStatus() != StatusDisconnected {
		t.Fatalf("ConnectionStatus = %q, want disconnected initially", c.ConnectionStatus())
	}
	if err := c.Connect(); err != nil {
		t.Fatal(err)
	}
	if c.ConnectionStatus() != StatusConnected {
		t.Fatalf("ConnectionStatus = %q, want connected after Connect", c.ConnectionStatus())
	}
	if err := c.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if c.ConnectionStatus() != StatusDisconnected {
		t.Fatalf("ConnectionStatus = %q, want disconnected after Disconnect", c.ConnectionStatus())
	}
}
