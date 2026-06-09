package api

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server/middlewares"
	"git.brobridge.com/pilotwave/pilotwave/pkg/pagination"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	// "github.com/go-ozzo/ozzo-validation/is"
)

type Router struct {
	app    app.App
	server http_server.Server
	router *gin.RouterGroup
}

func NewRouter(a app.App, s http_server.Server) *Router {
	return &Router{
		app:    a,
		server: s,
	}
}

func (api *Router) Register() {

	api.router = api.server.GetEngine().Group("/api/v1")
	//	g.Router.Use(middlewares.RequiredRouter())
	// middlewares.RequiredAuth()

	// main
	api.router.GET("/routers", middlewares.RequiredAuth(), api.GetRouters)
	api.router.POST("/routers", middlewares.RequiredAuth(), api.CreateRouter)
	api.router.GET("/router/:namespace/:name", middlewares.RequiredAuth(), api.GetRouter)
	api.router.DELETE("/router/:namespace/:name", middlewares.RequiredAuth(), api.DeleteRouter)
	api.router.PUT("/router/:namespace/:name", middlewares.RequiredAuth(), api.UpdateRouter)

	// gateway mapping for router
	api.router.GET("/router/:namespace/:name/gateways", middlewares.RequiredAuth(), api.GetRouterGatewayMapping)
	api.router.POST("/router/:namespace/:name/gateways", middlewares.RequiredAuth(), api.CreateRouterGatewayMapping)
	api.router.PUT("/router/:namespace/:name/gateways", middlewares.RequiredAuth(), api.UpdateRouterGatewayMapping)
	api.router.DELETE("/router/:namespace/:name/gateways", middlewares.RequiredAuth(), api.DeleteRouterGatewayMapping)

	// router rules
	api.router.GET("/router/:namespace/:name/rules", middlewares.RequiredAuth(), api.GetRouterRule)
	api.router.PUT("/router/:namespace/:name/rules", middlewares.RequiredAuth(), api.UpdateRouterRule)

	// router collect information
	api.router.GET("/router/:namespace/:name/successrate", middlewares.RequiredAuth(), api.GetRouterSuccessRate)
	api.router.GET("/router/:namespace/:name/latency", middlewares.RequiredAuth(), api.GetRouterLatency)
	api.router.GET("/router/:namespace/:name/ops", middlewares.RequiredAuth(), api.GetRouterOPS)

	// grafana
	api.router.POST("/grafanas", middlewares.RequiredAuth(), api.UpdateGrafana)
	api.router.GET("/grafana", middlewares.RequiredAuth(), api.GetGrafana)
	api.router.POST("/monitoring/test", middlewares.RequiredAuth(), api.TestGrafana)
	// api.router.DELETE("/grafana/:grafanaId", middlewares.RequiredAuth(), api.DeleteGrafana)

	//	r := server.GetEngine()
}

func (api *Router) GetRouters(c *gin.Context) {

	// Preparing pagination conditions from querystring
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	search := c.DefaultQuery("search", "")
	namespace := c.DefaultQuery("namespace", "")
	// isDisabled := c.DefaultQuery("isDisabled", "")

	var resp []router_manager.RouterResponse
	total := 0

	// istio
	clusterMgr := api.app.GetClusterBridge()

	resp, total, err := clusterMgr.GetRouters(page, perPage, search, namespace)
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
		"routers": resp,
	})
}

type CreateResponse struct {
	Status string `json:"status"`
}

// @Summary Create Router
// @tags router
// @version v1
// @accept application/json
// @produce application/json
// @security Authentication
// @Success 200 {object} CreateResponse "http return code"
// @param body body router_manager.RouterRequest true "body"
// @Router /api/v1/routers [post]
func (api *Router) CreateRouter(c *gin.Context) {

	// Parsing body
	var body router_manager.RouterRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Name, validation.Required, validation.Match(regexp.MustCompile(`[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`))),
		validation.Field(&body.Namespace, validation.Required),
		validation.Field(&body.Protocol, validation.Required, validation.In("http", "http2", "https", "grpc", "socket", "tcp", "udp", "tls")),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err = clusterMgr.CreateRouter(body.Name, body.Namespace, body.Protocol, body.Hosts)
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
		"status": "create_success",
	})
}

func (api *Router) GetRouter(c *gin.Context) {

	// Check querys
	name := c.Param("name")
	namespace := c.Param("namespace")
	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	var resp *router_manager.RouterResponse
	resp, err := clusterMgr.GetRouter(name, namespace)
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
		"name": resp.Name,
		// "description": resp.Description,
		"protocol":        resp.Protocol,
		"hosts":           resp.Hosts,
		"namespace":       resp.Namespace,
		"createdAt":       resp.CreatedAt,
		"resourceversion": resp.ResourceVersion,
	})
}

func (api *Router) DeleteRouter(c *gin.Context) {

	// Check params
	name := c.Param("name")
	namespace := c.Param("namespace")
	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err := clusterMgr.DeleteRouter(name, namespace)
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
		"status": "delete_success",
	})
}

func (api *Router) UpdateRouter(c *gin.Context) {
	// Check params
	name := c.Param("name")
	namespace := c.Param("namespace")
	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// Parsing body
	var body router_manager.RouterUpdateRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Protocol, validation.Required, validation.In("http", "http2", "https", "grpc", "socket", "tcp", "udp", "tls")),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err = clusterMgr.UpdateRouter(name, namespace, body.Protocol, body.Hosts, body.ResourceVersion)
	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Router was modified in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "update_success",
	})
}

func (api *Router) GetRouterRule(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Param("namespace")
	err := validation.Errors{
		"name":      validation.Validate(name, validation.Required),
		"namespace": validation.Validate(namespace, validation.Required),
	}.Filter()
	if err != nil {
		c.JSON(http.StatusConflict, err)
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	res, err := clusterMgr.GetRouterRule(name, namespace)

	if err != nil {
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, res)
}

func (api *Router) UpdateRouterRule(c *gin.Context) {
	// Check params
	name := c.Param("name")
	namespace := c.Param("namespace")
	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// Parsing body
	var body router_manager.RouterRuleRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Https, validation.Required),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err = clusterMgr.UpdateRouterRule(name, namespace, body)
	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Router rules were modified in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "update_success",
	})

}

// @Summary Get Router's Request Success Rate metrics
// @tags router
// @version v1
// @accept application/json
// @produce application/json
// @security Authentication
// @Success 200 {object} router_manager.RouterSuccessRateResponse "http return code"
// @param name query string true "router name"
// @param namespace query string true "namespace"
// @param service query []string true "service"
// @param startTime query int false "startTime"
// @param endTime query int false "endTime"
// @param interval query string false "interval" default(30s)
// @Router /api/v1/router/{routerId}/successRate [get]
func (api *Router) GetRouterSuccessRate(c *gin.Context) {
	// Params for istio
	name := c.Param("name")
	namespace := c.Param("namespace")
	// services := c.QueryArray("service")

	err := validation.Errors{
		"name":      validation.Validate(name, validation.Required),
		"namespace": validation.Validate(namespace, validation.Required),
		// "service":   validation.Validate(services, validation.Required, validation.Length(1, 0), validation.Each(is.Domain, validation.Required)),
	}.Filter()
	if err != nil {
		c.JSON(http.StatusConflict, err)
		c.Abort()
		return
	}

	// Get StartTime and EndTime
	startTime, _ := strconv.Atoi(c.DefaultQuery("startTime", fmt.Sprint(time.Now().Unix()-(60*60))))
	endTime, _ := strconv.Atoi(c.DefaultQuery("endTime", fmt.Sprint(time.Now().Unix())))
	interval := strings.Replace(c.DefaultQuery("interval", "30"), ")", "", -1)

	// startTime should be smaller than endTime
	if startTime > endTime {
		c.JSON(http.StatusConflict, gin.H{"error": "startTime should be smaller than endTime"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	resp := make([]string, 0)
	resp, err = clusterMgr.GetRouterServices(name, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// use router manager  get SuccessRate
	routerMgr := api.app.GetRouter()

	services := make([]string, 0)
	for _, r := range resp {
		services = append(services, r+"."+namespace+".svc.cluster.local")
	}

	requestData := router_manager.RouterSuccessRateRequest{
		Name:      name,
		Namespace: namespace,
		Services:  services,
		StartTime: startTime,
		EndTime:   endTime,
		Interval:  interval,
	}

	res, err := routerMgr.GetRouterSuccessRate(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, res)
}

// @Summary Get Router's latency metrics
// @version v1
// @tags router
// @produce application/json
// @accept application/json
// @security Authentication
// @Success 200 {object} router_manager.RouterLatencyResponse "http return code"
// @param name query string true "router name"
// @param namespace query string true "namespace"
// @param service query []string true "service"
// @param startTime query int false "startTime"
// @param endTime query int false "endTime"
// @param interval query string false "interval" default(30s)
// @param percentage query float32 false "percentage" default(0.99)
// @Router /api/v1/router/{routerId}/latency [get]
func (api *Router) GetRouterLatency(c *gin.Context) {

	// Params for istio
	name := c.Param("name")
	namespace := c.Param("namespace")
	// services := c.QueryArray("service")

	percentage, err := strconv.ParseFloat(c.DefaultQuery("percentage", "0.99"), 64)
	if err != nil {
		c.JSON(http.StatusConflict, err)
		c.Abort()
		return
	}

	err = validation.Errors{
		"name":      validation.Validate(name, validation.Required),
		"namespace": validation.Validate(namespace, validation.Required),
		// "service":    validation.Validate(services, validation.Required, validation.Length(1, 0), validation.Each(is.Domain, validation.Required)),
		"percentage": validation.Validate(percentage, validation.Required, validation.Min(0.10), validation.Max(0.99)),
	}.Filter()
	if err != nil {
		c.JSON(http.StatusConflict, err)
		c.Abort()
		return
	}

	// Get StartTime and EndTime
	startTime, _ := strconv.Atoi(c.DefaultQuery("startTime", fmt.Sprint(time.Now().Unix()-(60*60))))
	endTime, _ := strconv.Atoi(c.DefaultQuery("endTime", fmt.Sprint(time.Now().Unix())))
	interval := strings.Replace(c.DefaultQuery("interval", "30"), ")", "", -1)

	// startTime should be smaller than endTime
	if startTime > endTime {
		c.JSON(http.StatusConflict, gin.H{"error": "startTime should be smaller than endTime"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	resp := make([]string, 0)
	resp, err = clusterMgr.GetRouterServices(name, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	//  use router manager  get SuccessRate
	routerMgr := api.app.GetRouter()

	services := make([]string, 0)
	for _, r := range resp {
		services = append(services, r+"."+namespace+".svc.cluster.local")
	}

	requestData := router_manager.RouterLatencyRequest{
		Name:       name,
		Namespace:  namespace,
		Percentage: percentage,
		Services:   services,
		StartTime:  startTime,
		EndTime:    endTime,
		Interval:   interval,
	}

	res, err := routerMgr.GetRouterLatency(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, res)
}

// @Summary Get Router's ops metrics
// @tags router
// @version v1
// @accept application/json
// @produce application/json
// @security Authentication
// @Success 200 {object} router_manager.RouterOPSResponse "http return code"
// @param name query string true "router name"
// @param namespace query string true "namespace"
// @param service query []string true "service"
// @param startTime query int false "startTime"
// @param endTime query int false "endTime"
// @param interval query string false "interval" default(30s)
// @Router /api/v1/router/{routerId}/ops [get]
func (api *Router) GetRouterOPS(c *gin.Context) {

	// Params for istio
	name := c.Param("name")
	namespace := c.Param("namespace")
	// services := c.QueryArray("service")
	err := validation.Errors{
		"name":      validation.Validate(name, validation.Required),
		"namespace": validation.Validate(namespace, validation.Required),
		// "service":   validation.Validate(services, validation.Required, validation.Length(1, 0), validation.Each(is.Domain, validation.Required)),
	}.Filter()
	if err != nil {
		c.JSON(http.StatusConflict, err)
		c.Abort()
		return
	}

	// Get StartTime and EndTime
	startTime, _ := strconv.Atoi(c.DefaultQuery("startTime", fmt.Sprint(time.Now().Unix()-(60*60))))
	endTime, _ := strconv.Atoi(c.DefaultQuery("endTime", fmt.Sprint(time.Now().Unix())))
	interval := strings.Replace(c.DefaultQuery("interval", "30"), ")", "", -1)

	// startTime should be smaller than endTime
	if startTime > endTime {
		c.JSON(http.StatusConflict, gin.H{"error": "startTime should be smaller than endTime"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	resp := make([]string, 0)
	resp, err = clusterMgr.GetRouterServices(name, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	//  use router manager  get SuccessRate
	routerMgr := api.app.GetRouter()

	services := make([]string, 0)
	for _, r := range resp {
		services = append(services, r+"."+namespace+".svc.cluster.local")
	}

	requestData := router_manager.RouterOPSRequest{
		Name:      name,
		Namespace: namespace,
		Services:  services,
		StartTime: startTime,
		EndTime:   endTime,
		Interval:  interval,
	}

	res, err := routerMgr.GetRouterOPS(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, res)
}

func (api *Router) GetRouterGatewayMapping(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	clusterMgr := api.app.GetClusterBridge()

	resp, err := clusterMgr.GetRouterGatewayMapping(name, namespace)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (api *Router) CreateRouterGatewayMapping(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	var body router_manager.RouterMappingRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Gateways, validation.Required),
	)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// validate
	if body.Gateways == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gateways items is required"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err = clusterMgr.CreateRouterGatewayMapping(name, namespace, body.Gateways, body.ResourceVersion)
	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Router gateway mapping changed in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "create_success",
	})

}

func (api *Router) UpdateRouterGatewayMapping(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	var body router_manager.RouterMappingRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// validate
	if body.Gateways == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gateways items is required"})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.UpdateRouterGatewayMapping(name, namespace, body.Gateways, body.ResourceVersion)
	if err != nil {
		log.Println(err)
		if isUpdateConflict(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Router gateway mapping changed in Kubernetes. Reload before applying changes."})
			c.Abort()
			return
		}
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "update_success",
	})

}

func (api *Router) DeleteRouterGatewayMapping(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err := clusterMgr.DeleteRouterGatewayMapping(name, namespace)
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
		"status": "delete_success",
	})

}

func (api *Router) UpdateGrafana(c *gin.Context) {

	// Parsing body
	var body router_manager.GrafanaRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Provider, validation.In("", "grafana", "prometheus")),
		validation.Field(&body.Host, validation.Required),
		validation.Field(&body.Port, validation.Required),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	routerMgr := api.app.GetRouter()

	resp := ""
	resp, err = routerMgr.UpdateGrafana(body)

	if err != nil || resp == "" {
		log.Println(err)
		errMsg := "unable to update monitoring source"
		if err != nil {
			errMsg = err.Error()
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": resp,
	})
}

func (api *Router) GetGrafana(c *gin.Context) {

	routerMgr := api.app.GetRouter()

	var resp *router_manager.GrafanaConfig
	resp, err := routerMgr.GetGrafana()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"id":            resp.ID,
		"configured":    strings.TrimSpace(resp.ID) != "",
		"provider":      resp.Provider,
		"host":          resp.Host,
		"port":          resp.Port,
		"token":         resp.Token,
		"datasourceId":  resp.DatasourceID,
		"isTls":         resp.Tls,
		"skipTlsVerify": resp.SkipTLSVerify,
		"createdAt":     resp.CreatedAt,
		"updatedAt":     resp.UpdatedAt,
	})

}

func (api *Router) TestGrafana(c *gin.Context) {
	var body router_manager.GrafanaRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	err = validation.ValidateStruct(&body,
		validation.Field(&body.Provider, validation.In("", "grafana", "prometheus")),
		validation.Field(&body.Host, validation.Required),
		validation.Field(&body.Port, validation.Required),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	routerMgr := api.app.GetRouter()
	resp, err := routerMgr.TestGrafana(body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, resp)
}
