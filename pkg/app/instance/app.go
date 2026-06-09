package instance

import (
	authenticator "git.brobridge.com/pilotwave/pilotwave/pkg/auth/authenticator"
	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge/bridge"
	database "git.brobridge.com/pilotwave/pilotwave/pkg/database"
	gateway "git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager/gateway"
	http_server "git.brobridge.com/pilotwave/pilotwave/pkg/http_server/server"
	mux_manager "git.brobridge.com/pilotwave/pilotwave/pkg/mux_manager/manager"
	router "git.brobridge.com/pilotwave/pilotwave/pkg/router_manager/router"
	security "git.brobridge.com/pilotwave/pilotwave/pkg/security_manager/security"
	user "git.brobridge.com/pilotwave/pilotwave/pkg/user_manager/user"
	log "github.com/sirupsen/logrus"
)

type AppInstance struct {
	done          chan bool
	muxManager    *mux_manager.MuxManager
	httpServer    *http_server.Server
	database      *database.Database
	authenticator *authenticator.Authenticator
	user          *user.User
	gateway       *gateway.Gateway
	router        *router.Router
	clusterBridge *cluster_bridge.ClusterBridge
	security      *security.Security
}

func NewAppInstance() *AppInstance {

	a := &AppInstance{
		done: make(chan bool),
	}

	return a
}

func (a *AppInstance) Init() error {

	log.Info("Starting application")

	// Initializing modules
	a.database = database.NewDatabase()
	a.clusterBridge = cluster_bridge.NewClusterBridge(a)
	a.authenticator = authenticator.NewAuthenticator(a)
	a.user = user.NewUser(a)
	a.gateway = gateway.NewGateway(a)
	a.router = router.NewRouter(a)
	a.muxManager = mux_manager.NewMuxManager(a)
	a.security = security.NewSecurity(a)
	a.httpServer = http_server.NewServer(a)

	// Initializing database connector
	err := a.initDatabase()
	if err != nil {
		return err
	}

	a.initMuxManager()

	// Initializing ClusterBridge
	err = a.initClusterBridge()
	if err != nil {
		return err
	}

	// Initializing HTTP server
	err = a.initHTTPServer()
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInstance) Uninit() {
}

func (a *AppInstance) Run() error {

	// HTTP
	go func() {
		err := a.runHTTPServer()
		if err != nil {
			log.Error(err)
		}
	}()

	err := a.runMuxManager()
	if err != nil {
		return err
	}

	<-a.done

	return nil
}
