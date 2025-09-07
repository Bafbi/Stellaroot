package metadata

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
