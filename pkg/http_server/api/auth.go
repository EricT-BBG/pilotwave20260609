package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/auth"
	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server/middlewares"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
)

type Auth struct {
	app    app.App
	server http_server.Server
	router *gin.RouterGroup
}

type SignInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PatchNamespaceIstioInjectionRequest struct {
	Mode                 string `json:"mode"`
	Revision             string `json:"revision"`
	AllowSystemNamespace bool   `json:"allowSystemNamespace"`
}

func NewAuth(a app.App, s http_server.Server) *Auth {
	return &Auth{
		app:    a,
		server: s,
	}
}

func (api *Auth) Register() {

	api.router = api.server.GetEngine().Group("/api/v1")
	//	g.Router.Use(middlewares.RequiredAuth())

	api.router.POST("/auth/signin", api.SignIn)
	api.router.GET("/cluster/capabilities", middlewares.RequiredAuth(), api.GetClusterCapabilities)
	api.router.GET("/namespaces", middlewares.RequiredAuth(), api.GetNamespaces)
	api.router.PATCH("/namespace/:name/istio-injection", middlewares.RequiredAuth(), api.PatchNamespaceIstioInjection)

	//	r := server.GetEngine()
}

func (api *Auth) SignIn(c *gin.Context) {

	// Parsing body
	var body SignInRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		// c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Username, validation.Required),
		validation.Field(&body.Password, validation.Required),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	authenticator := api.app.GetAuthenticator()

	// Authenticate with username and password
	var resp *auth.AuthenticateResponse
	if viper.GetString("auth.method") != "ad" {

		// Built-in
		resp, err = authenticator.Authenticate(body.Username, body.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
	} else {

		// Active directory or LDAP
		resp, err = authenticator.AuthenticateWithAD(body.Username, body.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
	}

	if resp == nil {
		c.Status(http.StatusUnauthorized)
		c.Abort()
		return
	}

	// TODO: check permission for logging

	// Build a token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":      resp.ID,
		"username": resp.Username,
	})

	secret := viper.GetString("auth.secret")

	// Sign and get the complete encoded token as a string using the secret
	tokenStr, _ := token.SignedString([]byte(secret))

	// Response
	c.JSON(http.StatusOK, gin.H{
		"uid":         resp.ID,
		"name":        resp.Name,
		"username":    resp.Username,
		"email":       resp.Email,
		"permissions": resp.Permissions,
		"token":       tokenStr,
	})
}

func (api *Auth) GetClusterCapabilities(c *gin.Context) {
	clusterMgr := api.app.GetClusterBridge()

	c.JSON(http.StatusOK, gin.H{
		"istio": clusterMgr.GetIstioCapabilities(),
	})
}

func (api *Auth) GetNamespaces(c *gin.Context) {

	// kubernetes
	clusterMgr := api.app.GetClusterBridge()

	namespaces, err := clusterMgr.GetNamespaces()
	if err != nil {
		log.Println(err)
		if writeClusterUnavailable(c, err) {
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	items, err := clusterMgr.GetNamespaceMetadata()
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
		"namespaces": namespaces,
		"items":      items,
	})
}

func (api *Auth) PatchNamespaceIstioInjection(c *gin.Context) {

	namespace := strings.TrimSpace(c.Param("name"))
	if err := validateNamespaceName(namespace); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	var body PatchNamespaceIstioInjectionRequest
	if err := c.BindJSON(&body); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	body.Mode = strings.TrimSpace(body.Mode)
	body.Revision = strings.TrimSpace(body.Revision)
	if err := validateNamespaceIstioInjectionPatch(namespace, body); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	clusterMgr := api.app.GetClusterBridge()
	resp, err := clusterMgr.PatchNamespaceIstioInjection(namespace, body.Mode, body.Revision)
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
		"namespace": resp,
	})
}

func validateNamespaceName(name string) error {
	if name == "" {
		return errors.New("namespace name is required")
	}

	if errs := k8svalidation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid namespace name: %s", strings.Join(errs, ", "))
	}

	return nil
}

func validateNamespaceIstioInjectionPatch(namespace string, body PatchNamespaceIstioInjectionRequest) error {
	if cluster_bridge.IsSystemNamespaceName(namespace) && !body.AllowSystemNamespace {
		return errors.New("system namespace injection changes require explicit confirmation")
	}

	switch body.Mode {
	case cluster_bridge.NamespaceIstioInjectionModeDisabled, cluster_bridge.NamespaceIstioInjectionModeEnabled:
		if body.Revision != "" {
			return errors.New("revision must be empty unless mode is revision")
		}
	case cluster_bridge.NamespaceIstioInjectionModeRevision:
		if body.Revision == "" {
			return errors.New("revision is required when mode is revision")
		}
		if errs := k8svalidation.IsValidLabelValue(body.Revision); len(errs) > 0 {
			return fmt.Errorf("invalid revision: %s", strings.Join(errs, ", "))
		}
	default:
		return errors.New("mode must be disabled, enabled, or revision")
	}

	return nil
}
