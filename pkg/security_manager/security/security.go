package security

import (
	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
)

type Security struct {
	app app.App
}

func NewSecurity(a app.App) *Security {
	return &Security{
		app: a,
	}
}
