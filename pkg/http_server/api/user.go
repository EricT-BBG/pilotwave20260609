package api

import (
	"log"
	"net/http"
	"strconv"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server/middlewares"
	"git.brobridge.com/pilotwave/pilotwave/pkg/pagination"
	"git.brobridge.com/pilotwave/pilotwave/pkg/user_manager"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	is "github.com/go-ozzo/ozzo-validation/is"
	// "github.com/spf13/viper"
)

type User struct {
	app    app.App
	server http_server.Server
	router *gin.RouterGroup
}

func NewUser(a app.App, s http_server.Server) *User {
	return &User{
		app:    a,
		server: s,
	}
}

func (api *User) Register() {

	api.router = api.server.GetEngine().Group("/api/v1")

	api.router.GET("/users", middlewares.RequiredAuth(), api.GetUsers)
	api.router.POST("/users", middlewares.RequiredAuth(), api.CreateUser)
	api.router.GET("/user/:userId", middlewares.RequiredAuth(), api.GetUser)
	api.router.DELETE("/user/:userId", middlewares.RequiredAuth(), api.DeleteUser)
	api.router.PUT("/user/:userId/enabled", middlewares.RequiredAuth(), api.EnableUser)
	api.router.PUT("/user/:userId", middlewares.RequiredAuth(), api.UpdateUser)
	api.router.PUT("/user/:userId/resetpassword", middlewares.RequiredAuth(), api.UpdateUserPassword)

	// TODO: set permissions
}

func (api *User) GetUsers(c *gin.Context) {

	// Preparing pagination conditions from querystring
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.DefaultQuery("search", "")
	isDisabled := c.DefaultQuery("isDisabled", "")

	userMgr := api.app.GetUser()

	var resp []user_manager.UserResponse
	resp, total, err := userMgr.GetUsers(page, perPage, search, isDisabled)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
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
		"users": resp,
	})
}

func (api *User) CreateUser(c *gin.Context) {

	// Parsing body
	var body user_manager.UserRequest
	err := c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Name, validation.Required),
		validation.Field(&body.Username, validation.Required, validation.RuneLength(1, 255)),
		validation.Field(&body.Password, validation.Required),
		validation.Field(&body.Email, validation.Required, is.Email),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	userMgr := api.app.GetUser()

	resp := ""
	resp, err = userMgr.CreateUser(body.Name, body.Username, body.Password, body.Email, body.Permissions)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": resp,
	})
}

func (api *User) UpdateUser(c *gin.Context) {

	// Check parameter
	userId := c.Param("userId")
	err := validation.Validate(userId, validation.Required, is.UUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Parsing body
	var body user_manager.UserRequest

	err = c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Name, validation.Required),
		validation.Field(&body.Email, validation.Required, is.Email),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	userMgr := api.app.GetUser()

	resp := ""
	resp, err = userMgr.UpdateUser(userId, body.Name, body.Email, body.Permissions)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": resp,
	})
}

func (api *User) GetUser(c *gin.Context) {

	// Check parameter
	userId := c.Param("userId")
	err := validation.Validate(userId, validation.Required, is.UUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	userMgr := api.app.GetUser()

	var resp *user_manager.UserResponse
	resp, err = userMgr.GetUser(userId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"id":          resp.ID,
		"name":        resp.Name,
		"username":    resp.Username,
		"email":       resp.Email,
		"permissions": resp.Permissions,
		"isDisabled":  resp.IsDisabled,
		"createdAt":   resp.CreatedAt,
		"updatedAt":   resp.UpdatedAt,
	})
}

func (api *User) DeleteUser(c *gin.Context) {

	// Check parameter
	userId := c.Param("userId")
	err := validation.Validate(userId, validation.Required, is.UUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	userMgr := api.app.GetUser()

	resp := ""
	resp, err = userMgr.DeleteUser(userId)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Response
	c.JSON(http.StatusOK, gin.H{
		"id": resp,
	})
}

func (api *User) EnableUser(c *gin.Context) {

	// Check parameter
	userId := c.Param("userId")
	err := validation.Validate(userId, validation.Required, is.UUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Parsing body
	var body user_manager.UserEnableRequest

	log.Println(body)
	err = c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	userMgr := api.app.GetUser()

	resp := ""
	resp, err = userMgr.EnableUser(userId, body.Enabled)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": resp,
	})
}

func (api *User) UpdateUserPassword(c *gin.Context) {

	// Check parameter
	userId := c.Param("userId")
	err := validation.Validate(userId, validation.Required, is.UUID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Parsing body
	var body user_manager.UserPasswordRequest

	log.Println(body)
	err = c.BindJSON(&body)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	// Validate fields
	err = validation.ValidateStruct(&body,
		validation.Field(&body.Password, validation.Required),
	)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	userMgr := api.app.GetUser()

	resp := ""
	resp, err = userMgr.UpdateUserPassword(userId, body.Password)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": resp,
	})
}
