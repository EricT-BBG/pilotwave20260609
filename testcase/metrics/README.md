# Metrics Smoke Check

Use this lightweight check against a locally running Pilotwave server to confirm
that `/metrics` is scrapeable without an auth token and includes build and HTTP
request metrics.

Start the backend in another terminal:

```sh
make run-server
```

Run the smoke check:

```sh
testcase/metrics/smoke.sh
```

Override the target URL when needed:

```sh
PILOTWAVE_BASE_URL=http://127.0.0.1:22112 testcase/metrics/smoke.sh
```

The script curls `/metrics` without credentials and checks for:

- `pilotwave_build_info`
- `pilotwave_http_requests_total`
- `pilotwave_http_request_duration_seconds`

Cluster-enabled deployments also export:

- `pilotwave_kubernetes_write_operations_total`
- `pilotwave_kubernetes_write_operation_duration_seconds`
- `pilotwave_kubernetes_write_conflicts_total`
- `pilotwave_istio_resource_info`
- `pilotwave_istio_resource_generation`
- `pilotwave_istio_namespace_injection_state`
- `pilotwave_istio_tls_certificate_not_after_timestamp`
- `pilotwave_istio_tls_certificate_days_until_expiry`
- `pilotwave_istio_tls_certificate_expired`
- `pilotwave_istio_gateway_tls_secret_missing`
- `pilotwave_istio_gateway_tls_secret_invalid`
