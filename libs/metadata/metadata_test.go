package metadata

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/bafbi/stellaroot/libs/constant"
)

// startEmbeddedNATSServer attempts to start an in-process NATS server using the nats-server binary
// present in PATH. If unavailable, tests relying on Client will be skipped.
func startEmbeddedNATSServer(t *testing.T) (url string, shutdown func()) {
	// For simplicity we rely on an external nats-server already running on default port.
	// If it's not reachable we'll skip integration-like tests.
	url = "nats://127.0.0.1:4222"
	conn, err := nats.Connect(url, nats.Timeout(500*time.Millisecond))
	if err != nil {
		// Not available
		t.Skipf("skipping client integration tests: can't reach local nats-server: %v", err)
	}
	conn.Close()
	return url, func() {}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestMetadataBasicLabelAnnotationOps(t *testing.T) {
	m := &Metadata{}
	if m.HasLabel("env", "prod") {
		// Should be false initially
		// (defensive; if true something is wrong)
		t.Fatalf("expected no labels initially")
	}
	m.SetLabel("env", "prod")
	if v, ok := m.GetLabel("env"); !ok || v != "prod" {
		t.Fatalf("expected env=prod, got %q (ok=%v)", v, ok)
	}
	if !m.HasLabel("env", "prod") {
		t.Fatalf("HasLabel returned false")
	}
	m.DeleteLabel("env")
	if _, ok := m.GetLabel("env"); ok {
		t.Fatalf("expected label removed")
	}

	m.SetAnnotation("a", "b")
	if v, ok := m.GetAnnotation("a"); !ok || v != "b" {
		t.Fatalf("expected annotation a=b, got %q (ok=%v)", v, ok)
	}
	m.DeleteAnnotation("a")
	if _, ok := m.GetAnnotation("a"); ok {
		t.Fatalf("expected annotation removed")
	}
}

func TestMetadataStructuredHelpers(t *testing.T) {
	m := &Metadata{}
	list := []string{"one", "two"}
	if err := m.SetStringListAnnotation("list", list); err != nil {
		t.Fatalf("SetStringListAnnotation failed: %v", err)
	}
	got, found, err := m.GetStringListAnnotation("list")
	if err != nil || !found || len(got) != 2 || got[1] != "two" {
		t.Fatalf("unexpected list result: %v found=%v got=%v", err, found, got)
	}

	type sample struct {
		A int
		B string
	}
	val := sample{A: 7, B: "x"}
	if err := m.SetStructuredAnnotation("obj", val); err != nil {
		t.Fatalf("SetStructuredAnnotation failed: %v", err)
	}
	var out sample
	found, err = m.GetStructuredAnnotation("obj", &out)
	if err != nil || !found || out.A != 7 || out.B != "x" {
		t.Fatalf("unexpected structured result: %v found=%v out=%+v", err, found, out)
	}

	m.SetBoolAnnotation("flag", true)
	b, found, err := m.GetBoolAnnotation("flag")
	if err != nil || !found || !b {
		t.Fatalf("expected bool true got b=%v found=%v err=%v", b, found, err)
	}
	// Invalid bool
	m.SetAnnotation("flag2", "not-bool")
	if _, _, err := m.GetBoolAnnotation("flag2"); err == nil {
		t.Fatalf("expected error for invalid bool annotation")
	}
}

func TestClientPlayerServerFlow(t *testing.T) {
	url, _ := startEmbeddedNATSServer(t)
	cfg := &Config{
		NATSUrl:        url,
		PlayersBucket:  "players_test",
		ServersBucket:  "servers_test",
		ReconnectDelay: 100 * time.Millisecond,
		MaxReconnects:  1,
	}
	client, err := NewClient(context.Background(), cfg, newTestLogger())
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	defer client.Close()

	// Update player and server
	uuid := "player-123"
	if err := client.UpdatePlayer(uuid, func(m *Metadata) {
		m.SetLabel("tier", "gold")
		m.SetAnnotation(string(constant.PlayerName), "Hero")
	}); err != nil {
		t.Fatalf("UpdatePlayer failed: %v", err)
	}

	// Wait a bit for watcher to process put
	time.Sleep(200 * time.Millisecond)
	p, ok := client.GetPlayer(uuid)
	if !ok || !p.HasLabel("tier", "gold") {
		t.Fatalf("player not in cache or missing label: ok=%v p=%+v", ok, p)
	}
	if _, ok := client.GetPlayerByName("Hero"); !ok {
		t.Fatalf("GetPlayerByName failed")
	}

	if err := client.UpdateServer("srv-1", func(m *Metadata) { m.SetLabel("region", "eu") }); err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	if s, ok := client.GetServer("srv-1"); !ok || !s.HasLabel("region", "eu") {
		t.Fatalf("server missing after update")
	}

	players := client.GetPlayersByLabel("tier", "gold")
	if len(players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(players))
	}
	servers := client.GetServersByLabel("region", "eu")
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
}
