//go:build integration

package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/H-0-O/echo-go"
)

type reverbEnv struct {
	Host     string
	Port     int
	Key      string
	AuthURL  string
	Token    string
	AppID    string
	Secret   string
}

func requireReverb(t *testing.T) reverbEnv {
	t.Helper()
	if os.Getenv("ECHO_TEST_REVERB_KEY") == "" {
		t.Skip("ECHO_TEST_REVERB_KEY not set")
	}
	port, err := strconv.Atoi(os.Getenv("ECHO_TEST_REVERB_PORT"))
	if err != nil || port == 0 {
		t.Fatal("ECHO_TEST_REVERB_PORT must be a non-zero port")
	}
	host := os.Getenv("ECHO_TEST_REVERB_HOST")
	if host == "" {
		t.Fatal("ECHO_TEST_REVERB_HOST must be set")
	}
	authURL := os.Getenv("ECHO_TEST_AUTH_URL")
	if authURL == "" {
		t.Fatal("ECHO_TEST_AUTH_URL must be set")
	}
	token := os.Getenv("ECHO_TEST_AUTH_TOKEN")
	if token == "" {
		t.Fatal("ECHO_TEST_AUTH_TOKEN must be set")
	}
	return reverbEnv{
		Host:    host,
		Port:    port,
		Key:     os.Getenv("ECHO_TEST_REVERB_KEY"),
		AuthURL: authURL,
		Token:   token,
		AppID:   os.Getenv("ECHO_TEST_REVERB_APP_ID"),
		Secret:  os.Getenv("ECHO_TEST_REVERB_SECRET"),
	}
}

func newEchoClient(t *testing.T, env reverbEnv) *echo.Echo {
	t.Helper()
	falseVal := false
	client, err := echo.New(echo.Config{
		Broadcaster: "reverb",
		Key:         env.Key,
		Host:        env.Host,
		Port:        env.Port,
		BearerToken: env.Token,
		Auth: echo.AuthConfig{
			Endpoint: env.AuthURL,
			Headers:  map[string]string{"Accept": "application/json"},
		},
		AutoConnect: &falseVal,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect() })
	return client
}

func randomName(prefix string) string {
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), rand.Intn(1_000_000))
}

func waitFor(t *testing.T, timeout time.Duration, desc string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s after %v", desc, timeout)
}

func waitConnected(t *testing.T, client *echo.Echo) {
	t.Helper()
	waitFor(t, 20*time.Second, "connected", func() bool {
		return client.ConnectionStatus() == echo.StatusConnected
	})
}

func triggerReverbEvent(t *testing.T, env reverbEnv, channel, event string, data any) {
	t.Helper()
	if env.AppID == "" || env.Secret == "" {
		t.Skip("ECHO_TEST_REVERB_APP_ID and ECHO_TEST_REVERB_SECRET required for event trigger")
	}

	body, err := json.Marshal(map[string]any{
		"name":     event,
		"channels": []string{channel},
		"data":     data,
	})
	if err != nil {
		t.Fatal(err)
	}

	md5sum := md5.Sum(body)
	bodyMD5 := hex.EncodeToString(md5sum[:])
	query := fmt.Sprintf("auth_key=%s&auth_timestamp=%d&auth_version=1.0&body_md5=%s",
		env.Key, time.Now().Unix(), bodyMD5)
	mac := hmac.New(sha256.New, []byte(env.Secret))
	mac.Write([]byte(fmt.Sprintf("POST\n/apps/%s/events\n%s", env.AppID, query)))
	sig := hex.EncodeToString(mac.Sum(nil))
	url := fmt.Sprintf("http://%s:%d/apps/%s/events?%s&auth_signature=%s",
		env.Host, env.Port, env.AppID, query, sig)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("trigger %s on %s: %v", event, channel, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("trigger status %d: %s", resp.StatusCode, b)
	}
}
