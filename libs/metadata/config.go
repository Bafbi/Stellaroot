package metadata

import (
	"os"
	"time"
)

type Config struct {
	NATSUrl      string
	NATSUser     string
	NATSPassword string
	NATSToken    string

	PlayersBucket string
	ServersBucket string

	ReconnectDelay time.Duration
	MaxReconnects  int
}

func NewConfigFromEnv() *Config {
	return &Config{
		NATSUrl:        getEnv("NATS_URL", "nats://localhost:4222"),
		NATSUser:       getEnv("NATS_USER", ""),
		NATSPassword:   getEnv("NATS_PASSWORD", ""),
		NATSToken:      getEnv("NATS_TOKEN", ""),
		PlayersBucket:  getEnv("PLAYERS_BUCKET", "players"),
		ServersBucket:  getEnv("SERVERS_BUCKET", "servers"),
		ReconnectDelay: 5 * time.Second,
		MaxReconnects:  -1, // unlimited
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
