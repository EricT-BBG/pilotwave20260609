package api

import (
	"strings"
	"testing"

	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
)

func TestValidateNamespaceIstioInjectionPatchRequiresSystemConfirmation(t *testing.T) {
	err := validateNamespaceIstioInjectionPatch("kube-system", PatchNamespaceIstioInjectionRequest{
		Mode: cluster_bridge.NamespaceIstioInjectionModeEnabled,
	})
	if err == nil || !strings.Contains(err.Error(), "explicit confirmation") {
		t.Fatalf("expected explicit confirmation error for system namespace, got %v", err)
	}

	err = validateNamespaceIstioInjectionPatch("kube-system", PatchNamespaceIstioInjectionRequest{
		Mode:                 cluster_bridge.NamespaceIstioInjectionModeEnabled,
		AllowSystemNamespace: true,
	})
	if err != nil {
		t.Fatalf("expected confirmed system namespace change to pass, got %v", err)
	}
}

func TestValidateNamespaceIstioInjectionPatchAllowsDefaultNamespace(t *testing.T) {
	err := validateNamespaceIstioInjectionPatch("default", PatchNamespaceIstioInjectionRequest{
		Mode: cluster_bridge.NamespaceIstioInjectionModeEnabled,
	})
	if err != nil {
		t.Fatalf("expected default namespace to pass without system confirmation, got %v", err)
	}
}
