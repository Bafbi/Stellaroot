package metadata

import (
	"encoding/json"

	"github.com/nats-io/nats.go"

	"github.com/bafbi/stellaroot/libs/constant"
)

// warmUpCaches loads initial state from KV into memory caches.
func (c *Client) warmUpCaches() error {
	// Players
	entries, err := c.playersKV.Keys()
	if err != nil && err != nats.ErrNoKeysFound {
		return err
	}
	for _, uuid := range entries {
		val, err := c.playersKV.Get(uuid)
		if err != nil || val == nil {
			continue
		}
		var player Metadata
		if err := json.Unmarshal(val.Value(), &player); err != nil {
			continue
		}
		c.playersMu.Lock()
		c.playersCache[uuid] = &player
		if player.Annotations != nil {
			if name, ok := player.Annotations[string(constant.PlayerUsername)]; ok && name != "" {
				c.playersNameToUUID[name] = uuid
			}
		}
		c.playersMu.Unlock()
	}

	// Servers
	serverKeys, err := c.serversKV.Keys()
	if err != nil && err != nats.ErrNoKeysFound {
		return err
	}
	for _, name := range serverKeys {
		val, err := c.serversKV.Get(name)
		if err != nil || val == nil {
			continue
		}
		var server Metadata
		if err := json.Unmarshal(val.Value(), &server); err != nil {
			continue
		}
		c.serversMu.Lock()
		c.serversCache[name] = &server
		c.serversMu.Unlock()
	}
	return nil
}
