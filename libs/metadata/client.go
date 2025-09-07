package metadata

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/asaskevich/EventBus"
	"github.com/nats-io/nats.go"
)

type Client struct {
	config *Config
	nc     *nats.Conn
	js     nats.JetStreamContext

	playersKV nats.KeyValue
	serversKV nats.KeyValue

	playersCache      map[string]*Metadata
	playersNameToUUID map[string]string // maps player name to UUID
	serversCache      map[string]*Metadata

	playersMu sync.RWMutex
	serversMu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	logger *slog.Logger

	wg sync.WaitGroup

	eventBus EventBus.Bus

	watcherStatusCh chan WatcherStatus
}

func NewClient(parentCtx context.Context, config *Config, logger *slog.Logger) (*Client, error) {
	ctx, cancel := context.WithCancel(parentCtx)

	client := &Client{
		config:            config,
		playersCache:      make(map[string]*Metadata),
		playersNameToUUID: make(map[string]string),
		serversCache:      make(map[string]*Metadata),
		ctx:               ctx,
		cancel:            cancel,
		logger:            logger,
		eventBus:          EventBus.New(),
		watcherStatusCh:   make(chan WatcherStatus, 4),
	}

	if err := client.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err := client.initKV(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize KV: %w", err)
	}

	if err := client.warmUpCaches(); err != nil {
		cancel()
		return nil, err
	}

	client.wg.Add(2)
	go client.watchPlayers()
	go client.watchServers()

	return client, nil
}

func (c *Client) Close() error {
	c.cancel()
	c.wg.Wait()
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}
