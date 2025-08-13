# Agent Development Guidelines

Updated: 2025-08-13

## Build & Run

### Common
- Build everything: `bazel build //...`
- Test everything: `bazel test //...`
- Run dashboard service: `bazel run //services/dashboard:dashboard`

### Dashboard Image Targets
- Build image: `bazel build //services/dashboard:image`
- Load image into local Docker daemon: `bazel run //services/dashboard:load`
- (Kind helper script) Load into kind cluster: `bazel run //services/dashboard:load_to_kind`
- Push to GHCR: `bazel run //services/dashboard:push_to_ghcr`

### Fake Data (Seeder) Targets
- Run CLI locally: `bazel run //tools/fakedata:fakedata -- -players 50 -servers 5`
- Build image: `bazel build //tools/fakedata:image`
- Load image to Docker: `bazel run //tools/fakedata:load`
- Load image to kind: `bazel run //tools/fakedata:load_to_kind`
- Push to GHCR: `bazel run //tools/fakedata:push_to_ghcr`

### Release Mode
- Release build: `RELEASE_MODE=true bazel build --config=release //...`

## Frontend (Templ + htmx + Alpine)
- Server‑side templates use [Templ]; Bazel `genrule` (`//services/dashboard/templates:templ_generated_files`) invokes `templ generate` per *.templ file and a `go_library` wraps generated code.
- View models for templates live in `services/dashboard/templates/viewmodels.go` (PlayerViewModel, ServerViewModel) and are shared between fragment handlers and (optionally) JSON.
- Progressive enhancement: htmx is now used for Players/Servers list hydration & refresh (endpoints `/players/fragment`, `/servers/fragment`). Alpine.js still powers modal state & editing UX.
- Migration approach: move CRUD modals and actions to htmx incrementally (hx-post to update endpoints returning row partials + toast). Keep Alpine only for transient UI state until replaced.
- Fragments: `PlayersFragment`, `ServersFragment` + `PlayerRow` / `ServerRow` Templ partials return `<tr>` rows consumed by htmx.

### Adding a New Fragment + htmx Flow
1. Add view model fields (or reuse existing) in `viewmodels.go`.
2. Create Templ partial(s) returning the minimal HTML snippet.
3. Add handler in `main.go` that builds view models and renders the fragment.
4. Point a container/element with `hx-get` + `hx-trigger="load"` (and optionally a refresh button) at the new route.

## Backend (Dashboard Service)
- Gin router in `services/dashboard/main.go` serves pages & JSON APIs under `/api/*`.
- New fragment routes: `/players/fragment`, `/servers/fragment` (HTML).
- JSON APIs retained: `/api/players`, `/api/servers`, update endpoints for mutations.

## Metadata Library
- Backed by NATS JetStream KV (players & servers buckets). Client caches + watcher pattern.
- Canonical player name annotation key: `player/name` (see `libs/schema`). Frontend & fakedata align on this.

## Fake Data Generator (tools/fakedata)
- Seeds players & servers once by default; optional periodic updates with `-updates` flag.
- Environment variable overrides (all optional):
	- `FAKER_PLAYERS`, `FAKER_SERVERS`, `FAKER_PREFIX`, `FAKER_UPDATES`, `FAKER_INTERVAL`, `FAKER_SEED`.
- Uses same NATS & bucket env vars as dashboard (see below).
- Kubernetes Job manifest: `kubernetes/deployment/services/fakedata/job.yaml` (added to services aggregator for seeding run).

## Kubernetes Manifests
Structure under `kubernetes/deployment/`:
- `deployment.yaml` top-level orchestrator includes `services`, `nats`, `kgateway`, `cert-manager` sets.
- `services/deployment.yaml` lists namespace + individual service/job paths (now includes `fakedata` then `dashboard`).
- Images resolved via `images/<target>.yaml` mapping logical image names (e.g. `stellaroot/dashboard`, `stellaroot/fakedata`) to concrete registry references.
- Add a new service/job: create folder under `services/` with manifest(s), append `- path: <name>` to `services/deployment.yaml`, add image mapping entry, ensure namespace present in `configs/common.yaml`.

## Environment Variables
Shared (dashboard & fakedata):
- `PORT` (dashboard only, default 8080)
- `NATS_URL`, `NATS_USER`, `NATS_PASSWORD`, `NATS_TOKEN`
- `PLAYERS_BUCKET` (default `players`)
- `SERVERS_BUCKET` (default `servers`)

Fakedata specific (also flags):
- `FAKER_PLAYERS`, `FAKER_SERVERS`, `FAKER_PREFIX`, `FAKER_UPDATES`, `FAKER_INTERVAL`, `FAKER_SEED`

## Code Style Guidelines
- Language: Go 1.24.3 with Bazel build system
- Imports: stdlib, blank line, external, blank line, internal
- Naming: PascalCase exports; camelCase unexported; ALL_CAPS constants
- Errors: Wrap with `fmt.Errorf("context: %w", err)`
- JSON tags: snake_case (`json:"field_name"`)
- Logging: `log/slog` structured key/value
- Context: First param when cancellation/timeout needed
- Dependencies: Gin (HTTP), NATS (messaging), Templ (templates), Casbin (auth), htmx (HTML over the wire), Alpine.js (transitional state mgmt)
- Repo layout: `libs/` shared libraries, `services/` binaries, `tools/` utilities, Bazel `BUILD.bazel` in each package

## Adding a New Executable + Image Quick Recipe
1. Create Go package + `go_binary`.
2. Define mtree spec/mutate + tar (see dashboard or fakedata BUILD files).
3. Add `oci_image` target (base usually `@distroless_base`).
4. Add labels/tags templates (`expand_template`).
5. Add `oci_push` + `oci_load` + optional kind helper script.
6. Map image in `kubernetes/deployment/images/<env>.yaml` & reference via `{{ images.get_image('logical/name') }}` in manifest.

## Frontend Migration Notes
- Current: Lists via htmx fragments, modals via Alpine.
- Next steps (optional): Move edit forms to htmx (inline row expansion or modal partial fetch), add SSE or polling triggers (`hx-trigger="every 5s"`) for live updates, unify toast handling server‑side (return snippet + `hx-swap-oob`).

## Troubleshooting
- Templ generation issues: build `//services/dashboard/templates:templ_generated_files` with `--sandbox_debug` to inspect logs.
- Missing fragment data: verify fragment handlers registered and view models exported in `viewmodels.go`.
- Image not updating in cluster: ensure `bazel run :load` executed and (for kind) `:load_to_kind` script run; confirm mapping in images file.

## Glossary
- Fragment: HTML snippet returned for partial page updates via htmx.
- ViewModel: Struct optimized for rendering layer (avoid leaking internal structs directly).
- Logical Image Name: Key used in images mapping (e.g. `stellaroot/dashboard`).

---
This document should stay concise. When adding new capabilities, extend the relevant section instead of duplicating instructions.
