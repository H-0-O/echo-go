package echo

import (
	"testing"

	"github.com/H-0-O/echo-go/internal/connector"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{
		BearerToken: "tok",
		CSRFToken:   "csrf",
		Auth: AuthConfig{
			Headers: map[string]string{"X-Custom": "1"},
		},
	}
	got := applyDefaults(cfg)

	if got.Auth.Endpoint != "/broadcasting/auth" {
		t.Errorf("Auth.Endpoint = %q, want /broadcasting/auth", got.Auth.Endpoint)
	}
	if got.UserAuthentication.Endpoint != "/broadcasting/user-auth" {
		t.Errorf("UserAuthentication.Endpoint = %q, want /broadcasting/user-auth", got.UserAuthentication.Endpoint)
	}
	if got.Namespace == nil || *got.Namespace != "App.Events" {
		t.Errorf("Namespace = %v, want App.Events", got.Namespace)
	}
	if got.Auth.Headers["Authorization"] != "Bearer tok" {
		t.Errorf("Authorization = %q, want Bearer tok", got.Auth.Headers["Authorization"])
	}
	if got.Auth.Headers["X-CSRF-TOKEN"] != "csrf" {
		t.Errorf("X-CSRF-TOKEN = %q, want csrf", got.Auth.Headers["X-CSRF-TOKEN"])
	}
	if got.Auth.Headers["X-Custom"] != "1" {
		t.Errorf("X-Custom = %q, want 1 (caller override)", got.Auth.Headers["X-Custom"])
	}
}

func TestBearerCSRFMerge(t *testing.T) {
	cfg := Config{
		BearerToken: "abc",
		CSRFToken:   "xyz",
		Auth: AuthConfig{
			Headers: map[string]string{"Authorization": "Bearer override"},
		},
	}
	got := applyDefaults(cfg)

	if got.Auth.Headers["Authorization"] != "Bearer override" {
		t.Errorf("Auth Authorization = %q, want caller override", got.Auth.Headers["Authorization"])
	}
	if got.UserAuthentication.Headers["Authorization"] != "Bearer abc" {
		t.Errorf("UserAuth Authorization = %q, want Bearer abc", got.UserAuthentication.Headers["Authorization"])
	}
	if got.Auth.Headers["X-CSRF-TOKEN"] != "xyz" {
		t.Errorf("Auth X-CSRF-TOKEN = %q, want xyz", got.Auth.Headers["X-CSRF-TOKEN"])
	}
	if got.UserAuthentication.Headers["X-CSRF-TOKEN"] != "xyz" {
		t.Errorf("UserAuth X-CSRF-TOKEN = %q, want xyz", got.UserAuthentication.Headers["X-CSRF-TOKEN"])
	}
}

func TestDeprecatedAuthFields(t *testing.T) {
	cfg := Config{
		AuthEndpoint: "http://example.com/auth",
		AuthHeaders:  map[string]string{"Cookie": "session=1"},
	}
	got := applyDefaults(cfg)

	if got.Auth.Endpoint != "http://example.com/auth" {
		t.Errorf("Auth.Endpoint = %q, want deprecated migration", got.Auth.Endpoint)
	}
	if got.Auth.Headers["Cookie"] != "session=1" {
		t.Errorf("Auth.Headers Cookie = %q, want session=1", got.Auth.Headers["Cookie"])
	}
}

func TestNamespaceDisabled(t *testing.T) {
	empty := ""
	cfg := Config{Namespace: &empty}
	got := applyDefaults(cfg)
	if got.Namespace == nil || *got.Namespace != "" {
		t.Errorf("Namespace = %v, want empty string", got.Namespace)
	}
}

func TestUnknownBroadcaster(t *testing.T) {
	falseVal := false
	_, err := New(Config{
		Broadcaster: "socket.io",
		AutoConnect: &falseVal,
	})
	if err == nil {
		t.Fatal("expected error for unknown broadcaster")
	}
}

func TestNullBroadcasterConnectionStatus(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.ConnectionStatus() != StatusDisconnected {
		t.Errorf("ConnectionStatus = %q, want disconnected before Connect", e.ConnectionStatus())
	}
	if err := e.Connect(); err != nil {
		t.Fatal(err)
	}
	if e.ConnectionStatus() != StatusConnected {
		t.Errorf("ConnectionStatus = %q, want connected after Connect", e.ConnectionStatus())
	}
	if err := e.Disconnect(); err != nil {
		t.Fatal(err)
	}
	if e.ConnectionStatus() != StatusDisconnected {
		t.Errorf("ConnectionStatus = %q, want disconnected after Disconnect", e.ConnectionStatus())
	}
}

func TestAutoConnectFalse(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if e == nil {
		t.Fatal("expected client")
	}
}

func TestOnConnectionChangeNullConnector(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	unsub := e.OnConnectionChange(func(ConnectionStatus) {})
	unsub()
}

func TestOnConnectionChangeUnsubscribe(t *testing.T) {
	c, err := connector.NewPusherConnector(connector.PusherConfig{
		Key:     "test-key",
		Cluster: "mt1",
	})
	if err != nil {
		t.Fatal(err)
	}

	var statuses []connector.ConnectionStatus
	unsub := c.OnConnectionChange(func(s connector.ConnectionStatus) {
		statuses = append(statuses, s)
	})

	c.Connect()
	if len(statuses) == 0 {
		t.Fatal("expected at least one status callback on Connect")
	}

	unsub()
	n := len(statuses)
	c.Disconnect()
	if len(statuses) != n {
		t.Errorf("got %d callbacks after unsubscribe, want %d", len(statuses), n)
	}
}

func TestJoinAlias(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.Join("room") != e.Presence("room") {
		t.Fatal("Join must return same instance as Presence")
	}
}

func TestListenShorthand(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	ch := e.Listen("orders", "OrderCreated", func(data any) {})
	if ch != e.Channel("orders") {
		t.Fatal("Listen must return the public channel instance")
	}
}

func TestEchoLeave(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	e.Channel("x")
	e.Private("x")
	e.EncryptedPrivate("x")
	e.Presence("x")
	e.Leave("x")
	e.LeaveChannel("private-y")
	e.LeaveAllChannels()
}

func TestEncryptedPrivateNullBroadcaster(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	first := e.EncryptedPrivate("secrets")
	second := e.EncryptedPrivate("secrets")
	if first != second {
		t.Fatal("EncryptedPrivate must return cached instance")
	}
	_ = first.Listen("Event", func(data any) {})
}

func TestSigninNullBroadcaster(t *testing.T) {
	falseVal := false
	e, err := New(Config{
		Broadcaster: "null",
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	e.Signin()
}

func TestNullBroadcasterNoKey(t *testing.T) {
	e, err := New(Config{Broadcaster: "null"})
	if err != nil {
		t.Fatal(err)
	}
	if e == nil {
		t.Fatal("expected client")
	}
	if e.ConnectionStatus() != StatusConnected {
		t.Errorf("ConnectionStatus = %q, want connected (AutoConnect default)", e.ConnectionStatus())
	}
}

// TestNullBroadcasterCI is the CI smoke test for the null broadcaster (no WebSocket server).
func TestNullBroadcasterCI(t *testing.T) {
	TestNullBroadcasterNoKey(t)
}

type stubConnector struct {
	*connector.NullConnector
	connectCalls int
	channelCalls int
}

func (s *stubConnector) Connect() error {
	s.connectCalls++
	return s.NullConnector.Connect()
}

func (s *stubConnector) Channel(name string) Channel {
	s.channelCalls++
	return s.NullConnector.Channel(name)
}

func TestCustomConnectorInjection(t *testing.T) {
	falseVal := false
	stub := &stubConnector{NullConnector: connector.NewNullConnector()}
	e, err := New(Config{
		Connector:   stub,
		Broadcaster: "socket.io", // ignored when Connector is set
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Connect(); err != nil {
		t.Fatal(err)
	}
	if stub.connectCalls != 1 {
		t.Fatalf("Connect calls = %d, want 1", stub.connectCalls)
	}
	e.Channel("orders")
	if stub.channelCalls != 1 {
		t.Fatalf("Channel calls = %d, want 1", stub.channelCalls)
	}
	e.Signin()
	e.Leave("orders")
}
