# Repository Guidelines

## Project Structure & Module Organization
`cmd/pilotwave/` contains the main Go entrypoint. Core backend logic lives under `pkg/`, grouped by domain such as `auth/`, `cluster_bridge/`, `gateway_manager/`, `http_server/`, `router_manager/`, `security_manager/`, and `user_manager/`. Default runtime configuration is in `configs/config.toml.dist`. Deployment assets live in `manifests/`, `build/docker/`, and `build/helm/pilotwave/`. The Vue 3 + Vite frontend lives in `web/`, with views in `web/src/views/`, reusable components in `web/src/components/`, Vuex modules in `web/src/store/`, and translations in `web/src/plugins/translate/`. API and collection artifacts are under `doc/` and `testcase/`. Project notes such as architecture, frontend notes, AI change guidance, dependency security, and TODO live in `docs/`.

## Build, Test, and Development Commands
Use the repository `Makefile` for standard builds:

- `make build-web` builds the production frontend into `web/dist`.
- `make build-go` runs `go generate ./cmd/pilotwave` and compiles the backend binary.
- `make build-all` builds web, Go, and Docker assets in sequence.
- `cd web && npm run dev` starts the Vite development server.
- `make run-server` runs the backend locally with the cluster bridge disabled for UI/API smoke tests.
- `KUBECONFIG=/path/to/config make run-server-k8s` runs the backend against a cluster.
- `ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio` runs the local Istio smoke suite; this is the default validation context.
- `ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio` runs the browser/API Istio E2E suite against the same local validation context.
- `cd web && npm run lint` runs the frontend ESLint rules.
- `go test ./...` is the quickest backend regression pass.

## Coding Style & Naming Conventions
Follow Go defaults: tabs, `gofmt`, package names in lowercase, and exported identifiers in PascalCase. Keep new backend code inside the existing domain folders. In the frontend, match the current Vue 3 conventions: component files use PascalCase (`RouterDetail.vue`), Vuex modules use singular domain names (`Auth.js`, `Gateway.js`), and translation keys belong in `web/src/plugins/translate/`.

## Testing Guidelines
Prefer `go test ./...` for broad coverage. The checked-in tests in `unit/` use `httpexpect` against a running server on `http://127.0.0.1:22112`, so run the backend first before using `go test ./unit/...` or `cd unit && go test unit_test.go -v`. Add new Go tests as `*_test.go` files near the package they validate.

## Commit & Pull Request Guidelines
Recent history uses short, imperative subjects such as `update gateway` and `add update password for personal`. Keep commit messages concise and scoped to one change. Pull requests should explain the affected area, list required config or cluster assumptions, and include screenshots for `web/` UI changes. Link the relevant issue or task when one exists.

## Security & Configuration Tips
Do not commit live secrets or cluster-specific values into `configs/config.toml`. Keep local databases, build outputs, swap files, and ad hoc binaries out of commits; files like `pilotwave.db`, `tmp/`, and editor artifacts should stay local, so check `git status` before opening a PR.
