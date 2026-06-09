package static

import (
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"

	httpserver "git.brobridge.com/pilotwave/pilotwave/pkg/http_server"

	"github.com/gin-gonic/gin"
)

var regexpAsset = regexp.MustCompile(`src="([^"]+\.js)"`)

type testServer struct {
	engine *gin.Engine
}

func (server testServer) Init(string) error {
	return nil
}

func (server testServer) Serve() error {
	return nil
}

func (server testServer) GetEngine() *gin.Engine {
	return server.engine
}

var _ httpserver.Server = testServer{}

func TestRegisterHandlerServesEmbeddedDistAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	RegisterHandler(testServer{engine: engine})

	htmlRecorder := httptest.NewRecorder()
	engine.ServeHTTP(htmlRecorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if htmlRecorder.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", htmlRecorder.Code, http.StatusOK)
	}

	matches := regexpAsset.FindStringSubmatch(htmlRecorder.Body.String())
	if len(matches) != 2 {
		t.Fatalf("index HTML did not reference a built JS asset")
	}

	assetPath := path.Clean(matches[1])
	if !strings.HasPrefix(assetPath, "/dist/assets/") {
		t.Fatalf("asset path = %q, want /dist/assets/*", assetPath)
	}

	assetRecorder := httptest.NewRecorder()
	engine.ServeHTTP(assetRecorder, httptest.NewRequest(http.MethodGet, assetPath, nil))
	if assetRecorder.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", assetPath, assetRecorder.Code, http.StatusOK)
	}

	contentType := assetRecorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "javascript") {
		t.Fatalf("GET %s Content-Type = %q, want JavaScript", assetPath, contentType)
	}
	if strings.Contains(assetRecorder.Body.String(), "<div id=\"app\"></div>") {
		t.Fatalf("GET %s returned index HTML instead of the JS asset", assetPath)
	}
}

func TestRegisterHandlerFallsBackToIndexForDistHistoryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	RegisterHandler(testServer{engine: engine})

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/dist/dashboard", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /dist/dashboard status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "<div id=\"app\"></div>") {
		t.Fatalf("GET /dist/dashboard did not return index HTML")
	}
}
