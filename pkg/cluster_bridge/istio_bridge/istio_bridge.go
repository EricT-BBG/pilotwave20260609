package istio_bridge

import (
	app "git.brobridge.com/pilotwave/pilotwave/pkg/app"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
)

type IstioBridge struct {
	app               app.App
	stopCh            chan struct{}
	gatewayResources  GatewayResourceClient
	routerResources   RouterResourceClient
	securityResources SecurityResourceClient
}

func NewIstioBridge(a app.App, clientset istioclient.Interface) *IstioBridge {

	br := new(IstioBridge)
	br.stopCh = make(chan struct{})
	resources := newIstioResourceClient(clientset)
	br.gatewayResources = resources
	br.routerResources = resources
	br.securityResources = resources
	br.app = a

	return br
}
