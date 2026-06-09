package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server/middlewares"

	"github.com/gin-gonic/gin"
)

func TestMiddlewareUsesBoundedRouteLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Middleware())
	router.GET("/api/v1/router/:namespace/:name", func(c *gin.Context) {
		c.Status(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/router/default/router-a?token=secret", nil)
	router.ServeHTTP(httptest.NewRecorder(), req)

	metric := counterValue(t, RequestsTotal(), map[string]string{
		"method": http.MethodGet,
		"route":  "/api/v1/router/:namespace/:name",
		"status": "202",
	})
	if metric != 1 {
		t.Fatalf("expected route-pattern request counter to be 1, got %v", metric)
	}

	durationCount := histogramCount(t, RequestDurationSeconds(), map[string]string{
		"method": http.MethodGet,
		"route":  "/api/v1/router/:namespace/:name",
		"status": "202",
	})
	if durationCount != 1 {
		t.Fatalf("expected route-pattern duration histogram count to be 1, got %v", durationCount)
	}
}

func TestBuildInfoGaugeUsesExplicitSafeLabels(t *testing.T) {
	SetBuildInfo("test-version", "test-commit", "test-time")

	value := gaugeValue(t, BuildInfo(), map[string]string{
		"version":    "test-version",
		"commit":     "test-commit",
		"build_time": "test-time",
	})
	if value != 1 {
		t.Fatalf("expected build info gauge to be 1, got %v", value)
	}
}

func TestSetIstioClusterSnapshotPublishesBoundedGaugeLabels(t *testing.T) {
	SetIstioClusterSnapshot(IstioClusterSnapshot{
		Resources: []IstioResourceMetric{
			{Resource: "Gateway", Namespace: "edge", Name: "edge-gateway", Generation: 7},
		},
		NamespaceInjections: []IstioNamespaceInjectionMetric{
			{Namespace: "app", Mode: "revision", Revision: "canary"},
		},
		TLSCertificates: []IstioTLSCertificateMetric{
			{Namespace: "edge", Gateway: "edge-gateway", Secret: "wildcard-cert", NotAfterUnix: 1780000000, DaysUntilExpiry: 21, Expired: false},
		},
		GatewaySecretMissing: []IstioGatewayTLSSecretIssueMetric{
			{Namespace: "edge", Gateway: "edge-gateway", Secret: "missing-cert", Reason: "not_found"},
		},
		GatewaySecretInvalid: []IstioGatewayTLSSecretIssueMetric{
			{Namespace: "edge", Gateway: "edge-gateway", Secret: "invalid-cert", Reason: "parse_error"},
		},
	})

	if value := gaugeValue(t, IstioResourceInfo(), map[string]string{"resource": "Gateway", "namespace": "edge", "name": "edge-gateway"}); value != 1 {
		t.Fatalf("expected resource info gauge to be 1, got %v", value)
	}
	if value := gaugeValue(t, IstioResourceGeneration(), map[string]string{"resource": "Gateway", "namespace": "edge", "name": "edge-gateway"}); value != 7 {
		t.Fatalf("expected resource generation gauge to be 7, got %v", value)
	}
	if value := gaugeValue(t, IstioNamespaceInjectionState(), map[string]string{"namespace": "app", "mode": "revision", "revision": "canary"}); value != 1 {
		t.Fatalf("expected namespace injection gauge to be 1, got %v", value)
	}
	certLabels := map[string]string{"namespace": "edge", "gateway": "edge-gateway", "secret": "wildcard-cert"}
	if value := gaugeValue(t, IstioTLSCertificateNotAfterTimestamp(), certLabels); value != 1780000000 {
		t.Fatalf("expected certificate not_after gauge to be 1780000000, got %v", value)
	}
	if value := gaugeValue(t, IstioTLSCertificateDaysUntilExpiry(), certLabels); value != 21 {
		t.Fatalf("expected certificate days_until_expiry gauge to be 21, got %v", value)
	}
	if value := gaugeValue(t, IstioTLSCertificateExpired(), certLabels); value != 0 {
		t.Fatalf("expected certificate expired gauge to be 0, got %v", value)
	}
	if value := gaugeValue(t, IstioGatewayTLSSecretMissing(), map[string]string{"namespace": "edge", "gateway": "edge-gateway", "secret": "missing-cert", "reason": "not_found"}); value != 1 {
		t.Fatalf("expected missing secret gauge to be 1, got %v", value)
	}
	if value := gaugeValue(t, IstioGatewayTLSSecretInvalid(), map[string]string{"namespace": "edge", "gateway": "edge-gateway", "secret": "invalid-cert", "reason": "parse_error"}); value != 1 {
		t.Fatalf("expected invalid secret gauge to be 1, got %v", value)
	}
}

func TestMetricsEndpointAllowsUnauthenticatedScrapesAndIncludesHTTPMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetBuildInfo("handler-version", "handler-commit", "handler-time")

	router := gin.New()
	router.Use(Middleware())
	router.GET("/api/v1/users", middlewares.RequiredAuth(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/metrics", Handler())

	protectedRecorder := httptest.NewRecorder()
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	router.ServeHTTP(protectedRecorder, protectedReq)

	if protectedRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected protected route without auth to return 401, got %d", protectedRecorder.Code)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected /metrics status 200, got %d", recorder.Code)
	}

	body, err := io.ReadAll(recorder.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(body), `pilotwave_build_info{build_time="handler-time",commit="handler-commit",version="handler-version"} 1`) {
		t.Fatalf("expected build info metric in /metrics body, got:\n%s", string(body))
	}

	if !strings.Contains(string(body), `pilotwave_http_requests_total{method="GET",route="/api/v1/users",status="401"} 1`) {
		t.Fatalf("expected HTTP request counter for unauthenticated protected route in /metrics body, got:\n%s", string(body))
	}

	if !strings.Contains(string(body), `pilotwave_http_request_duration_seconds_bucket{method="GET",route="/api/v1/users",status="401",le=`) {
		t.Fatalf("expected HTTP request duration histogram for unauthenticated protected route in /metrics body, got:\n%s", string(body))
	}
}
