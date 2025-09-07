package metadata

import "encoding/json"

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
		server = &Metadata{}
	}

	serverCopy := &Metadata{Labels: make(map[string]string), Annotations: make(map[string]string)}
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
