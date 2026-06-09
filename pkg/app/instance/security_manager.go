package instance

import (
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"
)

func (a *AppInstance) initSecurityManager() error {
	return nil
}

func (a *AppInstance) GetSecurity() security_manager.Security {
	return security_manager.Security(a.router)
}
