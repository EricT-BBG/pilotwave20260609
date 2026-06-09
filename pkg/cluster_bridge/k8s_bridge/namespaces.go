package k8s_bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

const (
	istioInjectionLabel = "istio-injection"
	istioRevisionLabel  = "istio.io/rev"
)

func (br *K8sBridge) GetNamespaces() ([]string, error) {

	namespaces := make([]string, 0)

	listOpts := k8smetav1.ListOptions{}
	nList, err := br.clientset.CoreV1().Namespaces().List(context.TODO(), listOpts)
	if err != nil {
		return namespaces, err
	}

	for _, namespace := range nList.Items {
		//Get namespace name
		namespaces = append(namespaces, namespace.Name)
	}

	sort.Strings(namespaces)

	return namespaces, nil
}

func (br *K8sBridge) GetNamespaceMetadata() ([]cluster_bridge.NamespaceMetadata, error) {

	namespaces := make([]cluster_bridge.NamespaceMetadata, 0)

	listOpts := k8smetav1.ListOptions{}
	nList, err := br.clientset.CoreV1().Namespaces().List(context.TODO(), listOpts)
	if err != nil {
		return namespaces, err
	}

	for _, namespace := range nList.Items {
		namespaces = append(namespaces, namespaceMetadata(namespace.Name, namespace.Labels, namespace.ResourceVersion))
	}

	sort.Slice(namespaces, func(i int, j int) bool {
		return namespaces[i].Name < namespaces[j].Name
	})

	return namespaces, nil
}

func (br *K8sBridge) PatchNamespaceIstioInjection(name string, mode string, revision string) (cluster_bridge.NamespaceMetadata, error) {

	labels, err := namespaceInjectionPatchLabels(mode, revision)
	if err != nil {
		return cluster_bridge.NamespaceMetadata{}, err
	}

	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": labels,
		},
	}

	patchData, err := json.Marshal(patch)
	if err != nil {
		return cluster_bridge.NamespaceMetadata{}, err
	}

	start := time.Now()
	namespace, err := br.clientset.CoreV1().Namespaces().Patch(context.TODO(), name, k8stypes.MergePatchType, patchData, k8smetav1.PatchOptions{})
	recordK8sWrite(k8sResourceNamespace, kubernetesWriteVerbPatch, start, err)
	if err != nil {
		return cluster_bridge.NamespaceMetadata{}, err
	}

	return namespaceMetadata(namespace.Name, namespace.Labels, namespace.ResourceVersion), nil
}

func namespaceMetadata(name string, labels map[string]string, resourceVersion string) cluster_bridge.NamespaceMetadata {

	labelCopy := make(map[string]string, len(labels))
	for key, value := range labels {
		labelCopy[key] = value
	}

	return cluster_bridge.NamespaceMetadata{
		Name:            name,
		IstioInjection:  labels[istioInjectionLabel],
		IstioRevision:   labels[istioRevisionLabel],
		SystemNamespace: cluster_bridge.IsSystemNamespaceName(name),
		Labels:          labelCopy,
		ResourceVersion: resourceVersion,
	}
}

func namespaceInjectionPatchLabels(mode string, revision string) (map[string]interface{}, error) {

	switch mode {
	case cluster_bridge.NamespaceIstioInjectionModeDisabled:
		return map[string]interface{}{
			istioInjectionLabel: cluster_bridge.NamespaceIstioInjectionModeDisabled,
			istioRevisionLabel:  nil,
		}, nil
	case cluster_bridge.NamespaceIstioInjectionModeEnabled:
		return map[string]interface{}{
			istioInjectionLabel: cluster_bridge.NamespaceIstioInjectionModeEnabled,
			istioRevisionLabel:  nil,
		}, nil
	case cluster_bridge.NamespaceIstioInjectionModeRevision:
		return map[string]interface{}{
			istioInjectionLabel: nil,
			istioRevisionLabel:  revision,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported namespace Istio injection mode: %s", mode)
	}
}
