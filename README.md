# Stellaroot

[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go&logoColor=white)](go.mod) [![Bazel](https://img.shields.io/badge/Bazel-Build-43A047?logo=bazel&logoColor=white)](BUILD.bazel) [![License](https://img.shields.io/github/license/bafbi/stellaroot)](LICENSE)

Cloud‑native Minecraft network platform. Go + Bazel + NATS JetStream. Server/player metadata is stored in NATS KV and rendered via Gin + Templ with progressive enhancement (htmx + Alpine.js). Images are built reproducibly with Bazel.

## Quick Features
• Realtime player & server metadata (NATS JetStream KV)
• Server‑rendered templates (Templ) + htmx fragments for dynamic lists
• Fake data seeder (Job or CLI) for local/demo environments
• Bazel‑built distroless images (dashboard, fakedata)
• Kubernetes manifests (kind friendly) under `kubernetes/`

## Fast Start
Clone & run dashboard (needs NATS for live data; otherwise lists will be empty):
```bash
git clone https://github.com/Bafbi/Stellaroot.git
cd Stellaroot

# Optional infra
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2 -js -m 8222

# Dashboard
bazel run //services/dashboard:dashboard
# Visit http://localhost:8080
```

Seed demo data (players/servers):
```bash
bazel run //tools/fakedata:fakedata -- -players 40 -servers 6
```
Or build/load the seeder image and use the provided Kubernetes Job.

## Build Images
```bash
bazel build //services/dashboard:image           # build dashboard image
bazel run   //services/dashboard:load            # load to local Docker
bazel run   //tools/fakedata:load                # load fakedata image
```

## Kubernetes (kind)
```bash
kind create cluster --name stellaroot-local --config kubernetes/kind-config.yaml
bazel run //services/dashboard:load_to_kind
bazel run //tools/fakedata:load_to_kind
kubectl apply -k kubernetes/
```
The fakedata Job seeds metadata once; rerun if you clear buckets.

## Config (env)
Dashboard & seeder respect:
`NATS_URL` (default nats://localhost:4222), `NATS_USER`, `NATS_PASSWORD`, `NATS_TOKEN`, `PLAYERS_BUCKET` (players), `SERVERS_BUCKET` (servers)
Seeder extras: `FAKER_PLAYERS`, `FAKER_SERVERS`, `FAKER_PREFIX`, `FAKER_UPDATES`, `FAKER_INTERVAL`, `FAKER_SEED`

## API (Dashboard)
Pages: `/`, `/players`, `/servers`
JSON: `/api/players`, `/api/servers`, update endpoints: `/api/players/:uuid/update`, `/api/servers/:name/update`
Fragments (htmx): `/players/fragment`, `/servers/fragment`

## Layout
```
libs/        shared (metadata, schema, proto stubs)
services/    dashboard, permission (WIP)
tools/       fakedata, build helpers
kubernetes/  cluster manifests (nats, services, job)
```

## Contributing
Small, focused PRs. Ensure `bazel build //...` passes. Open an issue for larger proposals.

## License
See [LICENSE](LICENSE).