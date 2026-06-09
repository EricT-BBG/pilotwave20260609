# Istio Smoke Fixtures

These manifests support manual and automated validation for Pilotwave Gateway,
VirtualService, DestinationRule, AuthorizationPolicy, RequestAuthentication, and
namespace injection behavior.

## Layout

- `hello-routing/base/manifest.yaml`: creates the demo namespace, v1/v2 workloads, Service, Gateway, DestinationRule, and default v1 VirtualService.
- `hello-routing/route-v2/virtualservice.yaml`: replaces the VirtualService with a 100% v2 route.
- `hello-routing/weighted-v2-75-v1-25/virtualservice.yaml`: replaces the VirtualService with a 75% v2 and 25% v1 weighted route.
- `gateway-tls/manifest.yaml`: adds an HTTPS Gateway and TLS secret for `hello-tls.pilotwave.local`.
- `security/authz-deny-path/authorizationpolicy.yaml`: denies only `/deny` on the ingress gateway.
- `security/requestauth-require-jwt/manifest.yaml`: adds RequestAuthentication and denies `/jwt` when no matching JWT principal is present.

## Hello Routing

Run the full ClusterIP-only smoke suite against the local validation target:

```sh
ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio
```

`colima-legacy-1-18` is the current repo validation cluster. A separate
current-Istio context may be added later for extra compatibility checks, but it
does not replace this context.

The smoke suite covers HTTP Gateway routing, VirtualService gateway mapping,
DestinationRule subset weighting, Gateway TLS, AuthorizationPolicy, and
RequestAuthentication. It leaves the demo namespace available and resets the
default route back to `v1`.

Expected smoke result:

- Local Istio target, currently `colima-legacy-1-18` / Istio 1.7.x, passes.
- Gateway TLS returns `hello from v1` through HTTPS.
- AuthorizationPolicy `/deny` returns `403`.
- RequestAuthentication `/jwt` without JWT returns `403`.

Apply the baseline fixture to create a small HTTP service with `v1` and `v2`
workloads plus an Istio Gateway, VirtualService, and DestinationRule:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio/hello-routing/base/manifest.yaml
```

The default route sends all traffic to subset `v1`.

```sh
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo run curl-v1 --restart=Never --image=curlimages/curl:8.8.0 -- \
  curl -sS -H 'Host: hello.pilotwave.local' http://istio-ingressgateway.istio-system.svc.cluster.local/
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo logs curl-v1
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo delete pod curl-v1
```

Expected response:

```text
hello from v1
```

Switch the route to subset `v2`:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio/hello-routing/route-v2/virtualservice.yaml
```

Expected response:

```text
hello from v2
```

Switch the route to weighted traffic, 75% `v2` and 25% `v1`:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio/hello-routing/weighted-v2-75-v1-25/virtualservice.yaml
```

Run multiple requests to confirm both responses appear. The exact count can vary
on small samples, but `v2` should dominate over a larger run.

```sh
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo run curl-weighted --restart=Never --image=curlimages/curl:8.8.0 -- \
  sh -c "for i in \$(seq 1 40); do curl -s -H 'Host: hello.pilotwave.local' http://istio-ingressgateway.istio-system.svc.cluster.local/; done | sort | uniq -c"
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo logs curl-weighted
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo delete pod curl-weighted
```

Observed on `colima-legacy-1-18`: a 40-request weighted sample returned
`10 hello from v1` and `30 hello from v2`.

## Gateway TLS

Apply the TLS fixture:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio/gateway-tls/manifest.yaml
```

Verify from inside the cluster through the ingress gateway ClusterIP service:

```sh
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo run curl-tls --restart=Never --image=curlimages/curl:8.8.0 -- \
  curl -k -sS --connect-to hello-tls.pilotwave.local:443:istio-ingressgateway.istio-system.svc.cluster.local:443 https://hello-tls.pilotwave.local/
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo logs curl-tls
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo delete pod curl-tls
```

Expected response:

```text
hello from v1
```

Observed on `colima-legacy-1-18`: HTTPS through the ingress gateway ClusterIP
returned `hello from v1`.

TLS replacement validation for the remaining roadmap:

- Replace a Pilotwave-managed Gateway certificate and confirm only
  `pilotwave-*` managed secrets created for the previous certificate are
  removed.
- Confirm user-managed TLS secrets referenced by Gateway servers remain present.
- Confirm the Gateway still serves the replacement certificate and routes to the
  expected backend after the update.

## Security Policies

Apply an AuthorizationPolicy that denies only `/deny`:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio/security/authz-deny-path/authorizationpolicy.yaml
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo run curl-deny --restart=Never --image=curlimages/curl:8.8.0 -- \
  curl -sS -o /tmp/body -w '%{http_code}' -H 'Host: hello.pilotwave.local' http://istio-ingressgateway.istio-system.svc.cluster.local/deny
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo logs curl-deny
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo delete pod curl-deny
kubectl --context colima-legacy-1-18 -n istio-system delete authorizationpolicy pilotwave-deny-hello-path
```

Expected status: `403`.

Observed on `colima-legacy-1-18`: `/deny` returned `403`.

Apply RequestAuthentication plus an AuthorizationPolicy that requires a matching
JWT principal only on `/jwt`:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio/security/requestauth-require-jwt/manifest.yaml
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo run curl-jwt-missing --restart=Never --image=curlimages/curl:8.8.0 -- \
  curl -sS -o /tmp/body -w '%{http_code}' -H 'Host: hello.pilotwave.local' http://istio-ingressgateway.istio-system.svc.cluster.local/jwt
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo logs curl-jwt-missing
kubectl --context colima-legacy-1-18 -n pilotwave-istio-demo delete pod curl-jwt-missing
kubectl --context colima-legacy-1-18 -n istio-system delete authorizationpolicy pilotwave-require-jwt-path
kubectl --context colima-legacy-1-18 -n istio-system delete requestauthentication pilotwave-hello-jwt
```

Expected status without JWT: `403`.

Observed on `colima-legacy-1-18`: `/jwt` without JWT returned `403`.

The sample namespace disables workload sidecar injection on purpose. The smoke
test validates ingress gateway routing and Istio config behavior without being
blocked by legacy Istio sidecar init-container compatibility.

## Web E2E Namespace Injection Checks

Run the browser-driven Istio UI checks against the local validation target:

```sh
ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio
```

The repo currently has three E2E/smoke entrypoints for this area:

- `ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio`: ClusterIP-only Istio
  fixture validation.
- `ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio`: browser/API Istio UI
  validation against the selected cluster context.
- `make e2e-helm-local-image`: Helm render E2E for local image values,
  existing production PVC reuse, monitoring resources, and
  `istio.required=true` missing-CRD failure messaging.

The web E2E creates dedicated namespaces by default:

- `pilotwave-e2e-istio` for Gateway, VirtualService, AuthorizationPolicy, and RequestAuthentication CRUD.
- `pilotwave-e2e-injection` for namespace injection label management.

Current E2E expectations:

- Reject sensitive namespace targets such as `default`, `kube-system`, and `istio-system`.
- Verify namespace dropdown and refresh paths list every cluster namespace.
- Create Gateway, VirtualService, AuthorizationPolicy, and RequestAuthentication through the UI.
- Confirm stale Gateway and VirtualService updates return conflict responses.
- Confirm Gateway and VirtualService metadata/custom fields outside the edited
  fragment survive Pilotwave edits.
- Confirm DestinationRule metadata, `exportTo`, and `trafficPolicy` survive
  Pilotwave-managed subset edits.
- Confirm AuthorizationPolicy and RequestAuthentication metadata survive
  Pilotwave-managed updates.
- Confirm Gateway TLS certificate replacement deletes only Pilotwave-managed
  secrets and preserves user-managed credentialName secrets.
- Add and remove Gateway servers/ports while keeping at least one server.
- Sort Gateway list rows while keeping visible row numbers sequential.
- Delete Gateway and VirtualService through the UI.
- Set legacy injection modes `enabled` and `disabled`, then verify
  `istio-injection` and `istio.io/rev` labels with kubectl.
- Set revision injection mode on a dedicated namespace, then verify
  `istio.io/rev=<test-revision>` is set and `istio-injection` is cleared.
- Show injected namespace status in the namespace menu, refresh namespace data,
  and keep the selected namespace visible.
- Show the global API-unavailable alert when the namespaces API request receives
  no response.
- Restore previous labels for a pre-existing injection test namespace, or delete
  the dedicated namespace if the E2E created it.

Revision-based injection is label-only on `colima-legacy-1-18`: the test proves
Pilotwave writes the namespace labels correctly, but the legacy cluster does not
provide a matching revisioned injector for sidecar admission.
