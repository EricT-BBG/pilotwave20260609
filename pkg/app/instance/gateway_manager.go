package instance

import "git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"

func (a *AppInstance) initGatewayManager() error {
	return nil
}

func (a *AppInstance) GetGateway() gateway_manager.Gateway {
	return gateway_manager.Gateway(a.gateway)
}
