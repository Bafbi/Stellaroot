# Stellaroot

A scalable Minecraft network server built with microservice architecture, designed for deployment on Kubernetes. This project provides a modern, maintainable foundation for running large-scale Minecraft networks.

## Features

- **Microservice Architecture**: Modular services for dashboard, permissions, and metadata management
- **Container-Native**: Built with Docker and Kubernetes in mind
- **Modern Go Stack**: Uses Go 1.24.3 with Gin, NATS, and Templ
- **Development-Friendly**: Easy local development with automatic dev tagging
- **Scalable**: Designed to handle multiple servers and thousands of players

## Quick Start

### Prerequisites

- [Bazel](https://bazel.build/) (build system)
- [Go 1.24.3+](https://golang.org/)
- [Git](https://git-scm.com/)
- [Docker](https://docker.com/) (for local testing)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Bafbi/Stellaroot.git
   cd stellaroot
   ```

2. **Build the project (development mode)**
   ```bash
   ./tools/build-dev.sh
   ```

3. **Run the dashboard service**
   ```bash
   bazel run //services/dashboard:dashboard
   ```

4. **Load images locally for testing**
   ```bash
   bazel run //services/dashboard:load_dashboard_image
   ```

### Development Workflow

#### Local Development (Default)
All builds use development tags by default:

```bash
# Build all services with dev tags (e.g., dev-a1b2c3d)
./tools/build-dev.sh

# Or use Bazel directly
bazel build //...

# Run specific services
bazel run //services/dashboard:dashboard
bazel run //services/permission:permission
```

#### Release Builds
For production releases:

```bash
# Build with version tags (requires clean git tree)
./tools/build-release.sh

# Or use environment variable
RELEASE_MODE=true bazel build --config=release //...

# Push to container registry
bazel run //services/dashboard:push_dashboard_to_ghcr
```

## Project Structure

```
stellaroot/
├── libs/                     # Shared libraries
│   ├── metadata/            # Server and player metadata
│   ├── permission_grpc/     # Permission service gRPC definitions
│   └── schema/              # Shared protobuf schemas
├── services/                # Microservices
│   ├── dashboard/          # Web dashboard for network management
│   └── permission/         # Permission and authorization service
├── tools/                  # Build and development tools
│   ├── build-dev.sh       # Development build script
│   ├── build-release.sh   # Release build script
│   └── bazel_stamp_vars.sh # Version stamping script
└── AGENTS.md              # Development guidelines
```

## Services

### Dashboard Service
Web-based dashboard for managing the Minecraft network:
- **Port**: 8080 (default)
- **Features**: Server status, player management, network overview
- **Tech Stack**: Go + Gin + Templ templates

### Permission Service
Handles player permissions and authorization:
- **Features**: Role-based access control, permission inheritance
- **Tech Stack**: Go + gRPC + Casbin

## Image Tagging Strategy

The project uses smart image tagging based on build mode:

- **Development**: `dev-{git-commit}` (e.g., `dev-a1b2c3d`)
- **Dirty working tree**: `dev-{git-commit}-dirty`
- **Release**: Version from git tag or MODULE.bazel (e.g., `v1.0.0`)

## Configuration

### Environment Variables

- `RELEASE_MODE=false` (default): Development mode with dev tags
- `RELEASE_MODE=true`: Release mode with version tags

### Bazel Configurations

- Default: Development mode
- `--config=release`: Release mode with version tags
- `--config=ci`: CI/CD optimized settings

## Testing

```bash
# Run all tests
bazel test //...

# Run specific service tests
bazel test //services/dashboard:dashboard_test
bazel test //libs/metadata:metadata_test
```

## Deployment

### Local Docker
```bash
# Load image to local Docker daemon
bazel run //services/dashboard:load_dashboard_image

# Run with Docker
docker run -p 8080:8080 stellaroot/dashboard:dev-latest
```

### Container Registry
```bash
# Push to GitHub Container Registry (requires authentication)
bazel run //services/dashboard:push_dashboard_to_ghcr
```

### Kubernetes
(Kubernetes manifests and Helm charts coming soon)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `bazel test //...`
5. Build: `./tools/build-dev.sh`
6. Submit a pull request

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.