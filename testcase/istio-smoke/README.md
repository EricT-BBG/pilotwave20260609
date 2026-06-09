# ClusterIP Istio Smoke Validation

This suite validates Pilotwave-manageable Istio features without requiring
NodePort or LoadBalancer access. All traffic checks run from temporary curl pods
inside the cluster and call the ingress gateway by its ClusterIP DNS name.

## Run

```sh
ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio
```

The default context is `colima-legacy-1-18`. Override these when needed:

- `ISTIO_CONTEXT`: Kubernetes context.
- `ISTIO_SMOKE_NAMESPACE`: workload and routing namespace, default `pilotwave-istio-smoke`.
- `ISTIO_SMOKE_INJECTION_NAMESPACE`: namespace label validation target, default `pilotwave-istio-injection-smoke`.
- `ISTIO_SMOKE_CURL_IMAGE`: curl pod image, default `curlimages/curl:8.8.0`.
- `ISTIO_SMOKE_INGRESS_HOST`: ingress gateway service DNS name, default `istio-ingressgateway.istio-system.svc.cluster.local`.

## Coverage

- Namespace injection labels: `istio-injection=enabled`, `istio-injection=disabled`, and revision mode with `istio.io/rev`.
- Gateway: HTTP Gateway routing through the ingress gateway ClusterIP.
- VirtualService: 100% v1, 100% v2, and weighted 75/25 traffic split.
- DestinationRule: subset routing to `version: v1` and `version: v2` workloads.
- Gateway TLS: HTTPS Gateway using a generated temporary Kubernetes TLS secret.
- AuthorizationPolicy: path-specific deny rule returning `403`.
- RequestAuthentication: JWT config plus path-specific policy returning `403` when no JWT is supplied.

## Manual Checks

Apply the baseline resources:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio-smoke/manifests/base.yaml
```

Call through ClusterIP:

```sh
kubectl --context colima-legacy-1-18 -n pilotwave-istio-smoke run curl-v1 --restart=Never --image=curlimages/curl:8.8.0 -- \
  curl -sS -H 'Host: smoke.pilotwave.local' http://istio-ingressgateway.istio-system.svc.cluster.local/
kubectl --context colima-legacy-1-18 -n pilotwave-istio-smoke logs curl-v1
kubectl --context colima-legacy-1-18 -n pilotwave-istio-smoke delete pod curl-v1
```

Expected response:

```text
hello from v1
```

Switch to v2:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio-smoke/manifests/route-v2.yaml
```

Switch to weighted split:

```sh
kubectl --context colima-legacy-1-18 apply -f testcase/istio-smoke/manifests/weighted-v2-75-v1-25.yaml
```

Security policies are intentionally scoped to the ingress gateway in
`istio-system`. The smoke runner deletes the temporary policy resources before
and after the security checks.
