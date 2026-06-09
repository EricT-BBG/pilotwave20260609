package cluster_bridge

import (
	"context"
	"sort"
	"strings"

	cluster_bridge_types "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	k8sadmissionv1 "k8s.io/api/admissionregistration/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	k8sclient "k8s.io/client-go/kubernetes"
)

var requiredIstioCRDs = []string{
	"gateways.networking.istio.io",
	"virtualservices.networking.istio.io",
	"destinationrules.networking.istio.io",
	"authorizationpolicies.security.istio.io",
	"requestauthentications.security.istio.io",
}

func disabledIstioCapabilities() cluster_bridge_types.IstioCapabilities {
	return cluster_bridge_types.IstioCapabilities{
		Installed: false,
		Disabled:  true,
		Message:   "Cluster bridge is disabled for local development",
	}
}

func detectIstioCapabilities(discoveryClient discovery.DiscoveryInterface) cluster_bridge_types.IstioCapabilities {
	resources, err := discoveryClient.ServerPreferredResources()
	if err != nil && !isPartialDiscoveryError(err) {
		return cluster_bridge_types.IstioCapabilities{
			Installed: false,
			Message:   "Unable to detect Istio CRDs in this cluster: " + err.Error(),
		}
	}

	return buildIstioCapabilities(resources)
}

func detectIstioCapabilitiesWithClient(k8sClient k8sclient.Interface) cluster_bridge_types.IstioCapabilities {
	capabilities := detectIstioCapabilities(k8sClient.Discovery())
	if capabilities.Installed {
		detectIstioInjectionCapabilities(context.Background(), k8sClient, &capabilities)
	}
	return capabilities
}

func isPartialDiscoveryError(err error) bool {
	_, ok := err.(*discovery.ErrGroupDiscoveryFailed)
	return ok || meta.IsNoMatchError(err)
}

func buildIstioCapabilities(resources []*k8smetav1.APIResourceList) cluster_bridge_types.IstioCapabilities {
	available := map[string]struct{}{}
	for _, resourceList := range resources {
		groupVersion, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil || !isIstioAPIGroup(groupVersion.Group) {
			continue
		}

		for _, resource := range resourceList.APIResources {
			available[resource.Name+"."+groupVersion.Group] = struct{}{}
		}
	}

	capabilities := cluster_bridge_types.IstioCapabilities{
		AvailableCRDs: make([]string, 0, len(requiredIstioCRDs)),
		MissingCRDs:   make([]string, 0),
	}
	for _, requiredCRD := range requiredIstioCRDs {
		if _, ok := available[requiredCRD]; ok {
			capabilities.AvailableCRDs = append(capabilities.AvailableCRDs, requiredCRD)
			continue
		}

		capabilities.MissingCRDs = append(capabilities.MissingCRDs, requiredCRD)
	}

	capabilities.Installed = len(capabilities.MissingCRDs) == 0
	if capabilities.Installed {
		capabilities.Message = "Istio CRDs are installed in this cluster"
	} else {
		capabilities.Message = "Istio CRDs are not installed or incomplete in this cluster"
	}

	return capabilities
}

func isIstioAPIGroup(group string) bool {
	return group == "networking.istio.io" || group == "security.istio.io"
}

func detectIstioInjectionCapabilities(ctx context.Context, k8sClient k8sclient.Interface, capabilities *cluster_bridge_types.IstioCapabilities) {
	revisions := map[string]struct{}{}
	revisionTags := map[string]string{}

	webhooks, err := k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, k8smetav1.ListOptions{})
	if err != nil {
		capabilities.RevisionDetectionMessage = "Unable to detect Istio injection webhooks: " + err.Error()
		detectIstioRevisionDeployments(ctx, k8sClient, revisions)
		applyDetectedIstioRevisions(capabilities, revisions, revisionTags)
		return
	}

	for i := range webhooks.Items {
		webhook := &webhooks.Items[i]
		labels := webhook.GetLabels()
		revision := strings.TrimSpace(labels["istio.io/rev"])
		tag := strings.TrimSpace(labels["istio.io/tag"])

		if isDefaultInjectionWebhook(webhook) || revision == "default" || tag == "default" {
			capabilities.DefaultInjectionAvailable = true
		}

		if tag != "" && tag != "default" {
			revisionTags[tag] = revision
			continue
		}

		if revision != "" && revision != "default" {
			revisions[revision] = struct{}{}
		}
	}

	detectIstioRevisionDeployments(ctx, k8sClient, revisions)
	applyDetectedIstioRevisions(capabilities, revisions, revisionTags)
}

func isDefaultInjectionWebhook(config *k8sadmissionv1.MutatingWebhookConfiguration) bool {
	if config == nil {
		return false
	}

	for _, webhook := range config.Webhooks {
		if webhook.NamespaceSelector == nil {
			continue
		}
		if webhook.NamespaceSelector.MatchLabels["istio-injection"] == "enabled" {
			return true
		}
	}

	return false
}

func detectIstioRevisionDeployments(ctx context.Context, k8sClient k8sclient.Interface, revisions map[string]struct{}) {
	deployments, err := k8sClient.AppsV1().Deployments("istio-system").List(ctx, k8smetav1.ListOptions{
		LabelSelector: "app=istiod",
	})
	if err != nil {
		return
	}

	for i := range deployments.Items {
		revision := deploymentIstioRevision(&deployments.Items[i])
		if revision == "" || revision == "default" {
			continue
		}
		revisions[revision] = struct{}{}
	}
}

func deploymentIstioRevision(deployment *k8sappsv1.Deployment) string {
	if deployment == nil {
		return ""
	}

	if revision := strings.TrimSpace(deployment.Labels["istio.io/rev"]); revision != "" {
		return revision
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == "REVISION" {
				return strings.TrimSpace(env.Value)
			}
		}
	}

	return ""
}

func applyDetectedIstioRevisions(capabilities *cluster_bridge_types.IstioCapabilities, revisions map[string]struct{}, revisionTags map[string]string) {
	capabilities.Revisions = sortedMapKeys(revisions)
	capabilities.RevisionTags = sortedRevisionTags(revisionTags)
	capabilities.RevisionInjectionAvailable = len(capabilities.Revisions) > 0 || len(capabilities.RevisionTags) > 0
}

func sortedMapKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sortedRevisionTags(values map[string]string) []cluster_bridge_types.IstioRevisionTag {
	result := make([]cluster_bridge_types.IstioRevisionTag, 0, len(values))
	for tag, revision := range values {
		result = append(result, cluster_bridge_types.IstioRevisionTag{
			Name:     tag,
			Revision: revision,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
