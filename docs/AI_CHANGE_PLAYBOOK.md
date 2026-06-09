# AI Change Playbook

## Purpose
Use this playbook when asking AI to modify Pilotwave. The project is old, stateful, and split across a Go backend, Vue 3 + Vite frontend, and Kubernetes/Istio integration. Good results come from keeping each change narrow and tracing one vertical slice at a time.

## Core Rule
Do not ask AI to "modernize everything" or "refactor the whole repo" in one step. For this codebase, the safe unit of work is:

`one feature area -> one request path -> one verification pass`

For longer-lived follow-up work, start with `TODO.md`. It tracks the current Istio compatibility and safe-update backlog.

Examples:
- router detail page field change
- gateway create API validation fix
- auth policy response shape adjustment
- user password flow update

## How To Scope A Change
When opening work, specify:
- feature area: `router`, `gateway`, `auth`, `user`, `security`
- target behavior: what should change
- non-goals: what must stay untouched
- verification: what command or screen should prove it works

Good prompt example:

> Update router detail so the API returns and displays `hosts` consistently. Only touch router-related backend/frontend code. Do not refactor auth or gateway logic. Verify with backend build and frontend lint.

## File Trace Order
For most changes, trace files in this order:

1. `web/src/views/` or `web/src/components/`
2. `web/src/store/modules/actions/`
3. `pkg/http_server/api/`
4. `pkg/*_manager/` or `pkg/cluster_bridge/`
5. config, models, or generated assets only if required

This prevents AI from editing deep infrastructure first and breaking unrelated flows.

## Common Change Patterns

### 1. Add or adjust a UI field
- Start at the Vue view/component.
- Find the Vuex action that loads or submits the data.
- Update the matching backend request/response shape.
- Check whether the source of truth is cluster-backed or DB-backed.

### 2. Change API validation or response format
- Start in `pkg/http_server/api/*.go`.
- Keep handler validation and response changes local.
- Only move into manager or bridge code if the data contract truly changes.

### 3. Change cluster resource behavior
- Trace from API handler to `pkg/cluster_bridge/bridge/` and then `pkg/cluster_bridge/istio_bridge/` or `k8s_bridge/`.
- Treat these files as high risk.
- Do not add Kubernetes or Istio full object `Update(...)` writes unless the change proves unknown schema and user-managed YAML fields are preserved.
- Prefer patch-first writes for Pilotwave-managed fields.
- Avoid opportunistic cleanup while changing live resource logic.

### 4. Change user or auth behavior
- Trace `pkg/http_server/api/auth.go`
- Then `pkg/auth/authenticator/`
- Then `pkg/user_manager/` and `pkg/database/` if persistence changes
- Re-check JWT, bcrypt, and config assumptions before editing

## High-Risk Areas
Use extra care in these areas:
- `pkg/cluster_bridge/istio_bridge/`
- `pkg/http_server/middlewares/required_auth.go`
- generated static asset flow under `pkg/http_server/static/`
- config handling with `viper`
- old runtime patterns: Gorm v1, historical Vuetify-style template assumptions, duplicated Axios calls in Vuex actions
- dependency baseline: Go 1.25 is currently required because `golang.org/x/net` was upgraded to clear reachable HTTP/2 vulnerabilities

## Verification Checklist
After each change, prefer this order:

1. `make verify`
2. `cd web && npm run lint`
3. `cd web && npm test`
4. `cd web && npm run build`
5. `go run golang.org/x/vuln/cmd/govulncheck@latest ./cmd/... ./pkg/...` for dependency or backend security changes

Use narrower checks when the environment is incomplete, but always state what was and was not verified.

## Current Development Notes
- `make run-server` is an alias for `make run-server-cluster`; it builds frontend assets, regenerates embedded static assets, and starts the Go backend against `KUBE_CONTEXT=colima-legacy-1-18` by default.
- Use `make run-server-mock` for UI-only smoke tests with `PILOTWAVE_CLUSTER_DISABLED=true`.
- Override `KUBE_CONTEXT` only when testing another real cluster.
- A newly initialized SQLite DB creates `admin` / `admin` for built-in auth.
- Frontend `npm audit` is currently clean after upgrading Vitest and the ESLint Vue parser chain.
- Backend reachable Go vulnerabilities are currently clean after upgrading Gin, CORS, JWT, and `golang.org/x/net`.

## Working Rules For AI
- Change one domain slice at a time.
- Do not mix router, gateway, auth, and security refactors in one patch unless requested.
- Preserve existing request paths and response keys unless the task explicitly changes them.
- Distinguish cluster-backed resources from local DB-backed data before editing.
- Do not treat old code as wrong just because it is old; preserve behavior first.
- If the repo is dirty, avoid reverting unrelated files.

## Recommended Task Types
Good first tasks for AI:
- add a missing field
- fix a response inconsistency
- tighten validation
- improve one page's UX copy or form handling
- document one module's behavior

Bad first tasks for AI:
- replace the frontend state or UI stack in one pass
- replace Gorm or auth stack broadly
- redesign all routing or policy APIs together
- refactor cluster bridge without a failing case or narrow target

## Handoff Template
Use this template when opening a new AI task:

```md
Area: router
Goal: Show and update `hosts` correctly on router detail page.
Source of truth: cluster-backed via Istio
Touch only:
- web/src/views/router/RouterDetail.vue
- web/src/store/modules/actions/Router.js
- pkg/http_server/api/router.go
- pkg/cluster_bridge/...
Do not touch:
- auth
- gateway
- helm/deploy files
Verify:
- make build-go
- cd web && npm run lint
```
