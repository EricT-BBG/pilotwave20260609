package api

import (
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server/middlewares"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"git.brobridge.com/pilotwave/pilotwave/pkg/pagination"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	// "github.com/spf13/viper"
)

type Gateway struct {
	app    app.App
	server http_server.Server
	router *gin.RouterGroup
}

func NewGateway(a app.App, s http_server.Server) *Gateway {
	return &Gateway{
		app:    a,
		server: s,
	}
}

func (api *Gateway) Register() {

	api.router = api.server.GetEngine().Group("/api/v1")
	//	g.Router.Use(middlewares.RequiredGateway())

	api.router.GET("/gateways", middlewares.RequiredAuth(), api.GetGateways)
	api.router.POST("/gateways", middlewares.RequiredAuth(), api.CreateGateway)
	api.router.GET("/gateway/tls-secret/exists", middlewares.RequiredAuth(), api.GetGatewayTLSSecretExists)
	api.router.GET("/gateway/:namespace/:name", middlewares.RequiredAuth(), api.GetGateway)
	api.router.GET("/gateway/:namespace/:name/tls-certificates", middlewares.RequiredAuth(), api.GetGatewayTLSCertificates)
	api.router.PUT("/gateway/:namespace/:name", middlewares.RequiredAuth(), api.UpdateGateway)
	api.router.DELETE("/gateway/:namespace/:name", middlewares.RequiredAuth(), api.DeleteGateway)

	api.router.GET("/gateway/:namespace/:name/routers", middlewares.RequiredAuth(), api.GetGatewayRouterMapping)
	api.router.POST("/gateways/:namespace/:name/routers", middlewares.RequiredAuth(), api.CreateGatewayRouterMapping)
	api.router.PUT("/gateway/:namespace/:name/routers", middlewares.RequiredAuth(), api.UpdateGatewayRouterMapping)
	api.router.DELETE("/gateway/:namespace/:name/routers", middlewares.RequiredAuth(), api.DeleteGatewayRouterMapping)

	//	r := server.GetEngine()
}

func gatewayTLSSecretNamespaceFromConfig() string {
	namespace := strings.TrimSpace(viper.GetString("gateway.tls_secret_namespace"))
	if namespace == "" {
		return "istio-system"
	}
	return namespace
}

func resolveGatewayTLSCredentialSecret(credentialName string) (string, string) {
	credentialName = strings.TrimSpace(credentialName)
	parts := strings.SplitN(credentialName, "/", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return gatewayTLSSecretNamespaceFromConfig(), credentialName
}

func (api *Gateway) GetGateways(c *gin.Context) {

	// Preparing pagination conditions from querystring
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	search := c.DefaultQuery("search", "")
	namespace := c.DefaultQuery("namespace", "")

	var resp []gateway_manager.GatewayResponse
	total := 0

	// istio
	clusterMgr := api.app.GetClusterBridge()

	resp, total, err := clusterMgr.GetGateways(page, perPage, search, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"meta": pagination.PaginationMeta{
			Page:    page,
			PerPage: perPage,
			Total:   total,
		},
		"gateways": resp,
	})

}

func (api *Gateway) GetGatewayTLSSecretExists(c *gin.Context) {
	credentialName := strings.TrimSpace(c.Query("credentialname"))
	if credentialName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "credentialname is required"})
		c.Abort()
		return
	}

	secretNamespace, secretName := resolveGatewayTLSCredentialSecret(credentialName)
	clusterMgr := api.app.GetClusterBridge()
	exists, err := clusterMgr.SecretsExist(secretName, secretNamespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exists":          exists,
		"secretNamespace": secretNamespace,
		"secretName":      secretName,
	})
}

func (api *Gateway) GetGatewayTLSCertificates(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	clusterMgr := api.app.GetClusterBridge()
	certificates, err := clusterMgr.GetGatewayTLSCertificates(name, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": certificates,
	})
}

func (api *Gateway) CreateGateway(c *gin.Context) {

	// Parsing body
	var body gateway_manager.GatewayRequest
	err := c.BindJSON(&body)

	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Gateway was modified in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Name, validation.Required, validation.Match(regexp.MustCompile(`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`))),
		validation.Field(&body.Namespace, validation.Required),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.CreateGateway(body.Name, body.Namespace, &body)

	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Gateway was modified in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "create_success",
	})
}

func (api *Gateway) UpdateGateway(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// Parsing body
	var body gateway_manager.GatewayRequest

	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Gateway was modified in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.UpdateGateway(name, namespace, &body)

	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Gateway was modified in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "update_success",
	})
}

func (api *Gateway) GetGateway(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	var resp *gateway_manager.GatewayResponse

	resp, err := clusterMgr.GetGateway(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"name":                resp.Name,
		"servers":             resp.Servers,
		"namespace":           resp.Namespace,
		"createdAt":           resp.CreatedAt,
		"selectormatchlabels": resp.SelectorMatchLabels,
		"resourceversion":     resp.ResourceVersion,
	})
}

func (api *Gateway) DeleteGateway(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	err := clusterMgr.DeleteGateway(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"status": "delete_success",
	})
}

func (api *Gateway) GetGatewayRouterMapping(c *gin.Context) {
	var resp gateway_manager.RouterMappingResponse

	name := c.Param("name")
	namespace := c.Param("namespace")

	// kubernetes
	clusterMgr := api.app.GetClusterBridge()

	resp, err := clusterMgr.GetGatewayRouterMapping(name, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"name":             resp.Name,
		"namespace":        resp.Namespace,
		"routers":          resp.Routers,
		"resourceversions": resp.ResourceVersions,
	})
}

func (api *Gateway) CreateGatewayRouterMapping(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	// Parsing body
	var body gateway_manager.GayewayMappingRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate
	if body.Routers == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Required routers not found",
		})
		c.Abort()
		return
	}

	// Create
	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.CreateGatewayRouterMapping(name, namespace, body.Routers, body.ResourceVersions)

	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Gateway router mapping changed in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "create_success",
	})
}
func (api *Gateway) UpdateGatewayRouterMapping(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Param("namespace")

	// Parsing body
	var body gateway_manager.GayewayMappingRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.UpdateGatewayRouterMapping(name, namespace, body.Routers, body.ResourceVersions)

	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Gateway router mapping changed in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "update_success",
	})
}

func (api *Gateway) DeleteGatewayRouterMapping(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	clusterMgr := api.app.GetClusterBridge()
	err := clusterMgr.DeleteGatewayRouterMapping(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"status": "delete_success",
	})
}
