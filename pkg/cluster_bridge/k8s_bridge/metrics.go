package k8s_bridge

import (
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/metrics"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	k8sResourceNamespace = "namespace"
	k8sResourceSecret    = "secret"

	kubernetesWriteVerbCreate = "create"
	kubernetesWriteVerbPatch  = "patch"
	kubernetesWriteVerbDelete = "delete"

	kubernetesWriteResultSuccess  = "success"
	kubernetesWriteResultError    = "error"
	kubernetesWriteResultConflict = "conflict"
)

func recordK8sWrite(resource string, verb string, start time.Time, err error) {
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
