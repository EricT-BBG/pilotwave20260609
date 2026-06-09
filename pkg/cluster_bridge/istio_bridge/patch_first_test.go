package istio_bridge

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"
	istioapi "istio.io/api/networking/v1alpha3"
	istiosecurityapi "istio.io/api/security/v1beta1"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurity "istio.io/client-go/pkg/apis/security/v1beta1"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
)

func newTestBridge(objects ...runtime.Object) (*IstioBridge, *istiofake.Clientset) {
	client := istiofake.NewSimpleClientset()
	for _, object := range objects {
		switch item := object.(type) {
		case *istionetworking.Gateway:
			if _, err := client.NetworkingV1alpha3().Gateways(item.Namespace).Create(context.TODO(), item, k8smetav1.CreateOptions{}); err != nil {
				panic(err)
			}
		case *istionetworking.VirtualService:
			if _, err := client.NetworkingV1alpha3().VirtualServices(item.Namespace).Create(context.TODO(), item, k8smetav1.CreateOptions{}); err != nil {
				panic(err)
			}
		case *istionetworking.DestinationRule:
			if _, err := client.NetworkingV1alpha3().DestinationRules(item.Namespace).Create(context.TODO(), item, k8smetav1.CreateOptions{}); err != nil {
				panic(err)
			}
		case *istiosecurity.AuthorizationPolicy:
			if _, err := client.SecurityV1beta1().AuthorizationPolicies(item.Namespace).Create(context.TODO(), item, k8smetav1.CreateOptions{}); err != nil {
				panic(err)
			}
		case *istiosecurity.RequestAuthentication:
			if _, err := client.SecurityV1beta1().RequestAuthentications(item.Namespace).Create(context.TODO(), item, k8smetav1.CreateOptions{}); err != nil {
				panic(err)
			}
		default:
			panic("unsupported test object")
		}
	}
	client.ClearActions()
	return NewIstioBridge(nil, client), client
}

func countActions(actions []k8stesting.Action, verb string, resource string) int {
	count := 0
	for _, action := range actions {
		if action.GetVerb() == verb && action.GetResource().Resource == resource {
			count++
		}
	}
	return count
}

func assertPatchFirst(t *testing.T, client *istiofake.Clientset, resource string) {
	t.Helper()
	actions := client.Actions()
	if count := countActions(actions, "update", resource); count != 0 {
		t.Fatalf("expected no full update actions for %s, got %d", resource, count)
	}
	if count := countActions(actions, "patch", resource); count == 0 {
		t.Fatalf("expected patch action for %s", resource)
	}
}

func patchDataForResource(t *testing.T, client *istiofake.Clientset, resource string) []byte {
	t.Helper()
	for _, action := range client.Actions() {
		if action.GetVerb() != "patch" || action.GetResource().Resource != resource {
			continue
		}
		patchAction, ok := action.(k8stesting.PatchAction)
		if !ok {
			t.Fatalf("patch action for %s has unexpected type %T", resource, action)
		}
		return patchAction.GetPatch()
	}
	t.Fatalf("expected patch action for %s", resource)
	return nil
}

func applyJSONMergePatch(target map[string]interface{}, patch map[string]interface{}) map[string]interface{} {
	for key, value := range patch {
		if value == nil {
			delete(target, key)
			continue
		}
		valueMap, valueIsMap := value.(map[string]interface{})
		targetMap, targetIsMap := target[key].(map[string]interface{})
		if valueIsMap && targetIsMap {
			target[key] = applyJSONMergePatch(targetMap, valueMap)
			continue
		}
		target[key] = value
	}
	return target
}

func assertPatchPreservesCustomMetadataAndSpec(t *testing.T, client *istiofake.Clientset, resource string) {
	t.Helper()
	data := patchDataForResource(t, client, resource)

	var patch map[string]interface{}
	if err := json.Unmarshal(data, &patch); err != nil {
		t.Fatalf("invalid patch json for %s: %v", resource, err)
	}

	object := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels":      map[string]interface{}{"pilotwave.io/custom-label": "keep"},
			"annotations": map[string]interface{}{"pilotwave.io/custom-annotation": "keep"},
		},
		"spec": map[string]interface{}{
			"pilotwave.io/customSpec": map[string]interface{}{"nested": "keep"},
			"pilotwave.io/customList": []interface{}{"keep"},
		},
	}
	merged := applyJSONMergePatch(object, patch)

	metadata := merged["metadata"].(map[string]interface{})
	labels := metadata["labels"].(map[string]interface{})
	if labels["pilotwave.io/custom-label"] != "keep" {
		t.Fatalf("custom metadata label was not preserved for %s: %v", resource, metadata)
	}
	annotations := metadata["annotations"].(map[string]interface{})
	if annotations["pilotwave.io/custom-annotation"] != "keep" {
		t.Fatalf("custom metadata annotation was not preserved for %s: %v", resource, metadata)
	}

	spec := merged["spec"].(map[string]interface{})
	customSpec := spec["pilotwave.io/customSpec"].(map[string]interface{})
	if customSpec["nested"] != "keep" {
		t.Fatalf("custom nested spec field was not preserved for %s: %v", resource, spec)
	}
	if _, ok := spec["pilotwave.io/customList"]; !ok {
		t.Fatalf("custom spec list was not preserved for %s: %v", resource, spec)
	}
}

func assertPatchResourceVersion(t *testing.T, client *istiofake.Clientset, resource string, expected string) {
	t.Helper()
	data := patchDataForResource(t, client, resource)

	var patch map[string]interface{}
	if err := json.Unmarshal(data, &patch); err != nil {
		t.Fatalf("invalid patch json for %s: %v", resource, err)
	}
	metadata, ok := patch["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected metadata in patch for %s: %v", resource, patch)
	}
	if metadata["resourceVersion"] != expected {
		t.Fatalf("expected resourceVersion %q in patch for %s, got %v", expected, resource, metadata["resourceVersion"])
	}
}

func TestMergePatchPreservesUnknownSpecFields(t *testing.T) {
	tests := []struct {
		name      string
		specPatch map[string]interface{}
	}{
		{name: "Gateway", specPatch: map[string]interface{}{"servers": []interface{}{map[string]interface{}{"hosts": []interface{}{"new.example.local"}}}}},
		{name: "VirtualService", specPatch: map[string]interface{}{"hosts": []interface{}{"new.example.local"}}},
		{name: "DestinationRule", specPatch: map[string]interface{}{"subsets": []interface{}{map[string]interface{}{"name": "v2"}}}},
		{name: "AuthorizationPolicy", specPatch: map[string]interface{}{"action": "DENY"}},
		{name: "RequestAuthentication", specPatch: map[string]interface{}{"jwtRules": []interface{}{map[string]interface{}{"issuer": "issuer-a"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := buildMergePatch(map[string]interface{}{"resourceVersion": "2"}, tt.specPatch)
			if err != nil {
				t.Fatalf("buildMergePatch returned error: %v", err)
			}

			var patch map[string]interface{}
			if err := json.Unmarshal(data, &patch); err != nil {
				t.Fatalf("invalid patch json: %v", err)
			}

			object := map[string]interface{}{
				"metadata": map[string]interface{}{"resourceVersion": "2"},
				"spec": map[string]interface{}{
					"pilotwave.io/unknownField": map[string]interface{}{"nested": "kept"},
					"untouchedList":             []interface{}{"keep"},
				},
			}
			merged := applyJSONMergePatch(object, patch)
			spec := merged["spec"].(map[string]interface{})
			unknown := spec["pilotwave.io/unknownField"].(map[string]interface{})
			if unknown["nested"] != "kept" {
				t.Fatalf("unknown nested spec field was not preserved: %v", spec)
			}
			if _, ok := spec["untouchedList"]; !ok {
				t.Fatalf("unknown list spec field was not preserved: %v", spec)
			}
		})
	}
}

func TestUpdateGatewayUsesPatch(t *testing.T) {
	gateway := &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "Gateway", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "gw-a", Namespace: "default"},
		Spec: istioapi.Gateway{
			Selector: map[string]string{"istio": "ingressgateway", "custom": "keep"},
			Servers: []*istioapi.Server{{
				Hosts: []string{"old.example.local"},
				Port:  &istioapi.Port{Name: "http", Number: 80, Protocol: "HTTP"},
			}},
		},
	}
	br, client := newTestBridge(gateway)

	err := br.UpdateGateway("gw-a", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"new.example.local"},
			Ports: []gateway_manager.PortsRequest{{Port: 80, Protocol: "HTTP"}},
		}},
		SelectorMatchLabels: map[string]string{"app": "pilotwave"},
	})
	if err != nil {
		t.Fatalf("UpdateGateway returned error: %v", err)
	}

	assertPatchFirst(t, client, "gateways")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "gateways")
}

func TestCreateGatewayDefaultsSelectorWhenRequestSelectorIsEmpty(t *testing.T) {
	br, client := newTestBridge()

	err := br.CreateGateway("gw-default", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"default.example.local"},
			Ports: []gateway_manager.PortsRequest{{Port: 80, Protocol: "HTTP"}},
		}},
	})
	if err != nil {
		t.Fatalf("CreateGateway returned error: %v", err)
	}

	gateway, err := client.NetworkingV1alpha3().Gateways("default").Get(context.TODO(), "gw-default", k8smetav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected created gateway: %v", err)
	}
	want := map[string]string{"istio": "ingressgateway"}
	if !reflect.DeepEqual(gateway.Spec.Selector, want) {
		t.Fatalf("selector mismatch: got %v want %v", gateway.Spec.Selector, want)
	}
}

func TestCreateGatewayUsesExplicitSelectorWithoutInjectingIstioLabel(t *testing.T) {
	br, client := newTestBridge()

	err := br.CreateGateway("gw-custom", "default", &gateway_manager.GatewayRequest{
		SelectorMatchLabels: map[string]string{"app": "custom-ingress"},
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"custom.example.local"},
			Ports: []gateway_manager.PortsRequest{{Port: 80, Protocol: "HTTP"}},
		}},
	})
	if err != nil {
		t.Fatalf("CreateGateway returned error: %v", err)
	}

	gateway, err := client.NetworkingV1alpha3().Gateways("default").Get(context.TODO(), "gw-custom", k8smetav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected created gateway: %v", err)
	}
	want := map[string]string{"app": "custom-ingress"}
	if !reflect.DeepEqual(gateway.Spec.Selector, want) {
		t.Fatalf("selector mismatch: got %v want %v", gateway.Spec.Selector, want)
	}
}

func TestUpdateRouterUsesPatch(t *testing.T) {
	vs := &istionetworking.VirtualService{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "rt-a", Namespace: "default", ResourceVersion: "2", Labels: map[string]string{"owner": "custom"}},
		Spec:       istioapi.VirtualService{Hosts: []string{"old.example.local"}},
	}
	br, client := newTestBridge(vs)

	if err := br.UpdateRouter("rt-a", "default", "http", []string{"new.example.local"}, ""); err != nil {
		t.Fatalf("UpdateRouter returned error: %v", err)
	}

	assertPatchFirst(t, client, "virtualservices")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "virtualservices")
	assertPatchResourceVersion(t, client, "virtualservices", "2")
}

func TestUpdateRouterRuleUsesPatch(t *testing.T) {
	vs := &istionetworking.VirtualService{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "rt-a", Namespace: "default", ResourceVersion: "2"},
		Spec:       istioapi.VirtualService{Hosts: []string{"app.example.local"}},
	}
	br, client := newTestBridge(vs)

	err := br.UpdateRouterRule("rt-a", "default", router_manager.RouterRuleRequest{
		Https: []router_manager.HttpsData{{
			Prefixs:      []string{"/"},
			Destinations: []router_manager.DestinationData{{Host: "svc.default.svc.cluster.local", Port: 80}},
		}},
	})
	if err != nil {
		t.Fatalf("UpdateRouterRule returned error: %v", err)
	}

	assertPatchFirst(t, client, "virtualservices")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "virtualservices")
	assertPatchResourceVersion(t, client, "virtualservices", "2")
}

func TestRouterGatewayMappingUsesPatch(t *testing.T) {
	gateway := &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "Gateway", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "gw-a", Namespace: "default"},
	}
	vs := &istionetworking.VirtualService{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "rt-a", Namespace: "default", ResourceVersion: "2"},
	}
	br, client := newTestBridge(gateway, vs)

	err := br.UpdateRouterGatewayMapping("rt-a", "default", []router_manager.RouterMappingGatewayData{{Name: "gw-a", Namespace: "default"}}, "")
	if err != nil {
		t.Fatalf("UpdateRouterGatewayMapping returned error: %v", err)
	}

	assertPatchFirst(t, client, "virtualservices")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "virtualservices")
	assertPatchResourceVersion(t, client, "virtualservices", "2")
}

func TestGatewayRouterMappingUsesPatch(t *testing.T) {
	gateway := &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "Gateway", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "gw-a", Namespace: "default"},
	}
	vs := &istionetworking.VirtualService{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "rt-a", Namespace: "default", ResourceVersion: "2"},
	}
	br, client := newTestBridge(gateway, vs)

	err := br.UpdateGatewayRouterMapping("gw-a", "default", []gateway_manager.GatewayMappinRouterData{{Name: "rt-a", Namespace: "default"}}, nil)
	if err != nil {
		t.Fatalf("UpdateGatewayRouterMapping returned error: %v", err)
	}

	assertPatchFirst(t, client, "virtualservices")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "virtualservices")
	assertPatchResourceVersion(t, client, "virtualservices", "2")
}

func TestUpdateDestinationRuleUsesPatch(t *testing.T) {
	dr := &istionetworking.DestinationRule{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "DestinationRule", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "rt-a", Namespace: "default"},
		Spec: istioapi.DestinationRule{
			Host:    "rt-a",
			Subsets: []*istioapi.Subset{{Name: "v1", Labels: map[string]string{"version": "v1"}}},
		},
	}
	br, client := newTestBridge(dr)

	if err := br.updateDestinationRule("rt-a", "default", []string{"v2"}); err != nil {
		t.Fatalf("updateDestinationRule returned error: %v", err)
	}

	assertPatchFirst(t, client, "destinationrules")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "destinationrules")
}

func TestUpdateAuthorizationPolicyUsesPatch(t *testing.T) {
	policy := &istiosecurity.AuthorizationPolicy{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "AuthorizationPolicy", APIVersion: "security.istio.io/v1beta1"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "allow-a", Namespace: "default"},
		Spec:       istiosecurityapi.AuthorizationPolicy{Action: istiosecurityapi.AuthorizationPolicy_ALLOW},
	}
	br, client := newTestBridge(policy)

	err := br.UpdateAuthorizationPolicy("allow-a", "default", &security_manager.AuthorizationPolicyUpdateRequest{
		Action: "deny",
		Rules:  []*security_manager.AuthorizationPolicyRuleData{},
	})
	if err != nil {
		t.Fatalf("UpdateAuthorizationPolicy returned error: %v", err)
	}

	assertPatchFirst(t, client, "authorizationpolicies")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "authorizationpolicies")
}

func TestUpdateRequestAuthenticationUsesPatch(t *testing.T) {
	requestAuth := &istiosecurity.RequestAuthentication{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "RequestAuthentication", APIVersion: "security.istio.io/v1beta1"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "jwt-a", Namespace: "default"},
		Spec:       istiosecurityapi.RequestAuthentication{},
	}
	br, client := newTestBridge(requestAuth)

	err := br.UpdateRequestAuthentication("jwt-a", "default", &security_manager.RequestAuthenticationUpdateRequest{
		JWTRules: []security_manager.RequestAuthenticationJWTData{{Issuer: "issuer-a", JwksUri: "https://example.local/jwks.json"}},
	})
	if err != nil {
		t.Fatalf("UpdateRequestAuthentication returned error: %v", err)
	}

	assertPatchFirst(t, client, "requestauthentications")
	assertPatchPreservesCustomMetadataAndSpec(t, client, "requestauthentications")
}

func TestUpdateGatewayStaleResourceVersionReturnsConflictWithoutPatch(t *testing.T) {
	currentRV := "2"
	staleRV := "1"
	gateway := &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "Gateway", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "gw-a", Namespace: "default", ResourceVersion: currentRV},
		Spec: istioapi.Gateway{
			Selector: map[string]string{"istio": "ingressgateway"},
			Servers: []*istioapi.Server{{
				Hosts: []string{"old.example.local"},
				Port:  &istioapi.Port{Name: "http", Number: 80, Protocol: "HTTP"},
			}},
		},
	}
	br, client := newTestBridge(gateway)

	err := br.UpdateGateway("gw-a", "default", &gateway_manager.GatewayRequest{
		ResourceVersion: &staleRV,
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"new.example.local"},
			Ports: []gateway_manager.PortsRequest{{Port: 80, Protocol: "HTTP"}},
		}},
	})
	if !k8serrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if count := countActions(client.Actions(), "patch", "gateways"); count != 0 {
		t.Fatalf("expected no gateway patch on stale resourceVersion, got %d", count)
	}
}

func TestUpdateRouterStaleResourceVersionReturnsConflictWithoutPatch(t *testing.T) {
	vs := &istionetworking.VirtualService{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "rt-a", Namespace: "default", ResourceVersion: "2"},
		Spec:       istioapi.VirtualService{Hosts: []string{"old.example.local"}},
	}
	br, client := newTestBridge(vs)

	err := br.UpdateRouter("rt-a", "default", "http", []string{"new.example.local"}, "1")
	if !k8serrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if count := countActions(client.Actions(), "patch", "virtualservices"); count != 0 {
		t.Fatalf("expected no virtualservice patch on stale resourceVersion, got %d", count)
	}
}

func TestUpdateAuthorizationPolicyStaleResourceVersionReturnsConflictWithoutPatch(t *testing.T) {
	staleRV := "1"
	policy := &istiosecurity.AuthorizationPolicy{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "AuthorizationPolicy", APIVersion: "security.istio.io/v1beta1"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "allow-a", Namespace: "default", ResourceVersion: "2"},
		Spec:       istiosecurityapi.AuthorizationPolicy{Action: istiosecurityapi.AuthorizationPolicy_ALLOW},
	}
	br, client := newTestBridge(policy)

	err := br.UpdateAuthorizationPolicy("allow-a", "default", &security_manager.AuthorizationPolicyUpdateRequest{
		ResourceVersion: &staleRV,
		Action:          "allow",
		Rules:           []*security_manager.AuthorizationPolicyRuleData{},
	})
	if !k8serrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if count := countActions(client.Actions(), "patch", "authorizationpolicies"); count != 0 {
		t.Fatalf("expected no authorizationpolicy patch on stale resourceVersion, got %d", count)
	}
}

func TestUpdateRequestAuthenticationStaleResourceVersionReturnsConflictWithoutPatch(t *testing.T) {
	staleRV := "1"
	requestAuth := &istiosecurity.RequestAuthentication{
		TypeMeta:   k8smetav1.TypeMeta{Kind: "RequestAuthentication", APIVersion: "security.istio.io/v1beta1"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "jwt-a", Namespace: "default", ResourceVersion: "2"},
		Spec:       istiosecurityapi.RequestAuthentication{},
	}
	br, client := newTestBridge(requestAuth)

	err := br.UpdateRequestAuthentication("jwt-a", "default", &security_manager.RequestAuthenticationUpdateRequest{
		ResourceVersion: &staleRV,
		JWTRules:        []security_manager.RequestAuthenticationJWTData{{Issuer: "issuer-a", JwksUri: "https://example.local/jwks.json"}},
	})
	if !k8serrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if count := countActions(client.Actions(), "patch", "requestauthentications"); count != 0 {
		t.Fatalf("expected no requestauthentication patch on stale resourceVersion, got %d", count)
	}
}
