# Pilotwave Grafana Dashboard

Import `pilotwave-dashboard.json` into Grafana and select the Prometheus datasource that scrapes Pilotwave `/metrics`.

For Helm installs, the chart can render this dashboard as a Grafana sidecar ConfigMap:

```yaml
grafanaDashboard:
  enabled: true
  namespace: monitoring
```

The dashboard expects the metrics emitted by Pilotwave itself, including:

- `pilotwave_http_requests_total`
- `pilotwave_http_request_duration_seconds`
- `pilotwave_kubernetes_write_operations_total`
- `pilotwave_kubernetes_write_conflicts_total`
- `pilotwave_istio_resource_info`
- `pilotwave_istio_namespace_injection_state`
- `pilotwave_istio_tls_certificate_days_until_expiry`
- `pilotwave_istio_tls_certificate_expired`
- `pilotwave_istio_gateway_tls_secret_missing`
- `pilotwave_istio_gateway_tls_secret_invalid`

The Helm chart enables Prometheus scrape annotations by default. If the cluster uses `ServiceMonitor` or `PodMonitor`, keep the scrape target pointed at `/metrics` on the Pilotwave service port.
