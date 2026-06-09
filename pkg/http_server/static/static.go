package static

import (
	http_server "git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"path"
	"strings"
)

//go:generate go run -tags=dev assets_generate.go
func RegisterHandler(server http_server.Server) {

	r := server.GetEngine()

	r.GET("/dist/*filepath", serveDistAsset)
	r.HEAD("/dist/*filepath", serveDistAsset)
	r.NoRoute(func(c *gin.Context) {
		serveIndex(c)
	})
}

func serveDistAsset(c *gin.Context) {
	name := strings.TrimPrefix(c.Param("filepath"), "/")
	if name == "" {
		serveIndex(c)
		return
	}

	if path.Ext(name) == "" {
		serveIndex(c)
		return
	}
	c.FileFromFS(name, Assets)
}

func serveIndex(c *gin.Context) {
	file, err := Assets.Open("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html not found")
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html could not be read")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", content)
}
