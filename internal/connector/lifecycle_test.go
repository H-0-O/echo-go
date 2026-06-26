package connector

import (
	"testing"
)

func nullRegistryLen(c *NullConnector) int {
	return len(c.channels)
}

func testConnectors(t *testing.T) []struct {
	name string
	c    Connector
	null *NullConnector
} {
	t.Helper()

	null := NewNullConnector()
	pusher, err := NewPusherConnector(PusherConfig{Key: "test-key", Cluster: "mt1"})
	if err != nil {
		t.Fatal(err)
	}

	return []struct {
		name string
		c    Connector
		null *NullConnector
	}{
		{"NullConnector", null, null},
		{"PusherConnector", pusher, nil},
	}
}

func TestChannelRegistryKeys(t *testing.T) {
	for _, tc := range testConnectors(t) {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.c
			pub := c.Channel("a")
			priv := c.PrivateChannel("a")
			pres := c.PresenceChannel("a")

			if tc.null != nil {
				if nullRegistryLen(tc.null) != 3 {
					t.Fatalf("registry len = %d, want 3 distinct keys", nullRegistryLen(tc.null))
				}
			} else {
				if pub == priv {
					t.Fatal("Channel and PrivateChannel must be distinct")
				}
				if pub == pres {
					t.Fatal("Channel and PresenceChannel must be distinct")
				}
				if priv == pres {
					t.Fatal("PrivateChannel and PresenceChannel must be distinct")
				}
			}
			if c.PrivateChannel("a") != priv {
				t.Fatal("second PrivateChannel call must return cached instance")
			}
			if c.PresenceChannel("a") != pres {
				t.Fatal("second PresenceChannel call must return cached instance")
			}
		})
	}
}

func TestLeaveAllVariants(t *testing.T) {
	for _, tc := range testConnectors(t) {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.c
			pub := c.Channel("x")
			priv := c.PrivateChannel("x")
			pres := c.PresenceChannel("x")

			c.Leave("x")

			if tc.null != nil {
				if nullRegistryLen(tc.null) != 0 {
					t.Fatalf("registry len = %d, want 0 after Leave", nullRegistryLen(tc.null))
				}
				return
			}
			if c.Channel("x") == pub {
				t.Fatal("expected new public channel after Leave")
			}
			if c.PrivateChannel("x") == priv {
				t.Fatal("expected new private channel after Leave")
			}
			if c.PresenceChannel("x") == pres {
				t.Fatal("expected new presence channel after Leave")
			}
		})
	}
}

func TestLeaveChannelExact(t *testing.T) {
	for _, tc := range testConnectors(t) {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.c
			pub := c.Channel("orders")
			priv := c.PrivateChannel("orders")

			c.LeaveChannel("private-orders")

			if tc.null != nil {
				if nullRegistryLen(tc.null) != 1 {
					t.Fatalf("registry len = %d, want 1 (public only)", nullRegistryLen(tc.null))
				}
				return
			}
			if c.PrivateChannel("orders") == priv {
				t.Fatal("expected new private channel after LeaveChannel")
			}
			if c.Channel("orders") != pub {
				t.Fatal("public channel should remain cached")
			}
		})
	}
}

func TestLeaveAllChannels(t *testing.T) {
	for _, tc := range testConnectors(t) {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.c
			pub := c.Channel("foo")
			c.PrivateChannel("bar")
			c.PresenceChannel("room")

			c.LeaveAllChannels()

			if tc.null != nil {
				if nullRegistryLen(tc.null) != 0 {
					t.Fatalf("registry len = %d, want 0", nullRegistryLen(tc.null))
				}
				if c.ConnectionStatus() != StatusConnected {
					t.Fatalf("ConnectionStatus = %q, want connected", c.ConnectionStatus())
				}
				return
			}
			if c.Channel("foo") == pub {
				t.Fatal("expected new channel instance after LeaveAllChannels")
			}
		})
	}
}

func TestLeaveIdempotent(t *testing.T) {
	c := NewNullConnector()
	c.Leave("never-subscribed")
	c.LeaveChannel("private-nope")
}
