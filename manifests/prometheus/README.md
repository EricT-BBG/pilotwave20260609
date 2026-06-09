# Pilotwave Prometheus Alerts

This directory contains Prometheus Operator alert rules for Pilotwave application
metrics.

## Apply

Install the Prometheus Operator CRDs before applying this manifest. The default
manifest creates a `PrometheusRule` in the `pilotwave` namespace:

```sh
kubectl apply -f deploy/prometheus/pilotwave-rules.yaml
```

If your Prometheus instance selects rules with extra labels, add those labels to
`metadata.labels` before applying. For example, some `kube-prometheus-stack`
installs require a `release` label that matches the Helm release name.

For Helm installs, prefer the chart toggle instead of editing this raw manifest:

```yaml
prometheusRule:
  enabled: true
  labels:
    release: legacy-monitoring
```

## Alerts

- `PilotwaveIstioTLSCertificateExpiringSoon`: certificate expiry is below 30 days.
- `PilotwaveIstioTLSCertificateExpired`: exported certificate state is expired.
- `PilotwaveIstioGatewayTLSSecretMissing`: an Istio Gateway TLS secret cannot be resolved.
- `PilotwaveIstioGatewayTLSSecretInvalid`: an Istio Gateway TLS secret is present but invalid.
- `PilotwaveKubernetesWriteConflicts`: more than 5 write conflicts for a resource and verb in 10 minutes.
- `PilotwaveHighHTTP5xxRate`: more than 5% of HTTP requests are 5xx responses with meaningful traffic.
- `PilotwaveHighP95Latency`: global HTTP p95 latency is above 1 second with meaningful traffic.

## Metric Dependencies

Pilotwave must expose `/metrics` and Prometheus must scrape these metrics:

- `pilotwave_http_requests_total`
- `pilotwave_http_request_duration_seconds_bucket`
- `pilotwave_kubernetes_write_conflicts_total`
- `pilotwave_istio_tls_certificate_days_until_expiry`
- `pilotwave_istio_tls_certificate_expired`
- `pilotwave_istio_gateway_tls_secret_missing`
- `pilotwave_istio_gateway_tls_secret_invalid`

Cluster-enabled Pilotwave runs are required for the Kubernetes write and Istio TLS
metrics. Local UI/API-only runs may expose only the HTTP metrics.
