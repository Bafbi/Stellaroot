# genconstants

Generate Go code from a small YAML spec. It produces:
- constants mode: typed string constants and helper functions.
- descriptors mode: typed annotation descriptor variables for safe get/set with the metadata package.

Use it to keep wire values in one place and get compile-time help across services.

## Quick start
1) Minimal YAML (libs/constant/constants.yaml)
```yaml
version: 1
enums: []
constants:
  - name: PLAYER_USERNAME
    group: annotations
    wire: player/username
    value_kind: string
    description: Player username annotation
  - name: PLAYER_ONLINE
    group: annotations
    wire: player/online
    value_kind: boolean
    description: Player online status annotation
```

2) Generate constants (default mode)
```fish
bazel build //libs/constant:generate_constants
```
This creates a Go file with:
- Types: AnnotationKey, LabelKey, NatsSubject, KvBucket, SubjectTemplate
- Constants: PlayerUsername = "player/username", PlayerOnline = "player/online"
- Slices: AllAnnotationKeys, etc.
- For subject_templates: a <Base>Subject(...) builder function.

3) Generate descriptors (optional)
We also generate typed descriptors for annotations into the metadata package:
```fish
bazel build //libs/metadata:generate_annotation_descriptors
```
You can then use, e.g., PlayerOnlineDesc with metadata.Get/Set.

## YAML spec (comprehensive)
Top-level fields:
- version: integer (required). Currently 1.
- enums: list of enum types (optional).
- constants: list of constants (optional).

### Enums
Each enum yields a string-typed Go enum and an All<Enum>s slice.
```yaml
enums:
  - name: PlayerStatus            # exported Go type name
    description: Lifecycle status
    values:
      - name: ONLINE              # turns into constant Online
        value: online             # wire value (string)
        description: Player is connected
      - name: OFFLINE
        value: offline
```
Generated (excerpt):
```go
type PlayerStatus string
const (
  Online PlayerStatus = "online"
  Offline PlayerStatus = "offline"
)
var AllPlayerStatuss = []PlayerStatus{Online, Offline}
```

### Constants
Common fields for each constant:
- name: UPPER_SNAKE identifier (required). Example: PLAYER_USERNAME
- group: one of
  - annotations -> generates AnnotationKey
  - labels -> LabelKey
  - nats_subjects -> NatsSubject
  - kv_buckets -> KvBucket
  - subject_templates -> SubjectTemplate + a builder function
- wire: the wire string value (required). Examples: player/username, system.health, players
- value_kind: semantic hint used by descriptors and docs (optional but recommended).
  - string
  - boolean
  - uuid
  - rfc3339_timestamp
  - enum:<EnumType> (e.g., enum:PlayerStatus)
  - template (only for subject_templates; informative)
- description: short text (optional). Becomes doc comment.

For subject_templates you can add variables:
```yaml
- name: PLAYER_EVENTS_TEMPLATE
  group: subject_templates
  wire: player.{player_id}.events
  value_kind: template
  description: Events for a specific player
  vars:
    - name: player_id
      kind: string
      description: Player UUID
```
This generates a builder:
```go
func PlayerEventsSubject(playerId string) NatsSubject { /* ... */ }
```
Notes:
- Vars appear in the order listed.
- Parameter names use lower camel case (player_id -> playerId).
- A template with no vars still gets a helper returning a fixed subject.

## What gets generated

### constants mode
- Types: distinct string types per group.
- Constants: one exported identifier per entry using the `name`.
- Slices: All<AnnotationKeys|LabelKeys|...> containing all values of that group.
- Subject helpers: `<Base>Subject(...)` for each `subject_templates` entry.

### descriptors mode
Emits descriptor variables for annotations only, named `<Name>Desc`, mapping value_kind to constructor:
- string -> NewStringAnnotationDesc(constant.<Name>)
- boolean -> NewBoolAnnotationDesc(constant.<Name>)
- uuid -> NewUUIDAnnotationDesc(constant.<Name>)
- rfc3339_timestamp -> NewTimeAnnotationDesc(constant.<Name>, time.RFC3339)
- enum:<EnumType> -> NewEnumAnnotationDesc[<EnumType>](constant.<Name>, constant.All<EnumType>s)

Usage with metadata:
```go
// read
v, found, err := metadata.Get(m, metadata.PlayerOnlineDesc)
// write
metadata.Set(m, metadata.PlayerOnlineDesc, true)
```

## Naming rules
- Constant names (UPPER_SNAKE) become exported identifiers (PascalCase): PLAYER_USERNAME -> PlayerUsername.
- Enum value names (UPPER_SNAKE) become exported identifiers: ONLINE -> Online.
- Subject builder name: strip trailing "Template" from constant name, add "Subject" suffix.
- Template parameter names are converted to lower camel case.

## Validation rules
- Unique `name` per constant and per enum value list.
- Allowed `group` values only.
- Non-empty `wire` for constants; enums require non-empty `value`.

## Example (everything together)
```yaml
version: 1
enums:
  - name: PlayerStatus
    description: Lifecycle status
    values:
      - name: ONLINE
        value: online
      - name: OFFLINE
        value: offline
constants:
  - name: PLAYER_USERNAME
    group: annotations
    wire: player/username
    value_kind: string
  - name: PLAYER_STATUS
    group: annotations
    wire: player/status
    value_kind: enum:PlayerStatus
  - name: SERVER_REGION
    group: labels
    wire: server/region
    value_kind: string
  - name: SYSTEM_HEALTH
    group: nats_subjects
    wire: system.health
  - name: PLAYERS_BUCKET
    group: kv_buckets
    wire: players
  - name: PLAYER_EVENTS_TEMPLATE
    group: subject_templates
    wire: player.{player_id}.events
    vars:
      - name: player_id
        kind: string
```

## Tips
- Keep descriptions short; they become doc comments.
- Prefer stable wire values (protocol compatibility).
- Use the generated subject helpers and descriptors everywhere to reduce bugs.

## Troubleshooting
- The generator fails fast and prints `error: ...` on invalid YAML (unknown group, duplicates, empty wire, etc.).
- In descriptors mode, time import is added only when needed by at least one RFC3339 annotation.


