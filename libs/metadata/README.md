## metadata

High-level library for player and server metadata backed by NATS JetStream Key-Value stores. It maintains in-memory caches with live watchers, provides typed accessors, and emits change events.

### Features
- Auto-connect to NATS and create/get KV buckets for players and servers.
- Local caches with background watchers that stay in sync.
- Lookups by UUID, by player name, and by labels.
- Change notifications via a lightweight event bus.
- Generic, type-safe annotation descriptors for safe get/set.

---

## Installation & build
- Bazel target: `//libs/metadata:metadata`
- Go import: `github.com/bafbi/stellaroot/libs/metadata`

Run tests (integration tests skip if no local NATS at 127.0.0.1:4222):
```fish
bazel test //libs/metadata:metadata_test
```

---

## Configuration
Use `metadata.NewConfigFromEnv()` or construct `metadata.Config` directly.

Environment variables (defaults in parentheses):
- `NATS_URL` (nats://localhost:4222)
- `NATS_USER` ("")
- `NATS_PASSWORD` ("")
- `NATS_TOKEN` ("")
- `PLAYERS_BUCKET` (players)
- `SERVERS_BUCKET` (servers)

Programmatic:
```go
cfg := &metadata.Config{
	NATSUrl:        "nats://127.0.0.1:4222",
	PlayersBucket:  "players",
	ServersBucket:  "servers",
	ReconnectDelay: 5 * time.Second,
	MaxReconnects:  -1, // unlimited
}
```

---

## Quick start
```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
client, err := metadata.NewClient(context.Background(), cfg, logger)
defer client.Close()
if err != nil { /* handle */ }

// Update a player (creates if absent)
uuid := "player-123"
err = client.UpdatePlayer(uuid, func(m *metadata.Metadata) {
	m.SetLabel("tier", "gold")
	// using generated constant key
	m.SetAnnotation(constant.PlayerUsername, "Hero")
})

// Read
p, ok := client.GetPlayer(uuid)
byName, ok := client.GetPlayerByName("Hero")

// Query by label
goldPlayers := client.GetPlayersByLabel("tier", "gold")
```

---

## Client API (overview)
- Construction/teardown
	- `NewClient(ctx, cfg, logger) (*Client, error)`
	- `(*Client).Close() error`

- Players
	- `GetPlayer(uuid string) (*Metadata, bool)`
	- `GetPlayerByName(name string) (*Metadata, bool)`
	- `UpdatePlayer(uuid string, fn func(*Metadata)) error`
	- `UpdatePlayerByName(name string, fn func(*Metadata)) error`
	- `GetPlayersByLabel(key, value string) map[string]*Metadata`
	- `GetPlayersByLabels(labels map[string]string) map[string]*Metadata`

- Servers
	- `GetServer(name string) (*Metadata, bool)`
	- `GetAllServers() map[string]*Metadata`
	- `UpdateServer(name string, fn func(*Metadata)) error`
	- `GetServersByLabel(key, value string) map[string]*Metadata`
	- `GetServersByLabels(labels map[string]string) map[string]*Metadata`

- Events and health
	- `SubscribeToPlayerChanges(cb MetadataChangeCallback) (unsubscribe func())`
	- `SubscribeToServerChanges(cb MetadataChangeCallback) (unsubscribe func())`
	- `WatcherStatusChan() <-chan WatcherStatus`

Types used by events:
- `MetadataChangeEvent{ Key string, OldValue *Metadata, NewValue *Metadata, Type ChangeType }`
- `ChangeType{ ChangeTypePut, ChangeTypeDelete }`
- `WatcherStatus{ Watcher string, Healthy bool, Error error }`

Event topics:
- `player.change`
- `server.change`

Example subscription:
```go
unsub := client.SubscribeToPlayerChanges(func(e metadata.MetadataChangeEvent){
	if e.Type == metadata.ChangeTypePut {
		logger.Info("player updated", "uuid", e.Key)
	}
})
defer unsub()
```

---

## Metadata structure
```go
type Metadata struct {
	Labels      map[string]string // user-defined
	Annotations map[string]string // system/tool-defined
}
```

Helpers on `Metadata`:
- Labels: `SetLabel`, `GetLabel`, `DeleteLabel`, `HasLabel`, `HasLabels(map[string]string)`
- Annotations (string-based):
	- `SetAnnotation(constant.AnnotationKey, string)`, `GetAnnotation(constant.AnnotationKey)`
	- `DeleteAnnotation(constant.AnnotationKey)`, `HasAnnotation(constant.AnnotationKey, string)`
	- Structured helpers: `SetStringListAnnotation`, `GetStringListAnnotation`, `SetStructuredAnnotation`, `GetStructuredAnnotation`, `SetBoolAnnotation`, `GetBoolAnnotation`

---

## Type-safe annotation access (descriptors)
Use generic descriptors for compile-time value typing and runtime validation. Constructors:
- `NewStringAnnotationDesc(key constant.AnnotationKey)`
- `NewBoolAnnotationDesc(key constant.AnnotationKey)`
- `NewTimeAnnotationDesc(key constant.AnnotationKey, layout string)`
- `NewUUIDAnnotationDesc(key constant.AnnotationKey)`
- `NewEnumAnnotationDesc[T ~string](key constant.AnnotationKey, allowed []T)`

Generic accessors:
- `metadata.Get(m *Metadata, d AnnotationDescriptor[T]) (T, bool, error)`
- `metadata.Set(m *Metadata, d AnnotationDescriptor[T], v T)`

Example:
```go
// bool
online, found, err := metadata.Get(p, metadata.PlayerOnlineDesc)
metadata.Set(p, metadata.PlayerOnlineDesc, true)

// time (RFC3339)
lastLogin, found, err := metadata.Get(p, metadata.PlayerLastLoginDesc)
metadata.Set(p, metadata.PlayerLastLoginDesc, time.Now().UTC())

// enum (requires corresponding enum + All<Enum>s slice in constants)
status, found, err := metadata.Get(p, metadata.PlayerStatusDesc)
```

Notes:
- Missing annotation -> `(zero, false, nil)`.
- Present but invalid -> `(zero, true, error)`.
- Descriptor variables like `PlayerOnlineDesc` are generated from YAML (descriptors mode) into this package.

---

## Buckets and cache behavior
- Buckets named via config (PlayersBucket, ServersBucket). The client will CreateKeyValue, and if exists, fallback to KeyValue.
- On start, client warms both caches by listing keys and reading values.
- Two watcher goroutines keep caches in sync and publish change events.
- Health updates on `WatcherStatusChan()` when watchers are healthy/unhealthy.

---

## Local development
Start a NATS server locally (optional for unit tests; integration tests skip if unreachable):
```fish
# Example using docker (adjust as needed)
docker run --rm -p 4222:4222 nats:2
```

---

## Gotchas & tips
- Always call `Close()` to stop watchers and close the NATS connection.
- For name lookups, the mapping uses the `constant.PlayerUsername` annotation.
- Use descriptors everywhere to avoid stringly-typed mistakes.
- For cross-service constants, add keys/enums in `libs/constant/constants.yaml` and rebuild.
