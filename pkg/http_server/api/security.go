package api

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"git.brobridge.com/pilotwave/pilotwave/pkg/pagination"
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
)

// Basic struct
type Security struct {
	app    app.App
	server http_server.Server
	router *gin.RouterGroup
}

func NewSecurity(a app.App, s http_server.Server) *Security {
	return &Security{
		app:    a,
		server: s,
	}
}

func (api *Security) Register() {

	api.router = api.server.GetEngine().Group("/api/v1")

	// authpolicy
	api.router.GET("/security/authpolicies", api.GetAuthorizationPolicies)
	api.router.POST("/security/authpolicies", api.CreateAuthorizationPolicy)

	api.router.GET("/security/authpolicy/:namespace/:name", api.GetAuthorizationPolicy)
	api.router.PUT("/security/authpolicy/:namespace/:name", api.UpdateAuthorizationPolicy)
	api.router.DELETE("/security/authpolicy/:namespace/:name", api.DeleteAuthorizationPolicy)

	// requrstauth
	api.router.GET("/security/requestauths", api.GetRequestAuthentications)

	api.router.POST("/security/requestauths", api.CreateRequestAuthentication)
	api.router.GET("/security/requestauth/:namespace/:name", api.GetRequestAuthentication)
	api.router.PUT("/security/requestauth/:namespace/:name", api.UpdateRequestAuthentication)
	api.router.DELETE("/security/requestauth/:namespace/:name", api.DeleteRequestAuthentication)

}

// AuthorizationPolicy

func (api *Security) GetAuthorizationPolicies(c *gin.Context) {

	name := c.DefaultQuery("name", "")
	namespace := c.DefaultQuery("namespace", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	clusterMgr := api.app.GetClusterBridge()
	resp, total, err := clusterMgr.GetAuthorizationPolicies(page, perPage, name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	} else if resp == nil {
		resp = make([]*security_manager.AuthorizationPolicyResponse, 0, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": pagination.PaginationMeta{
			Page:    page,
			PerPage: perPage,
			Total:   total,
		},
		"results": resp,
	})

}

func (api *Security) GetAuthorizationPolicy(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()

	var resp *security_manager.AuthorizationPolicyResponse

	resp, err := clusterMgr.GetAuthorizationPolicy(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return

	} else if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, resp)

}

func (api *Security) CreateAuthorizationPolicy(c *gin.Context) {

	// Parsing body
	var body security_manager.AuthorizationPolicyRequest
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
		validation.Field(&body.Action, validation.Required, validation.In("allow", "deny", "audit")),
	)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err = clusterMgr.CreateAuthorizationPolicy(body.Name, body.Namespace, &body)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "create_success"})

}

func (api *Security) UpdateAuthorizationPolicy(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// Parsing body
	var body security_manager.AuthorizationPolicyUpdateRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Action, validation.Required, validation.In("allow", "deny", "audit")),
	)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()

	err = clusterMgr.UpdateAuthorizationPolicy(name, namespace, &body)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "update_success"})

}

func (api *Security) DeleteAuthorizationPolicy(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()

	err := clusterMgr.DeleteAuthorizationPolicy(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "delete_success"})

}

// RequestAuthentication

func (api *Security) GetRequestAuthentications(c *gin.Context) {

	name := c.DefaultQuery("name", "")
	namespace := c.DefaultQuery("namespace", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	clusterMgr := api.app.GetClusterBridge()
	resp, total, err := clusterMgr.GetRequestAuthentications(page, perPage, name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	} else if resp == nil {
		resp = make([]*security_manager.RequestAuthenticationResponse, 0, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": pagination.PaginationMeta{
			Page:    page,
			PerPage: perPage,
			Total:   total,
		},
		"results": resp,
	})

}

func (api *Security) GetRequestAuthentication(c *gin.Context) {
	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusConflict, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()

	var resp *security_manager.RequestAuthenticationResponse
	resp, err := clusterMgr.GetRequestAuthentication(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	} else if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, resp)

}

func (api *Security) CreateRequestAuthentication(c *gin.Context) {

	// Parsing body
	var body security_manager.RequestAuthenticationRequest
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
	)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// check issuers
	if len(body.JWTRules) != 0 {
		for _, r := range body.JWTRules {
			if r.Issuer == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Issuer is empty."})
				c.Abort()
				return
			}
		}
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.CreateRequestAuthentication(body.Name, body.Namespace, &body)
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "create_success"})
}

func (api *Security) UpdateRequestAuthentication(c *gin.Context) {

	// Get URL parameter
	name := c.Param("name")
	namespace := c.Param("namespace")

	// check
	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	// Parsing body
	var body security_manager.RequestAuthenticationUpdateRequest

	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// check issuers
	if len(body.JWTRules) != 0 {
		for _, r := range body.JWTRules {
			if r.Issuer == "" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Issuer is empty."})
				c.Abort()
				return
			}
		}
	}

	// istio
	clusterMgr := api.app.GetClusterBridge()
	err = clusterMgr.UpdateRequestAuthentication(name, namespace, &body)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "update_success"})

}

func (api *Security) DeleteRequestAuthentication(c *gin.Context) {

	name := c.Param("name")
	namespace := c.Param("namespace")

	if name == "" || namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and namespace is required"})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()

	err := clusterMgr.DeleteRequestAuthentication(name, namespace)

	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "delete_success"})

}
