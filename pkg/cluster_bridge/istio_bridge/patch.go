package istio_bridge

import (
	"context"
	"encoding/json"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func buildMergePatch(metadata map[string]interface{}, spec map[string]interface{}) ([]byte, error) {
	patch := make(map[string]interface{})
	if len(metadata) > 0 {
		patch["metadata"] = metadata
	}
	if len(spec) > 0 {
		patch["spec"] = spec
	}

	return json.Marshal(patch)
}

func (br *IstioBridge) patchGateway(ctx context.Context, namespace string, name string, metadata map[string]interface{}, spec map[string]interface{}) error {
	data, err := buildMergePatch(metadata, spec)
	if err != nil {
		return err
	}

	_, err = br.gatewayResources.PatchGateway(ctx, namespace, name, k8stypes.MergePatchType, data, k8smetav1.PatchOptions{})
	return err
}

func (br *IstioBridge) patchVirtualService(ctx context.Context, namespace string, name string, metadata map[string]interface{}, spec map[string]interface{}) error {
	data, err := buildMergePatch(metadata, spec)
	if err != nil {
		return err
	}

	_, err = br.routerResources.PatchVirtualService(ctx, namespace, name, k8stypes.MergePatchType, data, k8smetav1.PatchOptions{})
	return err
}

func (br *IstioBridge) patchDestinationRule(ctx context.Context, namespace string, name string, metadata map[string]interface{}, spec map[string]interface{}) error {
	data, err := buildMergePatch(metadata, spec)
	if err != nil {
		return err
	}

	_, err = br.routerResources.PatchDestinationRule(ctx, namespace, name, k8stypes.MergePatchType, data, k8smetav1.PatchOptions{})
	return err
}

func (br *IstioBridge) patchAuthorizationPolicy(ctx context.Context, namespace string, name string, metadata map[string]interface{}, spec map[string]interface{}) error {
	data, err := buildMergePatch(metadata, spec)
	if err != nil {
		return err
	}

	_, err = br.securityResources.PatchAuthorizationPolicy(ctx, namespace, name, k8stypes.MergePatchType, data, k8smetav1.PatchOptions{})
	return err
}

func (br *IstioBridge) patchRequestAuthentication(ctx context.Context, namespace string, name string, metadata map[string]interface{}, spec map[string]interface{}) error {
	data, err := buildMergePatch(metadata, spec)
	if err != nil {
		return err
	}

	_, err = br.securityResources.PatchRequestAuthentication(ctx, namespace, name, k8stypes.MergePatchType, data, k8smetav1.PatchOptions{})
	return err
}
