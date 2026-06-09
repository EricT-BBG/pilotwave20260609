package gateway

import (
	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
)

type Gateway struct {
	app app.App
}

func NewGateway(a app.App) *Gateway {
	return &Gateway{
		app: a,
	}
}
