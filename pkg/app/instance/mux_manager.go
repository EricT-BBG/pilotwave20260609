package instance

import mux_manager "git.brobridge.com/pilotwave/pilotwave/pkg/mux_manager"

func (a *AppInstance) initMuxManager() error {
	return nil
}

func (a *AppInstance) runMuxManager() error {
	return a.muxManager.Serve()
}

func (a *AppInstance) GetMuxManager() mux_manager.Manager {
	return mux_manager.Manager(a.muxManager)
}
