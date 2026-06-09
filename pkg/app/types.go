package app

import (
	"git.brobridge.com/pilotwave/pilotwave/pkg/auth"
	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/http_server"
	"git.brobridge.com/pilotwave/pilotwave/pkg/mux_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/user_manager"
	"github.com/jinzhu/gorm"
)

type App interface {
	GetHTTPServer() http_server.Server
	GetMuxManager() mux_manager.Manager
	GetAuthenticator() auth.Authenticator
	GetUser() user_manager.User
	GetGateway() gateway_manager.Gateway
	GetSecurity() security_manager.Security
	GetRouter() router_manager.Router
	GetDatabase() *gorm.DB
	GetClusterBridge() cluster_bridge.Bridge
}
