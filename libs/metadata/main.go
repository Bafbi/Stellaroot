package metadata

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Hello metadata service")

	config := NewConfigFromEnv()
	client, err := NewClient(context.Background(), config, logger.With("component", "metadata", "service", "v1.0.0"))
	if err != nil {
		logger.Error("Failed to create metadata client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	logger.Info("Successfully connected to NATS and initialized KV buckets!")

	// Example usage with new API
	err = client.UpdatePlayer("example-uuid", func(m *Metadata) {
		m.SetLabel("tier", "premium")
		m.SetLabel("region", "us-east")
		m.SetAnnotation("last_login_ip", "192.168.1.100")
		m.SetAnnotation("client_version", "1.20.1")
		m.SetAnnotation("permission_sync", "pending")
	})
	if err != nil {
		logger.Warn("Failed to update player", "error", err)
	}

	// Example server with labels and annotations
	err = client.UpdateServer("survival-1", func(m *Metadata) {
		m.SetLabel("environment", "production")
		m.SetLabel("game_mode", "survival")
		m.SetLabel("region", "us-east")
		m.SetAnnotation("last_restart", time.Now().Add(-2*time.Hour).Format(time.RFC3339))
		m.SetAnnotation("max_players", "100")
	})
	if err != nil {
		logger.Warn("Failed to update server", "error", err)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down metadata service...")
}
