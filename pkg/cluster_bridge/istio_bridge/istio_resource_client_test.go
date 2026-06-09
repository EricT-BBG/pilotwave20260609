package istio_bridge

import (
	"context"
	"encoding/json"
	"testing"

	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var _ IstioResourceClient = (*clientsetIstioResourceClient)(nil)

type recordingIstioResourceClient struct {
	IstioResourceClient

	patchGatewayNamespace string
	patchGatewayName      string
	patchGatewayType      k8stypes.PatchType
	patchGatewayData      []byte
}

func (c *recordingIstioResourceClient) PatchGateway(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.Gateway, error) {
	c.patchGatewayNamespace = namespace
	c.patchGatewayName = name
	c.patchGatewayType = patchType
	c.patchGatewayData = data
	return &istionetworking.Gateway{}, nil
}

func TestPatchGatewayUsesIstioResourceClient(t *testing.T) {
	client := &recordingIstioResourceClient{}
	br := &IstioBridge{gatewayResources: client}

	err := br.patchGateway(context.TODO(), "default", "gw-a", map[string]interface{}{
		"resourceVersion": "123",
	}, map[string]interface{}{
		"selector": map[string]string{"istio": "ingressgateway"},
	})
	if err != nil {
		t.Fatalf("patchGateway returned error: %v", err)
	}

	if client.patchGatewayNamespace != "default" {
		t.Fatalf("expected namespace default, got %s", client.patchGatewayNamespace)
	}
	if client.patchGatewayName != "gw-a" {
		t.Fatalf("expected name gw-a, got %s", client.patchGatewayName)
	}
	if client.patchGatewayType != k8stypes.MergePatchType {
		t.Fatalf("expected merge patch type, got %s", client.patchGatewayType)
	}

	var patch map[string]interface{}
	if err := json.Unmarshal(client.patchGatewayData, &patch); err != nil {
		t.Fatalf("invalid patch json: %v", err)
	}
	if _, ok := patch["metadata"]; !ok {
		t.Fatalf("expected metadata in patch: %v", patch)
	}
	if _, ok := patch["spec"]; !ok {
		t.Fatalf("expected spec in patch: %v", patch)
	}
}
