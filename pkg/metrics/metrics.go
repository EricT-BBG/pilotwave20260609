package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "pilotwave"

	unknownRoute = "unmatched"
)

var (
	registry = prometheus.NewRegistry()

	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total HTTP requests handled by the Gin server.",
		},
		[]string{"method", "route", "status"},
	)

	requestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds for the Gin server.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	buildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "build_info",
			Help:      "Build metadata for the running Pilotwave process.",
		},
		[]string{"version", "commit", "build_time"},
	)

	kubernetesWriteOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "kubernetes_write_operations_total",
			Help:      "Total Kubernetes and Istio write operations.",
		},
		[]string{"resource", "verb", "result"},
	)

	kubernetesWriteOperationDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "kubernetes_write_operation_duration_seconds",
			Help:      "Kubernetes and Istio write operation duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"resource", "verb", "result"},
	)

	kubernetesWriteConflictsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "kubernetes_write_conflicts_total",
			Help:      "Total Kubernetes and Istio write operations rejected because of conflict or stale resourceVersion.",
		},
		[]string{"resource", "verb"},
	)

	istioResourceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_resource_info",
			Help:      "Istio-managed resource presence by resource kind, namespace, and name.",
		},
		[]string{"resource", "namespace", "name"},
	)

	istioResourceGeneration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_resource_generation",
			Help:      "Kubernetes metadata generation for Istio-managed resources.",
		},
		[]string{"resource", "namespace", "name"},
	)

	istioNamespaceInjectionState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_namespace_injection_state",
			Help:      "Namespace Istio sidecar injection state.",
		},
		[]string{"namespace", "mode", "revision"},
	)

	istioTLSCertificateNotAfterTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_tls_certificate_not_after_timestamp",
			Help:      "Unix timestamp of the earliest not_after value for a resolved Istio TLS certificate secret.",
		},
		[]string{"namespace", "gateway", "secret"},
	)

	istioTLSCertificateDaysUntilExpiry = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_tls_certificate_days_until_expiry",
			Help:      "Days until expiry for a resolved Istio TLS certificate secret.",
		},
		[]string{"namespace", "gateway", "secret"},
	)

	istioTLSCertificateExpired = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_tls_certificate_expired",
			Help:      "Whether a resolved Istio TLS certificate secret is expired.",
		},
		[]string{"namespace", "gateway", "secret"},
	)

	istioGatewayTLSSecretMissing = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_gateway_tls_secret_missing",
			Help:      "Istio Gateway TLS credentialName references that could not be resolved to Kubernetes secrets.",
		},
		[]string{"namespace", "gateway", "secret", "reason"},
	)

	istioGatewayTLSSecretInvalid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "istio_gateway_tls_secret_invalid",
			Help:      "Istio Gateway TLS credentialName secrets that are present but invalid for certificate health export.",
		},
		[]string{"namespace", "gateway", "secret", "reason"},
	)
)

func init() {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		requestsTotal,
		requestDurationSeconds,
		buildInfo,
		kubernetesWriteOperationsTotal,
		kubernetesWriteOperationDurationSeconds,
		kubernetesWriteConflictsTotal,
		istioResourceInfo,
		istioResourceGeneration,
		istioNamespaceInjectionState,
		istioTLSCertificateNotAfterTimestamp,
		istioTLSCertificateDaysUntilExpiry,
		istioTLSCertificateExpired,
		istioGatewayTLSSecretMissing,
		istioGatewayTLSSecretInvalid,
	)
}

// SetBuildInfo publishes a single build-info sample with bounded label values.
func SetBuildInfo(version, commit, buildTime string) {
	buildInfo.Reset()
	buildInfo.WithLabelValues(safeBuildLabel(version), safeBuildLabel(commit), safeBuildLabel(buildTime)).Set(1)
}

// Handler returns an unauthenticated Prometheus scrape endpoint handler.
func Handler() gin.HandlerFunc {
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return gin.WrapH(handler)
}

// Middleware records bounded HTTP request metrics for Gin routes.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = unknownRoute
		}

		status := strconv.Itoa(c.Writer.Status())
		labels := []string{c.Request.Method, route, status}
		requestsTotal.WithLabelValues(labels...).Inc()
		requestDurationSeconds.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	}
}

func RequestsTotal() *prometheus.CounterVec {
	return requestsTotal
}

func RequestDurationSeconds() *prometheus.HistogramVec {
	return requestDurationSeconds
}

func BuildInfo() *prometheus.GaugeVec {
	return buildInfo
}

func RecordKubernetesWrite(resource, verb, result string, start time.Time) {
	labels := []string{
		safeMetricsLabel(resource),
		safeMetricsLabel(verb),
		safeMetricsLabel(result),
	}
	kubernetesWriteOperationsTotal.WithLabelValues(labels...).Inc()
	kubernetesWriteOperationDurationSeconds.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
}

func RecordKubernetesWriteConflict(resource, verb string) {
	kubernetesWriteConflictsTotal.WithLabelValues(safeMetricsLabel(resource), safeMetricsLabel(verb)).Inc()
}

func KubernetesWriteOperationsTotal() *prometheus.CounterVec {
	return kubernetesWriteOperationsTotal
}

func KubernetesWriteOperationDurationSeconds() *prometheus.HistogramVec {
	return kubernetesWriteOperationDurationSeconds
}

func KubernetesWriteConflictsTotal() *prometheus.CounterVec {
	return kubernetesWriteConflictsTotal
}

type IstioResourceMetric struct {
	Resource   string
	Namespace  string
	Name       string
	Generation int64
}

type IstioNamespaceInjectionMetric struct {
	Namespace string
	Mode      string
	Revision  string
}

type IstioTLSCertificateMetric struct {
	Namespace       string
	Gateway         string
	Secret          string
	NotAfterUnix    float64
	DaysUntilExpiry float64
	Expired         bool
}

type IstioGatewayTLSSecretIssueMetric struct {
	Namespace string
	Gateway   string
	Secret    string
	Reason    string
}

type IstioClusterSnapshot struct {
	Resources            []IstioResourceMetric
	NamespaceInjections  []IstioNamespaceInjectionMetric
	TLSCertificates      []IstioTLSCertificateMetric
	GatewaySecretMissing []IstioGatewayTLSSecretIssueMetric
	GatewaySecretInvalid []IstioGatewayTLSSecretIssueMetric
}

func SetIstioClusterSnapshot(snapshot IstioClusterSnapshot) {
	istioResourceInfo.Reset()
	istioResourceGeneration.Reset()
	istioNamespaceInjectionState.Reset()
	istioTLSCertificateNotAfterTimestamp.Reset()
	istioTLSCertificateDaysUntilExpiry.Reset()
	istioTLSCertificateExpired.Reset()
	istioGatewayTLSSecretMissing.Reset()
	istioGatewayTLSSecretInvalid.Reset()

	for _, resource := range snapshot.Resources {
		labels := []string{
			resource.Resource,
			resource.Namespace,
			resource.Name,
		}
		istioResourceInfo.WithLabelValues(labels...).Set(1)
		istioResourceGeneration.WithLabelValues(labels...).Set(float64(resource.Generation))
	}

	for _, namespace := range snapshot.NamespaceInjections {
		istioNamespaceInjectionState.WithLabelValues(
			namespace.Namespace,
			namespace.Mode,
			namespace.Revision,
		).Set(1)
	}

	for _, cert := range snapshot.TLSCertificates {
		labels := []string{cert.Namespace, cert.Gateway, cert.Secret}
		istioTLSCertificateNotAfterTimestamp.WithLabelValues(labels...).Set(cert.NotAfterUnix)
		istioTLSCertificateDaysUntilExpiry.WithLabelValues(labels...).Set(cert.DaysUntilExpiry)
		if cert.Expired {
			istioTLSCertificateExpired.WithLabelValues(labels...).Set(1)
		} else {
			istioTLSCertificateExpired.WithLabelValues(labels...).Set(0)
		}
	}

	for _, issue := range snapshot.GatewaySecretMissing {
		istioGatewayTLSSecretMissing.WithLabelValues(
			issue.Namespace,
			issue.Gateway,
			issue.Secret,
			issue.Reason,
		).Set(1)
	}

	for _, issue := range snapshot.GatewaySecretInvalid {
		istioGatewayTLSSecretInvalid.WithLabelValues(
			issue.Namespace,
			issue.Gateway,
			issue.Secret,
			issue.Reason,
		).Set(1)
	}
}

func IstioResourceInfo() *prometheus.GaugeVec {
	return istioResourceInfo
}

func IstioResourceGeneration() *prometheus.GaugeVec {
	return istioResourceGeneration
}

func IstioNamespaceInjectionState() *prometheus.GaugeVec {
	return istioNamespaceInjectionState
}

func IstioTLSCertificateNotAfterTimestamp() *prometheus.GaugeVec {
	return istioTLSCertificateNotAfterTimestamp
}

func IstioTLSCertificateDaysUntilExpiry() *prometheus.GaugeVec {
	return istioTLSCertificateDaysUntilExpiry
}

func IstioTLSCertificateExpired() *prometheus.GaugeVec {
	return istioTLSCertificateExpired
}

func IstioGatewayTLSSecretMissing() *prometheus.GaugeVec {
	return istioGatewayTLSSecretMissing
}

func IstioGatewayTLSSecretInvalid() *prometheus.GaugeVec {
	return istioGatewayTLSSecretInvalid
}

func safeBuildLabel(value string) string {
	return safeMetricsLabel(value)
}

func safeMetricsLabel(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}
