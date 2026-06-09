# Dependency Security Notes

## Current Status
As of the latest recorded dependency refresh, the active application paths have clean dependency security checks:

- Frontend: `npm audit` reports `0 vulnerabilities`.
- Backend: `govulncheck ./cmd/... ./pkg/...` reports `0 reachable vulnerabilities`.
- Full repo verification: `make verify` passes.

## Backend Changes
The Go dependency refresh addressed reachable vulnerabilities reported in Gin, CORS, JWT, and HTTP/2 code paths.

Key updates:

- `github.com/gin-gonic/gin` upgraded to `v1.9.1`.
- `github.com/gin-contrib/cors` upgraded to `v1.6.0`.
- JWT usage moved from `github.com/golang-jwt/jwt` to `github.com/golang-jwt/jwt/v5`.
- `golang.org/x/net` upgraded to `v0.53.0`.
- `github.com/go-logr/logr` is not imported by Pilotwave directly; it is pulled through Kubernetes `k8s.io/klog/v2` and remains pinned to `v0.2.0` because the current `k8s.io/klog/v2@v2.4.0` API is not compatible with `logr` v1.x.

The Go baseline is now `go 1.25.0` because the upgraded `golang.org/x/net` module requires it.

## Kubernetes Client Pin Notes

Do not remove `github.com/go-logr/logr` or upgrade it to `v1.x` as a standalone cleanup. Pilotwave does not import it directly, but the current Kubernetes client stack does:

```text
pkg/cluster_bridge/bridge
k8s.io/apimachinery/pkg/api/meta
k8s.io/klog/v2
github.com/go-logr/logr
```

`k8s.io/klog/v2@v2.4.0` expects the older `logr` API. Testing `github.com/go-logr/logr v1.4.3` caused compile failures in `k8s.io/klog/v2` because `logr.Logger` changed from the older nil-compatible API. Treat this as part of a Kubernetes client upgrade, not as an isolated dependency removal.

## Frontend Changes
The frontend dependency refresh cleaned the Vite/Vitest and ESLint-related audit findings.

Key updates:

- `vitest` upgraded to `4.1.6`.
- `eslint-plugin-vue` upgraded to `10.9.1`.
- `vue-eslint-parser` added at `10.4.0`.
- `overrides` pins `ajv` to `6.14.0` and `brace-expansion` to `1.1.14`.

## Verification Commands
Use these before committing dependency changes:

```sh
make verify
cd web && npm audit
go run golang.org/x/vuln/cmd/govulncheck@latest ./cmd/... ./pkg/...
```

Avoid running `govulncheck ./...` until repo-local noise is cleaned up; `unit/` contains mixed Go package names, and `web/node_modules` may include non-project Go examples.
