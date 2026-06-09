package cluster_bridge

import (
	"strings"
	"testing"

	cluster_bridge_types "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
)

func TestIstioResourceMethodsReturnTypedUnavailableWhenClusterBridgeDisabled(t *testing.T) {
	bridge := &ClusterBridge{
		disabled:          true,
		istioCapabilities: disabledIstioCapabilities(),
	}

	err := bridge.CreateGateway("edge", "default", &gateway_manager.GatewayRequest{})
	if !cluster_bridge_types.IsIstioUnavailable(err) {
		t.Fatalf("expected typed Istio unavailable error, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "local development") {
		t.Fatalf("expected disabled local dev message, got %q", err.Error())
	}
}

func TestIstioResourceMethodsReturnTypedUnavailableWhenIstioCRDsMissing(t *testing.T) {
	bridge := &ClusterBridge{
		istioCapabilities: cluster_bridge_types.IstioCapabilities{
			Installed:   false,
			MissingCRDs: []string{"gateways.networking.istio.io"},
			Message:     "Istio CRDs are not installed or incomplete in this cluster",
		},
	}

	err := bridge.UpdateRouterRule("app", "default", router_manager.RouterRuleRequest{})
	if !cluster_bridge_types.IsIstioUnavailable(err) {
		t.Fatalf("expected typed Istio unavailable error, got %T %v", err, err)
	}
	if !strings.Contains(err.Error(), "missing: gateways.networking.istio.io") {
		t.Fatalf("expected missing CRD detail, got %q", err.Error())
	}
}
