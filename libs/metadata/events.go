package metadata

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

// WatcherStatus holds watcher health.
type WatcherStatus struct {
	Watcher string // "players" or "servers"
	Healthy bool
	Error   error
}
