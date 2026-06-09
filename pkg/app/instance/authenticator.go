package instance

import "git.brobridge.com/pilotwave/pilotwave/pkg/auth"

func (a *AppInstance) initAuthenticator() error {
	return nil
}

func (a *AppInstance) GetAuthenticator() auth.Authenticator {
	return auth.Authenticator(a.authenticator)
}
