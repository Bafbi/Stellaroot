package metadata

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/bafbi/stellaroot/libs/constant"
)

// SubscribeToPlayerChanges registers a callback for player metadata changes.
func (c *Client) SubscribeToPlayerChanges(cb MetadataChangeCallback) (unsubscribe func()) {
	c.eventBus.Subscribe(PlayerChangeEventKey, cb)
	return func() { c.eventBus.Unsubscribe(PlayerChangeEventKey, cb) }
}

// SubscribeToServerChanges registers a callback for server metadata changes.
func (c *Client) SubscribeToServerChanges(cb MetadataChangeCallback) (unsubscribe func()) {
	c.eventBus.Subscribe(ServerChangeEventKey, cb)
	return func() { c.eventBus.Unsubscribe(ServerChangeEventKey, cb) }
}

// WatcherStatusChan returns a channel for watcher health status updates.
func (c *Client) WatcherStatusChan() <-chan WatcherStatus { return c.watcherStatusCh }

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

				if entry.Operation() == nats.KeyValueDelete {
					var oldName string
					c.playersMu.RLock()
					if oldPlayer, exists := c.playersCache[uuid]; exists && oldPlayer.Annotations != nil {
						oldName, _ = oldPlayer.Annotations[string(constant.PlayerUsername)]
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

					c.eventBus.Publish(PlayerChangeEventKey, MetadataChangeEvent{Key: uuid, OldValue: oldValue, NewValue: nil, Type: changeType})
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
					oldName, _ = oldPlayer.Annotations[string(constant.PlayerUsername)]
					oldValue = oldPlayer
				}
				c.playersMu.RUnlock()
				if player.Annotations != nil {
					newName, _ = player.Annotations[string(constant.PlayerUsername)]
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
				c.eventBus.Publish(PlayerChangeEventKey, MetadataChangeEvent{Key: uuid, OldValue: oldValue, NewValue: &player, Type: changeType})

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
					c.eventBus.Publish(ServerChangeEventKey, MetadataChangeEvent{Key: name, OldValue: oldValue, NewValue: nil, Type: changeType})
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
				c.eventBus.Publish(ServerChangeEventKey, MetadataChangeEvent{Key: name, OldValue: oldValue, NewValue: &server, Type: changeType})

			case <-c.ctx.Done():
				return
			}
		}
	}
}
