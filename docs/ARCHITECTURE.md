# Pilotwave Architecture

## Overview
Pilotwave is a monolithic control-plane application for managing Istio and Kubernetes routing resources through a Go backend and a Vue 3 SPA frontend. Its main responsibility is not general business CRUD; it acts as a management UI/API over cluster resources such as `VirtualService`, `Gateway`, `AuthorizationPolicy`, and related objects.

## High-Level Structure
- `cmd/pilotwave/`: backend entrypoint. Loads config with `viper`, builds `AppInstance`, then starts the app.
- `pkg/app/instance/`: manual dependency wiring and module lifecycle.
- `pkg/http_server/`: Gin HTTP server, middleware, REST APIs, Swagger, and embedded static assets.
- `pkg/cluster_bridge/`: integration boundary to Kubernetes and Istio clients.
- `pkg/*_manager/`: domain modules for users, routers, gateways, and security resources.
- `pkg/database/`: Gorm v1 database setup for local persistence.
- `web/`: Vue 3 + Vite + Vuex + Vue Router frontend.
- `build/`, `manifests/`: container image, Helm, and raw deployment manifests.

## Backend Runtime Flow
1. `cmd/pilotwave/pilotwave.go` loads config from `--config`, then falls back to `./config.toml` or `./configs/config.toml`.
2. `AppInstance.Init()` creates database, cluster bridge, auth, domain managers, mux manager, and HTTP server.
3. `initDatabase()` connects SQLite or MySQL-compatible DSN.
4. `clusterBridge.Init()` builds Kubernetes and Istio clientsets from `KUBECONFIG` or in-cluster config, unless `cluster.disabled` / `PILOTWAVE_CLUSTER_DISABLED=true` enables local no-cluster mode. `PILOTWAVE_DEV_RESET_ADMIN_PASSWORD=true` is only for local development and resets built-in `admin` login during startup.
5. Emergency admin password reset is handled by `pilotwave admin reset-password`; it initializes only the database path, not Kubernetes/Istio clients.
6. `httpServer.Init()` registers static assets plus `/api/v1/*` routes.
7. `Run()` starts the HTTP server and the connection multiplexer.

`AppInstance` is effectively the service container. Most modules receive `app.App` and pull dependencies through getters instead of constructor injection per dependency.

## Request and Data Flow
Typical request path:

`Vue page -> Vuex action -> axios -> Gin API handler -> cluster bridge or manager -> Kubernetes/Istio or DB`

Two data paths coexist:
- Cluster-backed resources: routers, gateways, auth policies, request auth. These are read/written through `cluster_bridge`, not the local DB.
- Local persisted data: users and Grafana-related settings, stored via Gorm.

This distinction matters. Many APIs look like CRUD, but their source of truth is the cluster, not SQL tables.

Dashboard metrics are a separate cluster-backed read path. The home dashboard loads VirtualService resources across namespaces and uses the selected resource's own namespace for monitoring queries. The global topbar namespace picker remains a list-page filter for resource pages such as Gateway, VirtualService, AuthorizationPolicy, and RequestAuthentication; it intentionally does not control the home dashboard.

## Auth Model
- Sign-in endpoint: `POST /api/v1/auth/signin`
- JWT secret source: `auth.secret` in config
- Auth transport: `Authentication` header or `token` query parameter
- Middleware: `pkg/http_server/middlewares/required_auth.go`

Built-in auth checks bcrypt-hashed local users. There is also an AD/LDAP mode toggled by `auth.method`.

## Frontend Structure
- `web/src/router/`: route table and route guards
- `web/src/store/modules/`: Vuex state/getters
- `web/src/store/modules/actions/`: axios-based API calls
- `web/src/views/`: page-level screens
- `web/src/components/`: reusable UI blocks

The frontend is page-driven and Vuex-driven. API access is duplicated across action files instead of going through a shared typed client. Shared shell actions live in `Template.vue` and `Navigation.vue`: namespace filtering appears in the topbar on resource list pages, while Istio injection, language selection, monitoring source settings, and About live in the left navigation. The topbar exposes direct sign-out.

### Current Istio Resource UI

The UI intentionally uses Istio resource names in visible copy. The legacy internal `router` module maps to Istio `VirtualService`, but navigation, detail titles, topology labels, empty states, and association controls should say VirtualService.

Gateway management supports editing `spec.selector` labels. If a user does not provide selector labels, the backend defaults the Gateway selector to `istio=ingressgateway`; if labels are provided, Pilotwave preserves them as-is. This allows non-default ingress gateway workloads without forcing the old default label.

Gateway and VirtualService detail pages are read-only by default. Users open edit mode from the detail toolbar before changing basic settings, route rules, selector labels, or associations. RequestAuthentication and AuthorizationPolicy detail pages follow the same inspect-first pattern.

Gateway TLS certificate status is metadata-only in the browser. The backend reads the credential secret and returns certificate expiry/status details, but not raw certificate data or private keys. The Helm default places managed Gateway TLS secrets in `istio-system` through `gateway.tlsSecretNamespace`; operators can change that when their ingress gateway workload runs in another namespace.

## Important Design Characteristics
- Manual DI, no framework-managed lifecycle
- Thin-ish handlers, but no strong service-layer boundary
- Domain logic is partly in API handlers and partly in bridge/manager packages
- Legacy backend stack with Gorm v1 and a Vue 3 + Vite frontend migrated from the old Vue CLI app
- Static frontend is embedded/generated for backend serving
- Image workflows default to Docker Buildx but can use Podman through `CONTAINER_RUNTIME=podman`

## AI Modification Guidance
Best targets for safe AI changes:
- single API field additions
- validation or response-shape fixes
- one domain slice at a time, such as `router` or `gateway`
- isolated Vue page or Vuex module updates

High-risk areas:
- `pkg/cluster_bridge/istio_bridge/`
- auth and JWT behavior
- generated static asset flow
- cross-cutting config changes
- visible Istio resource naming, especially legacy internal `router` names leaking into VirtualService UI

Recommended workflow for future AI edits:
1. Pick one vertical slice, for example router detail.
2. Trace the files in order: Vue view -> Vuex action -> API handler -> manager/bridge -> model/config.
3. Keep cluster-backed behavior and DB-backed behavior separate.
4. Verify both backend build and frontend lint/build after changes.

## Practical Notes
- Default config template: `configs/config.toml.dist`
- Local DB paths and generated files should stay uncommitted.
- Existing tests are limited; `unit/` contains API-style tests that expect a running server.
