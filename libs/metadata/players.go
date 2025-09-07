package metadata

import (
	"encoding/json"
	"fmt"
)

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
		player = &Metadata{}
	}

	playerCopy := &Metadata{Labels: make(map[string]string), Annotations: make(map[string]string)}
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
