# Agent Development Guidelines

## Build Commands

### Development Mode (Default)
- **Build all**: `./tools/build-dev.sh` or `bazel build //...`
- **Build target**: `bazel build //services/dashboard:dashboard`
- **Run target**: `bazel run //services/dashboard:dashboard`
- **Test all**: `bazel test //...`
- **Test specific**: `bazel test //libs/metadata:metadata_test`

### Release Mode
- **Build release**: `./tools/build-release.sh` or `RELEASE_MODE=true bazel build --config=release //...`
- **Push image**: `bazel run //services/dashboard:push_dashboard_to_ghcr`

### Local Development
- **Load to Docker**: `bazel run //services/dashboard:load_dashboard_image`
- **Load to Kind**: `bazel run //services/dashboard:load_dashboard_to_kind`

## Code Style Guidelines
- **Language**: Go 1.24.3 with Bazel build system
- **Imports**: Group stdlib, external, internal (blank line separated)
- **Naming**: PascalCase exports, camelCase unexported, ALL_CAPS constants
- **Error handling**: Return explicit errors, wrap with `fmt.Errorf("context: %w", err)`
- **JSON tags**: Use snake_case: `json:"field_name"`
- **Logging**: Structured with `log/slog`: `logger.Info("msg", "key", value)`
- **Context**: First parameter in functions for cancellation/timeouts
- **Dependencies**: Gin (HTTP), NATS (messaging), Templ (templates), Casbin (auth)
- **Structure**: `libs/` shared, `services/` apps, Bazel BUILD.bazel everywhere
