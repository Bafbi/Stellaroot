package schema

import (
	"reflect"

	pb "github.com/bafbi/stellaroot/libs/schema/protos"
)

// AnnotationKey represents a canonical annotation key.
type AnnotationKey string

const (
	AnnotationKeyPlayerList   AnnotationKey = "server/player.list"
	AnnotationKeyServerStatus AnnotationKey = "server/status"
	// ... add more as needed
)

const (
	PlayerAnnotationKeyName AnnotationKey = "player/name"
)

// NATS Subject constants
const (
	NATSPlayersLocationUpdates = "updates.players.location"
	NATSServerLifecycleEvents  = "events.servers.lifecycle"
	// ... add more as needed
)

// AnnotationKeyRegistry maps AnnotationKey to the reflect.Type of its expected Go struct.
// This is used for type checking and unmarshaling.
var AnnotationKeyRegistry = map[AnnotationKey]reflect.Type{
	AnnotationKeyPlayerList:   reflect.TypeOf(pb.PlayerListAnnotation{}),
	AnnotationKeyServerStatus: reflect.TypeOf(pb.ServerStatusAnnotation{}),
}

// GetAnnotationSchemaType provides the expected type for a given annotation key.
func GetAnnotationSchemaType(key AnnotationKey) (reflect.Type, bool) {
	t, ok := AnnotationKeyRegistry[key]
	return t, ok
}

func GetAnnotationKeyAsString(key AnnotationKey) string {
	return string(key)
}

func S(key AnnotationKey) string {
	return GetAnnotationKeyAsString(key)
}
