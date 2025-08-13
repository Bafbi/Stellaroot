package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/asaskevich/EventBus"
	"github.com/nats-io/nats.go"

	"github.com/bafbi/stellaroot/libs/constant"
)

type ChangeType int

const (
	ChangeTypePut ChangeType = iota
	ChangeTypeDelete
)

const (
	PlayerChangeEventKey = "player.change"
	ServerChangeEventKey = "server.change"
)

// MetadataChangeEvent represents a change event for metadata.
type MetadataChangeEvent struct {
	Key      string
	OldValue *Metadata
	NewValue *Metadata
	Type     ChangeType
}

// Subscription callback type.
type MetadataChangeCallback func(event MetadataChangeEvent)

type WatcherStatus struct {
	Watcher string // "players" or "servers"
	Healthy bool
	Error   error
}

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

	// --- Initial cache warm-up for players ---
	entries, err := client.playersKV.Keys()
	if err != nil && err != nats.ErrNoKeysFound {
		cancel()
		return nil, fmt.Errorf("failed to list player keys: %w", err)
	}
	for _, uuid := range entries {
		val, err := client.playersKV.Get(uuid)
		if err != nil || val == nil {
			continue
		}
		var player Metadata
		if err := json.Unmarshal(val.Value(), &player); err != nil {
			continue
		}
		client.playersMu.Lock()
		client.playersCache[uuid] = &player
		if player.Annotations != nil {
			if name, ok := player.Annotations[string(constant.PlayerName)]; ok && name != "" {
				client.playersNameToUUID[name] = uuid
			}
		}
		client.playersMu.Unlock()
	}

	// --- Initial cache warm-up for servers ---
	serverKeys, err := client.serversKV.Keys()
	if err != nil && err != nats.ErrNoKeysFound {
		cancel()
		return nil, fmt.Errorf("failed to list server keys: %w", err)
	}
	for _, name := range serverKeys {
		val, err := client.serversKV.Get(name)
		if err != nil || val == nil {
			continue
		}
		var server Metadata
		if err := json.Unmarshal(val.Value(), &server); err != nil {
			continue
		}
		client.serversMu.Lock()
		client.serversCache[name] = &server
		client.serversMu.Unlock()
	}

	client.wg.Add(2)
	go client.watchPlayers()
	go client.watchServers()

	return client, nil
}

func (c *Client) connect() error {
	opts := []nats.Option{
		nats.Name("MetadataClient"),
		nats.ReconnectWait(c.config.ReconnectDelay),
		nats.MaxReconnects(c.config.MaxReconnects),
	}

	if c.config.NATSUser != "" && c.config.NATSPassword != "" {
		opts = append(opts, nats.UserInfo(c.config.NATSUser, c.config.NATSPassword))
	}

	if c.config.NATSToken != "" {
		opts = append(opts, nats.Token(c.config.NATSToken))
	}

	nc, err := nats.Connect(c.config.NATSUrl, opts...)
	if err != nil {
		return err
	}

	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	c.nc = nc
	c.js = js
	return nil
}

func (c *Client) initKV() error {
	// Create or get players bucket
	playersKV, err := c.js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: c.config.PlayersBucket,
	})
	if err != nil {
		// Try to get existing bucket
		playersKV, err = c.js.KeyValue(c.config.PlayersBucket)
		if err != nil {
			return fmt.Errorf("failed to create/get players bucket: %w", err)
		}
	}
	c.playersKV = playersKV

	// Create or get servers bucket
	serversKV, err := c.js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: c.config.ServersBucket,
	})
	if err != nil {
		// Try to get existing bucket
		serversKV, err = c.js.KeyValue(c.config.ServersBucket)
		if err != nil {
			return fmt.Errorf("failed to create/get servers bucket: %w", err)
		}
	}
	c.serversKV = serversKV

	return nil
}

// SubscribeToPlayerChanges registers a callback for player metadata changes.
func (c *Client) SubscribeToPlayerChanges(cb MetadataChangeCallback) (unsubscribe func()) {
	c.eventBus.Subscribe(PlayerChangeEventKey, cb)
	return func() {
		c.eventBus.Unsubscribe(PlayerChangeEventKey, cb)
	}
}

// SubscribeToServerChanges registers a callback for server metadata changes.
func (c *Client) SubscribeToServerChanges(cb MetadataChangeCallback) (unsubscribe func()) {
	c.eventBus.Subscribe(ServerChangeEventKey, cb)
	return func() {
		c.eventBus.Unsubscribe(ServerChangeEventKey, cb)
	}
}

// WatcherStatusChan returns a channel for watcher health status updates.
func (c *Client) WatcherStatusChan() <-chan WatcherStatus {
	return c.watcherStatusCh
}

func (c *Client) watchPlayers() {
	defer c.wg.Done()
	for {
		watcher, err := c.playersKV.WatchAll()
		if err != nil {
			c.logger.Error("Failed to create players watcher", "error", err)
			select {
			case c.watcherStatusCh <- WatcherStatus{Watcher: "players", Healthy: false, Error: err}:
			default:
			}
			select {
			case <-time.After(2 * time.Second):
				// retry after delay
				continue
			case <-c.ctx.Done():
				return
			}
		}
		select {
		case c.watcherStatusCh <- WatcherStatus{Watcher: "players", Healthy: true, Error: nil}:
		default:
		}
		defer watcher.Stop()

		for {
			select {
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				uuid := entry.Key()
				var oldValue *Metadata
				var changeType ChangeType

				// Handle deletion
				if entry.Operation() == nats.KeyValueDelete {
					var oldName string
					c.playersMu.RLock()
					if oldPlayer, exists := c.playersCache[uuid]; exists && oldPlayer.Annotations != nil {
						oldName, _ = oldPlayer.Annotations[string(constant.PlayerName)]
						oldValue = oldPlayer
					}
					c.playersMu.RUnlock()

					c.playersMu.Lock()
					if oldName != "" {
						delete(c.playersNameToUUID, oldName)
					}
					delete(c.playersCache, uuid)
					c.playersMu.Unlock()

					changeType = ChangeTypeDelete

					// Notify subscribers
					c.eventBus.Publish(PlayerChangeEventKey, MetadataChangeEvent{
						Key:      uuid,
						OldValue: oldValue,
						NewValue: nil,
						Type:     changeType,
					})
					continue
				}

				var player Metadata
				if err := json.Unmarshal(entry.Value(), &player); err != nil {
					c.logger.Warn("Failed to unmarshal player metadata", "error", err)
					continue
				}

				var oldName, newName string
				c.playersMu.RLock()
				if oldPlayer, exists := c.playersCache[uuid]; exists && oldPlayer.Annotations != nil {
					oldName, _ = oldPlayer.Annotations[string(constant.PlayerName)]
					oldValue = oldPlayer
				}
				c.playersMu.RUnlock()
				if player.Annotations != nil {
					newName, _ = player.Annotations[string(constant.PlayerName)]
				}

				c.playersMu.Lock()
				if oldName != "" {
					delete(c.playersNameToUUID, oldName)
				}
				if newName != "" {
					c.playersNameToUUID[newName] = uuid
				}
				c.playersCache[uuid] = &player
				c.playersMu.Unlock()

				changeType = ChangeTypePut

				// Notify subscribers
				c.eventBus.Publish(PlayerChangeEventKey, MetadataChangeEvent{
					Key:      uuid,
					OldValue: oldValue,
					NewValue: &player,
					Type:     changeType,
				})

			case <-c.ctx.Done():
				return
			}
		}
	}
}

func (c *Client) watchServers() {
	defer c.wg.Done()
	for {
		watcher, err := c.serversKV.WatchAll()
		if err != nil {
			c.logger.Error("Failed to create servers watcher", "error", err)
			select {
			case c.watcherStatusCh <- WatcherStatus{Watcher: "servers", Healthy: false, Error: err}:
			default:
			}
			select {
			case <-time.After(2 * time.Second):
				// retry after delay
				continue
			case <-c.ctx.Done():
				return
			}
		}
		select {
		case c.watcherStatusCh <- WatcherStatus{Watcher: "servers", Healthy: true, Error: nil}:
		default:
		}
		defer watcher.Stop()

		for {
			select {
			case entry := <-watcher.Updates():
				if entry == nil {
					continue
				}

				var oldValue *Metadata
				name := entry.Key()
				var changeType ChangeType

				// Handle deletion
				if entry.Operation() == nats.KeyValueDelete {
					c.serversMu.RLock()
					if oldServer, exists := c.serversCache[name]; exists {
						oldValue = oldServer
					}
					c.serversMu.RUnlock()

					c.serversMu.Lock()
					delete(c.serversCache, name)
					c.serversMu.Unlock()

					changeType = ChangeTypeDelete

					// Notify subscribers
					c.eventBus.Publish(ServerChangeEventKey, MetadataChangeEvent{
						Key:      name,
						OldValue: oldValue,
						NewValue: nil,
						Type:     changeType,
					})
					continue
				}

				var server Metadata
				if err := json.Unmarshal(entry.Value(), &server); err != nil {
					c.logger.Warn("Failed to unmarshal server metadata", "error", err)
					continue
				}

				c.serversMu.RLock()
				if oldServer, exists := c.serversCache[name]; exists {
					oldValue = oldServer
				}
				c.serversMu.RUnlock()

				c.serversMu.Lock()
				c.serversCache[name] = &server
				c.serversMu.Unlock()

				changeType = ChangeTypePut

				// Notify subscribers
				c.eventBus.Publish(ServerChangeEventKey, MetadataChangeEvent{
					Key:      name,
					OldValue: oldValue,
					NewValue: &server,
					Type:     changeType,
				})

			case <-c.ctx.Done():
				return
			}
		}
	}
}

// Player methods
func (c *Client) GetPlayer(uuid string) (*Metadata, bool) {
	c.playersMu.RLock()
	defer c.playersMu.RUnlock()
	player, exists := c.playersCache[uuid]
	return player, exists
}

func (c *Client) GetPlayerByName(name string) (*Metadata, bool) {
	c.playersMu.RLock()
	defer c.playersMu.RUnlock()

	if uuid, exists := c.playersNameToUUID[name]; exists {
		if player, playerExists := c.playersCache[uuid]; playerExists {
			return player, true
		}
	}
	return nil, false
}

func (c *Client) UpdatePlayer(uuid string, updateFunc func(*Metadata)) error {
	c.playersMu.Lock()
	defer c.playersMu.Unlock()

	player, exists := c.playersCache[uuid]
	if !exists {
		// Create new player if doesn't exist
		player = &Metadata{}
	}

	// Create a copy to avoid race conditions
	playerCopy := &Metadata{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	if player.Labels != nil {
		for k, v := range player.Labels {
			playerCopy.Labels[k] = v
		}
	}

	if player.Annotations != nil {
		for k, v := range player.Annotations {
			playerCopy.Annotations[k] = v
		}
	}

	updateFunc(playerCopy)

	data, err := json.Marshal(playerCopy)
	if err != nil {
		return err
	}
	_, err = c.playersKV.Put(uuid, data)
	return err
}

func (c *Client) UpdatePlayerByName(name string, updateFunc func(*Metadata)) error {
	c.playersMu.RLock()
	uuid, exists := c.playersNameToUUID[name]
	c.playersMu.RUnlock()

	if !exists {
		return fmt.Errorf("player with name '%s' not found", name)
	}

	return c.UpdatePlayer(uuid, updateFunc)
}

// Server methods
func (c *Client) GetServer(name string) (*Metadata, bool) {
	c.serversMu.RLock()
	defer c.serversMu.RUnlock()
	server, exists := c.serversCache[name]
	return server, exists
}

func (c *Client) GetAllServers() map[string]*Metadata {
	c.serversMu.RLock()
	defer c.serversMu.RUnlock()

	result := make(map[string]*Metadata)
	for k, v := range c.serversCache {
		result[k] = v
	}
	return result
}

func (c *Client) UpdateServer(name string, updateFunc func(*Metadata)) error {
	c.serversMu.Lock()
	defer c.serversMu.Unlock()

	server, exists := c.serversCache[name]
	if !exists {
		// Create new server if doesn't exist
		server = &Metadata{}
	}

	// Create a copy to avoid race conditions
	serverCopy := &Metadata{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	if server.Labels != nil {
		for k, v := range server.Labels {
			serverCopy.Labels[k] = v
		}
	}

	if server.Annotations != nil {
		for k, v := range server.Annotations {
			serverCopy.Annotations[k] = v
		}
	}

	updateFunc(serverCopy)

	data, err := json.Marshal(serverCopy)
	if err != nil {
		return err
	}
	_, err = c.serversKV.Put(name, data)
	return err
}

// Query methods
func (c *Client) GetPlayersByLabel(key, value string) map[string]*Metadata {
	c.playersMu.RLock()
	defer c.playersMu.RUnlock()

	result := make(map[string]*Metadata)
	for uuid, player := range c.playersCache {
		if player.HasLabel(key, value) {
			result[uuid] = player
		}
	}
	return result
}

func (c *Client) GetPlayersByLabels(labels map[string]string) map[string]*Metadata {
	c.playersMu.RLock()
	defer c.playersMu.RUnlock()

	result := make(map[string]*Metadata)
	for uuid, player := range c.playersCache {
		if player.HasLabels(labels) {
			result[uuid] = player
		}
	}
	return result
}

func (c *Client) GetServersByLabel(key, value string) map[string]*Metadata {
	c.serversMu.RLock()
	defer c.serversMu.RUnlock()

	result := make(map[string]*Metadata)
	for name, server := range c.serversCache {
		if server.HasLabel(key, value) {
			result[name] = server
		}
	}
	return result
}

func (c *Client) GetServersByLabels(labels map[string]string) map[string]*Metadata {
	c.serversMu.RLock()
	defer c.serversMu.RUnlock()

	result := make(map[string]*Metadata)
	for name, server := range c.serversCache {
		if server.HasLabels(labels) {
			result[name] = server
		}
	}
	return result
}

func (c *Client) Close() error {
	c.cancel()
	c.wg.Wait()
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}
