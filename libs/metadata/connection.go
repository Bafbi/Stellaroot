package metadata

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

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
