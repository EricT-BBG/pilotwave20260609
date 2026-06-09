# TODO

## Istio Compatibility & Safe Updates

### Current Status
- [x] Gateway, VirtualService, DestinationRule, AuthorizationPolicy, and
  RequestAuthentication writes use patch-first updates instead of full object
  `Update(...)` calls.
- [x] Web E2E covers stale `resourceVersion` conflicts, preservation checks for
  Gateway, VirtualService, DestinationRule, AuthorizationPolicy, and
  RequestAuthentication, namespace injection labels including revision mode, and
  Gateway TLS replacement.
- [x] Web E2E also covers Gateway server/port add-remove rules, Gateway list
  sorting with stable row numbers, namespace menu status/refresh persistence,
  and the global API-unavailable alert.
- [x] Helm local-image E2E render coverage is available through
  `make e2e-helm-local-image`; it validates local image values, existing
  production PVC wiring, ServiceMonitor/PodMonitor rendering, and the
  `istio.required=true` missing-CRD failure message.
- [x] Istio resource access is routed through the first `IstioResourceClient`
  adapter slice inside `pkg/cluster_bridge/istio_bridge/`.
- [x] Current local validation target is `colima-legacy-1-18` / Istio 1.7.x.
  Validation passed there with
  `make smoke-istio` and `make e2e-web-istio`.

### High Priority
- [x] Keep `colima-legacy-1-18` as the default Kubernetes/Istio context for local
  validation unless another context is explicitly requested.
- [ ] Split `IstioResourceClient` into smaller domain interfaces such as
  Gateway, Router, and Security clients so tests can mock narrower seams.
- [ ] Add backend unit tests for namespace injection validation and patching:
  `enabled`, `disabled`, `revision`, missing revision, and invalid revision.
- [ ] Add Go API tests for stale `resourceVersion` conflict mapping on Gateway
  and Router handlers.

### Medium Priority
- [ ] Keep Istio typed structs for building and validating managed fragments while
  keeping write behavior in adapter code.
- [ ] Evaluate a Kubernetes client stack upgrade before changing
  `github.com/go-logr/logr`; the current `k8s.io/klog/v2@v2.4.0` dependency
  requires the old `logr v0.2.0` API.
- [ ] Optional future compatibility work: add a separate current-Istio test target,
  then run the same smoke and E2E suites against it. Do not substitute that
  target for `colima-legacy-1-18` in local validation notes.

### Compatibility Track
- Default direction: single binary with dynamic client patch writes and runtime CRD discovery.
- Optional later direction: multi-binary builds for pinned Istio client versions, for example `pilotwave-istio175` and `pilotwave-istio-latest`.
- Even with multi-binary builds, keep patch-first writes to avoid overwriting custom YAML.

### Validation
- [x] Default local validation commands:
  `ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio` and
  `ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio`.
- [x] Helm render E2E command:
  `make e2e-helm-local-image`.
- [ ] If a future current-Istio target is added, run the same commands against it as
  an extra compatibility check, not as a replacement for `colima-legacy-1-18`.
- [x] Smoke expectations: HTTP routing, weighted VirtualService routing, Gateway
  TLS, AuthorizationPolicy deny, and RequestAuthentication missing-JWT deny.
- [x] E2E expectations: UI CRUD, namespace dropdown refresh, stale update conflicts,
  safe namespace injection labels, preservation checks, and TLS replacement
  checks.
- [x] Additional E2E expectations: Gateway server/port editing, Gateway list
  sort order, namespace menu injected-status persistence, API-unavailable alert,
  Helm local-image render values, existing PVC reuse, monitoring render output,
  and missing Istio CRD failure messaging.
- [x] Preservation checks must confirm Gateway, VirtualService, DestinationRule,
  AuthorizationPolicy, and RequestAuthentication fields outside Pilotwave's
  managed fragments survive Pilotwave edits.
- [x] Namespace injection E2E must stay scoped to dedicated test namespaces. Verify
  legacy `enabled` and `disabled` modes plus revision mode by reading
  `istio-injection` and `istio.io/rev` labels with kubectl.
