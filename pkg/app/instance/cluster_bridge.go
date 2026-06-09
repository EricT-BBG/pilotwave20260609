package instance

import (
	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	//	log "github.com/sirupsen/logrus"
)

func (a *AppInstance) initClusterBridge() error {

	// Initializing Cluster-Bridge
	err := a.clusterBridge.Init()
	if err != nil {
		return err
	}

	return nil
}

func (a *AppInstance) GetClusterBridge() cluster_bridge.Bridge {
	return cluster_bridge.Bridge(a.clusterBridge)
}

/*
func (a *AppInstance) runClusterBridge() error {
	err := a.clusterBridge.StartWatcher()
	if err != nil {
		log.Error(err)
		return err
	}

	return err
}

func (a *AppInstance) GetNumberOfNodes() (int, error) {
	return a.clusterBridge.GetNumberOfNodes()
}
*/
