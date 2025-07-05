# Agent Development Guidelines

## Build Commands

### Development Mode (Default)
- **Build all (dev)**: `./tools/build-dev.sh` or `bazel build //...`
- **Build specific target**: `bazel build //services/dashboard:dashboard`
- **Run target**: `bazel run //services/dashboard:dashboard`
- **Load image locally**: `bazel run //services/dashboard:load_dashboard_image`

### Release Mode
- **Build release**: `./tools/build-release.sh` or `RELEASE_MODE=true bazel build --config=release //...`
- **Test all**: `bazel test //...`
- **Push to registry**: `bazel run //services/dashboard:push_dashboard_to_ghcr`

### Image Tagging Strategy
- **Development**: Images tagged as `dev-{git-commit}` (e.g., `dev-a1b2c3d`)
- **Release**: Images tagged with version from git tag or MODULE.bazel (e.g., `v1.0.0`)
- **Dirty working tree**: Adds `-dirty` suffix (e.g., `dev-a1b2c3d-dirty`)

### Environment Variables
- `RELEASE_MODE=true`: Enable release mode with version tags
- `RELEASE_MODE=false` (default): Development mode with dev tags

## Code Style Guidelines
- **Language**: Go 1.24.3
- **Import grouping**: Standard library, external deps, internal packages (separated by blank lines)
- **Naming**: PascalCase for exported, camelCase for unexported, ALL_CAPS for constants
- **Error handling**: Return errors explicitly, wrap with `fmt.Errorf()` for context
- **Struct tags**: Use JSON tags for API structs: `json:"field_name"`
- **Logging**: Use `log/slog` with structured logging: `logger.Info("message", "key", value)`
- **Dependencies**: Gin for HTTP, NATS for messaging, Templ for HTML templates
- **Package structure**: `libs/` for shared code, `services/` for applications
- **Context**: Always pass `context.Context` as first parameter for cancellation
- **Concurrency**: Use `sync.RWMutex` for read-heavy operations, channels for communication
- **Build files**: Use Bazel BUILD.bazel files with `go_library`, `go_binary` targets