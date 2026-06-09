package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type conflictAPIApp struct {
	app.App

	bridge cluster_bridge.Bridge
}

func (a conflictAPIApp) GetClusterBridge() cluster_bridge.Bridge {
	return a.bridge
}

type conflictAPIServer struct {
	http_server.Server

	engine *gin.Engine
}

func (s conflictAPIServer) GetEngine() *gin.Engine {
	return s.engine
}

type conflictBridge struct {
	cluster_bridge.Bridge
}

func (b conflictBridge) UpdateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error {
	return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "gateways"}, name, nil)
}

func (b conflictBridge) UpdateRouter(name string, namespace string, protocol string, hosts []string, resourceVersion string) error {
	return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "virtualservices"}, name, nil)
}

func TestUpdateGatewayMapsStaleResourceVersionConflictToHTTPConflict(t *testing.T) {
	response := runConflictAPIRequest(t, func(engine *gin.Engine, app app.App, server http_server.Server) {
		gatewayAPI := NewGateway(app, server)
		engine.PUT("/api/v1/gateway/:namespace/:name", gatewayAPI.UpdateGateway)
	}, http.MethodPut, "/api/v1/gateway/default/example", map[string]interface{}{
		"resourceversion": "stale",
		"servers":         []interface{}{},
	})

	assertConflictResponse(t, response, "Gateway was modified in Kubernetes. Reload before applying changes.")
}

func TestUpdateRouterMapsStaleResourceVersionConflictToHTTPConflict(t *testing.T) {
	response := runConflictAPIRequest(t, func(engine *gin.Engine, app app.App, server http_server.Server) {
		routerAPI := NewRouter(app, server)
		engine.PUT("/api/v1/router/:namespace/:name", routerAPI.UpdateRouter)
	}, http.MethodPut, "/api/v1/router/default/example", map[string]interface{}{
		"protocol":        "http",
		"hosts":           []string{"example.local"},
		"resourceversion": "stale",
	})

	assertConflictResponse(t, response, "Router was modified in Kubernetes. Reload before applying changes.")
}

func runConflictAPIRequest(t *testing.T, register func(*gin.Engine, app.App, http_server.Server), method string, target string, body map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	server := conflictAPIServer{engine: engine}
	application := conflictAPIApp{bridge: conflictBridge{}}
	register(engine, application, server)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	request := httptest.NewRequest(method, target, bytes.NewReader(data))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func assertConflictResponse(t *testing.T, response *httptest.ResponseRecorder, expectedError string) {
	t.Helper()

	if response.Code != http.StatusConflict {
		t.Fatalf("expected HTTP 409, got %d with body %s", response.Code, response.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response body: %v", err)
	}
	if body["error"] != expectedError {
		t.Fatalf("unexpected error body: got %q want %q", body["error"], expectedError)
	}
}
