package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildSuccessRateQueryUsesIstio17CompatiblePercent(t *testing.T) {
	query := buildSuccessRateQuery("reviews.default.svc.cluster.local")

	if !strings.Contains(query, `100 *`) {
		t.Fatalf("success rate query should return a percentage, got %s", query)
	}

	if strings.Contains(query, `destination_service=~`) {
		t.Fatalf("success rate query should use an exact service matcher, got %s", query)
	}

	if strings.Contains(query, `[1m]`) {
		t.Fatalf("success rate numerator and denominator should use the same stable window, got %s", query)
	}

	if !strings.Contains(query, `istio_requests_total{reporter="destination",destination_service="reviews.default.svc.cluster.local",response_code!~"5.*"}[5m]`) {
		t.Fatalf("success rate query should include successful Istio destination requests, got %s", query)
	}
}

func TestBuildLatencyQueryPrefersIstio17SecondsMetric(t *testing.T) {
	query := buildLatencyQuery("reviews.default.svc.cluster.local", 0.99)

	secondsIndex := strings.Index(query, `istio_request_duration_seconds_bucket`)
	millisecondsIndex := strings.Index(query, `istio_request_duration_milliseconds_bucket`)

	if secondsIndex < 0 {
		t.Fatalf("latency query should support Istio 1.7 seconds buckets, got %s", query)
	}

	if millisecondsIndex < 0 {
		t.Fatalf("latency query should retain milliseconds bucket fallback for newer Istio, got %s", query)
	}

	if secondsIndex > millisecondsIndex {
		t.Fatalf("latency query should prefer Istio 1.7 seconds buckets before newer milliseconds buckets, got %s", query)
	}
}

func TestGrafanaPrometheusQueryPathUsesConfiguredDatasourceID(t *testing.T) {
	path := grafanaPrometheusQueryRangePath("42")

	if path != "/api/datasources/proxy/42/api/v1/query_range" {
		t.Fatalf("grafana proxy path should use configured datasource id, got %s", path)
	}
}

func TestGrafanaPrometheusQueryPathDefaultsToDatasourceOne(t *testing.T) {
	path := grafanaPrometheusQueryRangePath("")

	if path != "/api/datasources/proxy/1/api/v1/query_range" {
		t.Fatalf("grafana proxy path should default to datasource id 1, got %s", path)
	}
}

func TestRequestWithTimeoutReturnsPromptly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	router := &Router{}
	startedAt := time.Now()
	res, err := router.requestWithTimeout("GET", server.URL, false, false, "", "", 25*time.Millisecond)
	if res != nil {
		res.Body.Close()
	}

	if err == nil {
		t.Fatal("request should time out")
	}
	if time.Since(startedAt) > 150*time.Millisecond {
		t.Fatalf("request timeout took too long: %s", time.Since(startedAt))
	}
}
