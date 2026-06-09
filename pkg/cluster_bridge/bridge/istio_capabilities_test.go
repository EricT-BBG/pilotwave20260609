package cluster_bridge

import (
	"context"
	"reflect"
	"testing"

	cluster_bridge_types "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	k8sadmissionv1 "k8s.io/api/admissionregistration/v1"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildIstioCapabilitiesReportsInstalledWhenAllRequiredCRDsAreAvailable(t *testing.T) {
	capabilities := buildIstioCapabilities([]*v1.APIResourceList{
		{
			GroupVersion: "networking.istio.io/v1alpha3",
			APIResources: []v1.APIResource{
				{Name: "gateways"},
				{Name: "virtualservices"},
				{Name: "destinationrules"},
			},
		},
		{
			GroupVersion: "security.istio.io/v1beta1",
			APIResources: []v1.APIResource{
				{Name: "authorizationpolicies"},
				{Name: "requestauthentications"},
			},
		},
	})

	if !capabilities.Installed {
		t.Fatalf("expected Istio to be installed, got %#v", capabilities)
	}
	if capabilities.Disabled {
		t.Fatalf("expected disabled=false, got %#v", capabilities)
	}
	if len(capabilities.MissingCRDs) != 0 {
		t.Fatalf("expected no missing CRDs, got %v", capabilities.MissingCRDs)
	}
	if !reflect.DeepEqual(capabilities.AvailableCRDs, requiredIstioCRDs) {
		t.Fatalf("available CRDs mismatch: got %v want %v", capabilities.AvailableCRDs, requiredIstioCRDs)
	}
}

func TestBuildIstioCapabilitiesReportsMissingCRDs(t *testing.T) {
	capabilities := buildIstioCapabilities([]*v1.APIResourceList{
		{
			GroupVersion: "networking.istio.io/v1alpha3",
			APIResources: []v1.APIResource{
				{Name: "gateways"},
				{Name: "virtualservices"},
			},
		},
	})

	wantMissing := []string{
		"destinationrules.networking.istio.io",
		"authorizationpolicies.security.istio.io",
		"requestauthentications.security.istio.io",
	}
	if capabilities.Installed {
		t.Fatalf("expected Istio to be incomplete, got %#v", capabilities)
	}
	if !reflect.DeepEqual(capabilities.MissingCRDs, wantMissing) {
		t.Fatalf("missing CRDs mismatch: got %v want %v", capabilities.MissingCRDs, wantMissing)
	}
	if capabilities.Message != "Istio CRDs are not installed or incomplete in this cluster" {
		t.Fatalf("unexpected message: %q", capabilities.Message)
	}
}

func TestDisabledIstioCapabilities(t *testing.T) {
	capabilities := disabledIstioCapabilities()

	want := cluster_bridge_types.IstioCapabilities{
		Installed: false,
		Disabled:  true,
		Message:   "Cluster bridge is disabled for local development",
	}
	if !reflect.DeepEqual(capabilities, want) {
		t.Fatalf("disabled capabilities mismatch: got %#v want %#v", capabilities, want)
	}
}

func TestDetectIstioInjectionCapabilitiesFindsDefaultInjector(t *testing.T) {
	client := fake.NewSimpleClientset(&k8sadmissionv1.MutatingWebhookConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: "istio-sidecar-injector",
			Labels: map[string]string{
				"istio.io/rev": "default",
			},
		},
		Webhooks: []k8sadmissionv1.MutatingWebhook{
			{
				Name: "sidecar-injector.istio.io",
				NamespaceSelector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"istio-injection": "enabled",
					},
				},
			},
		},
	})

	capabilities := cluster_bridge_types.IstioCapabilities{Installed: true}
	detectIstioInjectionCapabilities(context.Background(), client, &capabilities)

	if !capabilities.DefaultInjectionAvailable {
		t.Fatalf("expected default injection to be available: %#v", capabilities)
	}
	if capabilities.RevisionInjectionAvailable {
		t.Fatalf("did not expect revision injection for default-only injector: %#v", capabilities)
	}
	if len(capabilities.Revisions) != 0 {
		t.Fatalf("did not expect default revision in advanced revision list: %v", capabilities.Revisions)
	}
}

func TestDetectIstioInjectionCapabilitiesFindsRevisionsAndTags(t *testing.T) {
	client := fake.NewSimpleClientset(
		&k8sadmissionv1.MutatingWebhookConfiguration{
			ObjectMeta: v1.ObjectMeta{
				Name: "istio-revision-1-22-0",
				Labels: map[string]string{
					"istio.io/rev": "1-22-0",
				},
			},
		},
		&k8sadmissionv1.MutatingWebhookConfiguration{
			ObjectMeta: v1.ObjectMeta{
				Name: "istio-revision-tag-prod-stable",
				Labels: map[string]string{
					"istio.io/rev": "1-22-0",
					"istio.io/tag": "prod-stable",
				},
			},
		},
	)

	capabilities := cluster_bridge_types.IstioCapabilities{Installed: true}
	detectIstioInjectionCapabilities(context.Background(), client, &capabilities)

	if !capabilities.RevisionInjectionAvailable {
		t.Fatalf("expected revision injection to be available: %#v", capabilities)
	}
	if !reflect.DeepEqual(capabilities.Revisions, []string{"1-22-0"}) {
		t.Fatalf("revision list mismatch: %v", capabilities.Revisions)
	}
	wantTags := []cluster_bridge_types.IstioRevisionTag{
		{Name: "prod-stable", Revision: "1-22-0"},
	}
	if !reflect.DeepEqual(capabilities.RevisionTags, wantTags) {
		t.Fatalf("revision tags mismatch: got %#v want %#v", capabilities.RevisionTags, wantTags)
	}
}

func TestDetectIstioInjectionCapabilitiesFallsBackToIstiodDeploymentRevision(t *testing.T) {
	client := fake.NewSimpleClientset(&k8sappsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "istiod-canary",
			Namespace: "istio-system",
			Labels: map[string]string{
				"app":          "istiod",
				"istio.io/rev": "canary",
			},
		},
		Spec: k8sappsv1.DeploymentSpec{
			Template: k8scorev1.PodTemplateSpec{
				Spec: k8scorev1.PodSpec{
					Containers: []k8scorev1.Container{{Name: "discovery"}},
				},
			},
		},
	})

	capabilities := cluster_bridge_types.IstioCapabilities{Installed: true}
	detectIstioInjectionCapabilities(context.Background(), client, &capabilities)

	if !capabilities.RevisionInjectionAvailable {
		t.Fatalf("expected revision injection from istiod deployment: %#v", capabilities)
	}
	if !reflect.DeepEqual(capabilities.Revisions, []string{"canary"}) {
		t.Fatalf("revision list mismatch: %v", capabilities.Revisions)
	}
}
