# AGENTS.md (Quick Reference)
1. Build all: `bazel build //...`  Release: `RELEASE_MODE=true bazel build --config=release //...`
2. Test all: `bazel test //...`  Single pkg: `bazel test //libs/metadata:go_default_test`  Single file/func: use `--test_filter='TestName'`.
3. Run dashboard: `bazel run //services/dashboard:dashboard`  Fakedata seeder: `bazel run //tools/fakedata:fakedata -- -players 50 -servers 5`.
4. Images (dashboard): build `bazel build //services/dashboard:image`; load local `bazel run //services/dashboard:load`; kind `:load_to_kind`; push `:push_to_ghcr` (same pattern for fakedata).
5. Style: Go 1.24.3; gofmt/goimports (imports grouped: stdlib, blank, external, blank, internal).
6. Naming: PascalCase exports; camelCase internals; ALL_CAPS consts; JSON tags snake_case (`json:"field_name"`).
7. Errors: wrap with context `fmt.Errorf("action detail: %w", err)`; sentinel errors via `errors.Is`/`errors.As`.
8. Logging: use `log/slog` structured (`slog.Info("msg", "key", val)`). Avoid fmt prints in prod code.
9. Context: accept `ctx context.Context` first where cancellation/timeout or I/O occurs; never store contexts.
10. Concurrency: prefer channels or contexts over global vars; protect shared state with mutex; avoid goroutine leaks (tie to ctx).
11. Templates: Templ fragments (`PlayersFragment`, `ServersFragment`, row partials) served at `/players/fragment`, `/servers/fragment`; htmx for hydration; Alpine only for transient UI.
12. Metadata: NATS JetStream KV buckets (`players`, `servers`); canonical annotation key `player/name`.
13. Env vars shared: `NATS_URL`, `NATS_USER`, `NATS_PASSWORD`, `NATS_TOKEN`, `PLAYERS_BUCKET`, `SERVERS_BUCKET`, `PORT` (dashboard), fakedata extras (`FAKER_*`).
14. Add fragment: view model field -> partial templ -> handler in `services/dashboard/main.go` -> element with `hx-get` + `hx-trigger="load"`.
15. Add executable+image: create package + `go_binary`; replicate image rules (see dashboard BUILD) -> add mapping in `kubernetes/deployment/images/*.yaml`.
16. Kubernetes deploy composition: root `kubernetes/deployment/deployment.yaml` aggregates `services`, `nats`, `kgateway`, `cert-manager`.
17. Testing tips: fast filter `bazel test //path:target --test_filter='TestThing'`; race (if enabled) add `--@io_bazel_rules_go//go/config:race` (if configured).
18. Lint (implicit): rely on compiler + formatting; keep functions < ~60 lines; avoid cyclic deps; keep packages focused.
19. Security: never commit secrets; prefer env vars; validate external input early; sanitize log values.
20. Troubleshooting: template gen `bazel build //services/dashboard/templates:templ_generated_files`; image mismatch -> reload image + check mapping; missing fragment -> handler + view model export.
