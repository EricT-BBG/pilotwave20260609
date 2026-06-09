package istio_bridge

import (
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/metrics"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	istioResourceGateway               = "gateway"
	istioResourceVirtualService        = "virtualservice"
	istioResourceDestinationRule       = "destinationrule"
	istioResourceAuthorizationPolicy   = "authorizationpolicy"
	istioResourceRequestAuthentication = "requestauthentication"

	kubernetesWriteVerbCreate = "create"
	kubernetesWriteVerbUpdate = "update"
	kubernetesWriteVerbPatch  = "patch"
	kubernetesWriteVerbDelete = "delete"

	kubernetesWriteResultSuccess  = "success"
	kubernetesWriteResultError    = "error"
	kubernetesWriteResultConflict = "conflict"
)

func recordIstioWrite(resource string, verb string, start time.Time, err error) {
	result := kubernetesWriteResultSuccess
	if err != nil {
		result = kubernetesWriteResultError
		if k8serrors.IsConflict(err) {
			result = kubernetesWriteResultConflict
			metrics.RecordKubernetesWriteConflict(resource, verb)
		}
	}

	metrics.RecordKubernetesWrite(resource, verb, result, start)
}

func recordIstioStaleConflict(resource string, verb string) {
	metrics.RecordKubernetesWriteConflict(resource, verb)
	metrics.RecordKubernetesWrite(resource, verb, kubernetesWriteResultConflict, time.Now())
}
