package instance

import "git.brobridge.com/pilotwave/pilotwave/pkg/user_manager"

func (a *AppInstance) initUserManager() error {
	return nil
}

func (a *AppInstance) GetUser() user_manager.User {
	return user_manager.User(a.user)
}
