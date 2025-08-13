package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bafbi/stellaroot/libs/metadata"
)

type options struct {
	players  int
	servers  int
	prefix   string
	updates  bool
	interval time.Duration
	seed     int64
}

func parseFlags() options {
	var o options
	flag.IntVar(&o.players, "players", envInt("FAKER_PLAYERS", 25), "number of players to seed")
	flag.IntVar(&o.servers, "servers", envInt("FAKER_SERVERS", 5), "number of servers to seed")
	flag.StringVar(&o.prefix, "prefix", envStr("FAKER_PREFIX", "local"), "name prefix for generated data")
	flag.BoolVar(&o.updates, "updates", envBool("FAKER_UPDATES", false), "whether to keep updating values periodically")
	flag.DurationVar(&o.interval, "interval", envDuration("FAKER_INTERVAL", 5*time.Second), "update interval when -updates is set")
	flag.Int64Var(&o.seed, "seed", envInt64("FAKER_SEED", time.Now().UnixNano()), "PRNG seed")
	flag.Parse()
	return o
}

func envStr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			return x
		}
	}
	return def
}

func envInt64(k string, def int64) int64 {
	if v := os.Getenv(k); v != "" {
		if x, err := strconv.ParseInt(v, 10, 64); err == nil {
			return x
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v := strings.ToLower(os.Getenv(k)); v != "" {
		return v == "1" || v == "true" || v == "yes" || v == "on"
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	opts := parseFlags()
	rand.Seed(opts.seed)

	cfg := metadata.NewConfigFromEnv()
	client, err := metadata.NewClient(context.Background(), cfg, logger.With("component", "fakedata"))
	if err != nil {
		logger.Error("failed to create metadata client", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	// Seed servers first so players can reference a server in annotations if desired.
	serverNames := seedServers(client, logger, opts)
	seedPlayers(client, logger, opts, serverNames)

	if !opts.updates {
		logger.Info("seed completed", "players", opts.players, "servers", opts.servers)
		return
	}

	// Periodic updates
	logger.Info("starting periodic updates", "interval", opts.interval.String())
	ticker := time.NewTicker(opts.interval)
	defer ticker.Stop()

	// Handle SIGINT/SIGTERM
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			randomServerUpdates(client, serverNames)
			randomPlayerUpdates(client, serverNames)
		case <-done:
			logger.Info("stopping updates")
			return
		}
	}
}

func seedServers(client *metadata.Client, logger *slog.Logger, opts options) []string {
	regions := []string{"us-east", "us-west", "eu-west"}
	modes := []string{"survival", "creative", "minigames"}
	names := make([]string, 0, opts.servers)
	for i := 1; i <= opts.servers; i++ {
		name := fmt.Sprintf("%s-srv-%d", opts.prefix, i)
		names = append(names, name)
		region := regions[rand.Intn(len(regions))]
		mode := modes[rand.Intn(len(modes))]
		status := []string{"online", "offline"}[rand.Intn(2)]
		currentPlayers := rand.Intn(50)

		_ = client.UpdateServer(name, func(m *metadata.Metadata) {
			m.SetLabel("region", region)
			m.SetLabel("game_mode", mode)
			m.SetAnnotation("status", status)
			m.SetAnnotation("current_players", fmt.Sprintf("%d", currentPlayers))
		})
	}
	logger.Info("seeded servers", "count", len(names))
	return names
}

func seedPlayers(client *metadata.Client, logger *slog.Logger, opts options, serverNames []string) {
	regions := []string{"us-east", "us-west", "eu-west"}
	tiers := []string{"free", "premium"}
	for i := 1; i <= opts.players; i++ {
		uuid := pseudoUUID()
		name := fmt.Sprintf("%s-player-%03d", opts.prefix, i)
		region := regions[rand.Intn(len(regions))]
		tier := tiers[rand.Intn(len(tiers))]
		online := rand.Intn(100) < 60 // ~60% online
		var server string
		if len(serverNames) > 0 && online {
			server = serverNames[rand.Intn(len(serverNames))]
		}

		_ = client.UpdatePlayer(uuid, func(m *metadata.Metadata) {
			m.SetLabel("tier", tier)
			m.SetLabel("region", region)
			// canonical player name annotation key
			m.SetAnnotation(schema.S(schema.PlayerAnnotationKeyName), name)
			// dashboard uses "online" annotation
			m.SetAnnotation("online", fmt.Sprintf("%t", online))
			if server != "" {
				m.SetAnnotation("current_server", server)
			}
		})
	}
	logger.Info("seeded players", "count", opts.players)
}

func randomServerUpdates(client *metadata.Client, serverNames []string) {
	if len(serverNames) == 0 {
		return
	}
	// pick 25% of servers to update
	n := max(1, len(serverNames)/4)
	for i := 0; i < n; i++ {
		name := serverNames[rand.Intn(len(serverNames))]
		delta := rand.Intn(7) - 3 // -3..+3
		_ = client.UpdateServer(name, func(m *metadata.Metadata) {
			// flip a coin for status
			if rand.Intn(10) == 0 {
				cur := strings.ToLower(m.Annotations["status"])
				if cur == "online" {
					m.SetAnnotation("status", "offline")
				} else {
					m.SetAnnotation("status", "online")
				}
			}
			// adjust player count within 0..100
			cur := 0
			if v, ok := m.GetLabel("max_players"); ok {
				_ = v // not used, placeholder for future
			}
			if s, ok := m.GetAnnotation("current_players"); ok {
				var x int
				fmt.Sscanf(s, "%d", &x)
				cur = x
			}
			cur += delta
			if cur < 0 {
				cur = 0
			}
			if cur > 100 {
				cur = 100
			}
			m.SetAnnotation("current_players", fmt.Sprintf("%d", cur))
		})
	}
}

func randomPlayerUpdates(client *metadata.Client, serverNames []string) {
	// toggle a handful of random players online/offline by scanning cached map
	players := client.GetPlayersByLabels(map[string]string{})
	if len(players) == 0 {
		return
	}
	uuids := make([]string, 0, len(players))
	for uuid := range players {
		uuids = append(uuids, uuid)
	}
	n := max(1, len(uuids)/10)
	for i := 0; i < n; i++ {
		uuid := uuids[rand.Intn(len(uuids))]
		_ = client.UpdatePlayer(uuid, func(m *metadata.Metadata) {
			cur, _ := m.GetAnnotation("online")
			next := "true"
			if cur == "true" {
				next = "false"
			}
			m.SetAnnotation("online", next)
			if next == "true" && len(serverNames) > 0 {
				m.SetAnnotation("current_server", serverNames[rand.Intn(len(serverNames))])
			} else {
				m.DeleteAnnotation("current_server")
			}
		})
	}
}

func pseudoUUID() string {
	// Generate a simple 32-hex string with dashes for readability. Not cryptographically secure.
	const hex = "0123456789abcdef"
	b := make([]byte, 32)
	for i := range b {
		b[i] = hex[rand.Intn(len(hex))]
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", b[0:8], b[8:12], b[12:16], b[16:20], b[20:32])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
