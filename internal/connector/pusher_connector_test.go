package connector

import "testing"

func TestUserAuthConfigPassthrough(t *testing.T) {
	endpoint := "http://localhost:8000/broadcasting/user-auth"
	headers := map[string]string{
		"Authorization": "Bearer tok",
		"X-CSRF-TOKEN":  "csrf",
	}

	c, err := NewPusherConnector(PusherConfig{
		Key:              "test-key",
		Cluster:          "mt1",
		UserAuthEndpoint: endpoint,
		UserAuthHeaders:  headers,
	})
	if err != nil {
		t.Fatal(err)
	}

	if c.options.UserAuthEndpoint != endpoint {
		t.Errorf("UserAuthEndpoint = %q, want %q", c.options.UserAuthEndpoint, endpoint)
	}
	for k, want := range headers {
		if got := c.options.UserAuthHeaders[k]; got != want {
			t.Errorf("UserAuthHeaders[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestSigninNullConnector(t *testing.T) {
	c := NewNullConnector()
	c.Signin()
}

func TestSigninPusherConnector(t *testing.T) {
	c, err := NewPusherConnector(PusherConfig{Key: "test-key", Cluster: "mt1"})
	if err != nil {
		t.Fatal(err)
	}
	c.Signin() // no-op when not connected
}
