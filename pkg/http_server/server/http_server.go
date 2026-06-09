package server

import (
	"net"
	"net/http"

	app "git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo"
	api "git.brobridge.com/pilotwave/pilotwave/pkg/http_server/api"
	static "git.brobridge.com/pilotwave/pilotwave/pkg/http_server/static"
	"git.brobridge.com/pilotwave/pilotwave/pkg/metrics"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
)

type Server struct {
	app      app.App
	engine   *gin.Engine
	instance *http.Server
	listener net.Listener
	host     string
}

func NewServer(a app.App) *Server {
	return &Server{
		app:      a,
		instance: &http.Server{},
	}
}

func (server *Server) Init(host string) error {

	// Put it to mux
	mux, err := server.app.GetMuxManager().AssertMux("http", host)
	if err != nil {
		return err
	}

	// Preparing listener
	lis := mux.Match(cmux.HTTP1Fast())
	server.host = host
	server.listener = lis
	server.engine = gin.Default()

	// Setup Cross
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = append(corsConfig.AllowMethods, "DELETE")
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authentication")
	server.engine.Use(cors.New(corsConfig))
	server.engine.Use(metrics.Middleware())
	server.engine.GET("/metrics", metrics.Handler())
	metrics.SetBuildInfo(buildinfo.Values())

	// Static assets
	static.RegisterHandler(server)

	// APIs
	api.NewAuth(server.app, server).Register()
	api.NewUser(server.app, server).Register()
	api.NewGateway(server.app, server).Register()
	api.NewRouter(server.app, server).Register()
	api.NewSecurity(server.app, server).Register()

	//Swagger
	api.NewSwag(server.app, server).Register()

	server.instance.Handler = server.engine

	return nil
}

func (server *Server) Serve() error {

	log.WithFields(log.Fields{
		"host": server.host,
	}).Info("Starting HTTP server")

	// Starting server
	if err := server.instance.Serve(server.listener); err != cmux.ErrListenerClosed {
		log.Error(err)
		return err
	}

	return nil
}

func (server *Server) GetEngine() *gin.Engine {
	return server.engine
}

func (server *Server) GetApp() app.App {
	return server.app
}
