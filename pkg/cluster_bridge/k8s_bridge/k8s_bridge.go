package k8s_bridge

import (
	app "git.brobridge.com/pilotwave/pilotwave/pkg/app"

	k8sclient "k8s.io/client-go/kubernetes"
)

type K8sBridge struct {
	app       app.App
	stopCh    chan struct{}
	clientset *k8sclient.Clientset
}

func NewK8sBridge(a app.App, clientset *k8sclient.Clientset) *K8sBridge {

	br := new(K8sBridge)
	br.stopCh = make(chan struct{})
	br.clientset = clientset
	br.app = a

	return br
}
