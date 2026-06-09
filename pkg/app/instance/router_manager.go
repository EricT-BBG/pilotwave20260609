package instance

import "git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"

func (a *AppInstance) initRouterManager() error {
	return nil
}

func (a *AppInstance) GetRouter() router_manager.Router {
	return router_manager.Router(a.router)
}
