# Stellaroot

[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&logoColor=white)](go.mod)
[![Bazel](https://img.shields.io/badge/Bazel-Build-43A047?logo=bazel&logoColor=white)](BUILD.bazel)
[![License](https://img.shields.io/github/license/bafbi/stellaroot)](LICENSE)

A scalable, cloud‑native Minecraft network platform built with Go, NATS JetStream, and Bazel. Stellaroot embraces a microservice architecture, ships as OCI images, and targets smooth local development and Kubernetes deployment.


## Highlights

- Microservices: Dashboard UI, Permissions, and shared libraries
- Realtime metadata: Player and server state via NATS JetStream KV
- Modern Go stack: Gin, Templ, Casbin, NATS
- Reproducible builds: Bazel rules for binaries and container images
- Kubernetes‑ready: Kind‑friendly and manifests under `kubernetes/`


## Architecture

- Dashboard service (HTTP, Gin + Templ) renders server/player views and exposes JSON APIs
- Metadata library wraps NATS JetStream KV for players/servers with in‑memory caches and watchers
- Permission service uses Casbin with a Redis adapter (WIP wiring for external config)

Data plane components:
- NATS (JetStream + KeyValue) for metadata distribution
- Optional Redis for Casbin policy storage


## Getting Started

Prerequisites:
- Go 1.24+
- Bazel
- Docker (for images and local infra)

Optional (recommended for local dev):
- NATS with JetStream
- Redis (for permissions service)

### Clone

```powershell
git clone https://github.com/Bafbi/Stellaroot.git
cd Stellaroot
```

### Start local infra (optional but useful)

- NATS (JetStream enabled) and a simple monitoring port:
```powershell
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2 -js -m 8222
```

- Redis for Casbin policies:
```powershell
docker run -d --name redis -p 6379:6379 redis:7
```

### Run services with Bazel

- Dashboard (defaults to port 8080):
```powershell
bazel run //services/dashboard:dashboard
```
Browse http://localhost:8080

- Permission service (uses Casbin + Redis; edit connection details in `services/permission/main.go`):
```powershell
bazel run //services/permission:permission
```

Environment for the dashboard metadata client (optional if NATS is non‑default):
```powershell
$env:NATS_URL = "nats://localhost:4222"
# $env:NATS_USER = "..."
# $env:NATS_PASSWORD = "..."
# $env:NATS_TOKEN = "..."
```


## Building Images

The repo includes OCI image rules for the dashboard service.

- Build image:
```powershell
bazel build //services/dashboard:image
```

- Load image into local Docker daemon:
```powershell
bazel run //services/dashboard:load
```
This produces repo tags like `stellaroot/dashboard:latest` (version stamping controlled by `tools/bazel_stamp_vars.sh`).

- Push image to GHCR:
```powershell
bazel run //services/dashboard:push_to_ghcr
```

- Load image into a kind cluster:
```powershell
bazel run //services/dashboard:load_to_kind
```


## Kubernetes (local)

Kind config and manifests live under `kubernetes/`. A quick local walkthrough:

```powershell
# Create cluster
kind create cluster --name stellaroot-local --config kubernetes/kind-config.yaml

# Load the dashboard image
bazel run //services/dashboard:load_to_kind

# Apply base manifests (adjust as the manifests evolve)
kubectl apply -k kubernetes/
```
See `kubernetes/local-cluster.md` for extra notes (hosts entries, cert-manager, etc.).


## Configuration

Metadata client (used by the dashboard) reads environment variables defined in `libs/metadata/config.go`:
- `NATS_URL` (default `nats://localhost:4222`)
- `NATS_USER`, `NATS_PASSWORD`, `NATS_TOKEN`
- `PLAYERS_BUCKET` (default `players`)
- `SERVERS_BUCKET` (default `servers`)

Permission service currently initializes the Casbin Redis adapter in code. Update `services/permission/main.go` with your Redis connection (address/password) and `model.conf` as needed. Future versions will move these to environment variables.


## API Overview (Dashboard)

- GET `/` — Home
- GET `/players` — Players page (Templ)
- GET `/servers` — Servers page (Templ)

JSON APIs:
- GET `/api/players` — List players with labels/annotations
- GET `/api/servers` — List servers with labels/annotations
- POST `/api/players/:uuid/update` — Upsert labels/annotations for a player
- POST `/api/servers/:name/update` — Upsert labels/annotations for a server


## Project Structure

```
stellaroot/
├─ libs/
│  ├─ metadata/          # NATS KV client, caches, watchers
│  ├─ permission_grpc/   # Permission service protobufs (future)
│  └─ schema/            # Shared schema + typed annotation keys
├─ services/
│  ├─ dashboard/         # Web UI (Gin + Templ), APIs, images
│  └─ permission/        # Casbin + Redis adapter (WIP)
├─ kubernetes/           # kind config, kustomize, cert-manager setup
├─ tools/                # Bazel stamping and helpers
└─ BUILD.bazel           # Gazelle + repo build config
```


## Roadmap

- Configurable permissions service (env‑driven Redis/Casbin)
- AuthN/AuthZ for dashboard endpoints
- More dashboards (cluster health, player sessions, server logs)
- Helm chart(s) and end‑to‑end Kubernetes examples
- gRPC/HTTP APIs for game server integrations


## Contributing

- Open an issue to discuss significant changes
- Fork and create a feature branch
- Keep commits scoped and descriptive
- Ensure builds pass: `bazel build //...` and `bazel test //...` (when tests are added)
- Submit a PR


## License

This project is licensed as noted in the [LICENSE](LICENSE) file.