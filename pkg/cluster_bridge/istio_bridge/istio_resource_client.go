package istio_bridge

import (
	"context"
	"time"

	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurity "istio.io/client-go/pkg/apis/security/v1beta1"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

type GatewayResourceClient interface {
	ListGateways(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istionetworking.GatewayList, error)
	GetGateway(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istionetworking.Gateway, error)
	CreateGateway(ctx context.Context, namespace string, gateway *istionetworking.Gateway, opts k8smetav1.CreateOptions) (*istionetworking.Gateway, error)
	DeleteGateway(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error
	PatchGateway(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.Gateway, error)
}

type RouterResourceClient interface {
	ListVirtualServices(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istionetworking.VirtualServiceList, error)
	GetVirtualService(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istionetworking.VirtualService, error)
	CreateVirtualService(ctx context.Context, namespace string, virtualService *istionetworking.VirtualService, opts k8smetav1.CreateOptions) (*istionetworking.VirtualService, error)
	DeleteVirtualService(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error
	PatchVirtualService(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.VirtualService, error)

	GetDestinationRule(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istionetworking.DestinationRule, error)
	CreateDestinationRule(ctx context.Context, namespace string, destinationRule *istionetworking.DestinationRule, opts k8smetav1.CreateOptions) (*istionetworking.DestinationRule, error)
	PatchDestinationRule(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.DestinationRule, error)
}

type SecurityResourceClient interface {
	ListAuthorizationPolicies(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istiosecurity.AuthorizationPolicyList, error)
	GetAuthorizationPolicy(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istiosecurity.AuthorizationPolicy, error)
	CreateAuthorizationPolicy(ctx context.Context, namespace string, policy *istiosecurity.AuthorizationPolicy, opts k8smetav1.CreateOptions) (*istiosecurity.AuthorizationPolicy, error)
	DeleteAuthorizationPolicy(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error
	PatchAuthorizationPolicy(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istiosecurity.AuthorizationPolicy, error)

	ListRequestAuthentications(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istiosecurity.RequestAuthenticationList, error)
	GetRequestAuthentication(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istiosecurity.RequestAuthentication, error)
	CreateRequestAuthentication(ctx context.Context, namespace string, requestAuthentication *istiosecurity.RequestAuthentication, opts k8smetav1.CreateOptions) (*istiosecurity.RequestAuthentication, error)
	DeleteRequestAuthentication(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error
	PatchRequestAuthentication(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istiosecurity.RequestAuthentication, error)
}

type IstioResourceClient interface {
	GatewayResourceClient
	RouterResourceClient
	SecurityResourceClient
}

type clientsetIstioResourceClient struct {
	clientset istioclient.Interface
}

func newIstioResourceClient(clientset istioclient.Interface) IstioResourceClient {
	return &clientsetIstioResourceClient{clientset: clientset}
}

func (c *clientsetIstioResourceClient) ListGateways(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istionetworking.GatewayList, error) {
	return c.clientset.NetworkingV1alpha3().Gateways(namespace).List(ctx, opts)
}

func (c *clientsetIstioResourceClient) GetGateway(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istionetworking.Gateway, error) {
	return c.clientset.NetworkingV1alpha3().Gateways(namespace).Get(ctx, name, opts)
}

func (c *clientsetIstioResourceClient) CreateGateway(ctx context.Context, namespace string, gateway *istionetworking.Gateway, opts k8smetav1.CreateOptions) (*istionetworking.Gateway, error) {
	start := time.Now()
	result, err := c.clientset.NetworkingV1alpha3().Gateways(namespace).Create(ctx, gateway, opts)
	recordIstioWrite(istioResourceGateway, kubernetesWriteVerbCreate, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) DeleteGateway(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error {
	start := time.Now()
	err := c.clientset.NetworkingV1alpha3().Gateways(namespace).Delete(ctx, name, opts)
	recordIstioWrite(istioResourceGateway, kubernetesWriteVerbDelete, start, err)
	return err
}

func (c *clientsetIstioResourceClient) PatchGateway(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.Gateway, error) {
	start := time.Now()
	result, err := c.clientset.NetworkingV1alpha3().Gateways(namespace).Patch(ctx, name, patchType, data, opts)
	recordIstioWrite(istioResourceGateway, kubernetesWriteVerbPatch, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) ListVirtualServices(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istionetworking.VirtualServiceList, error) {
	return c.clientset.NetworkingV1alpha3().VirtualServices(namespace).List(ctx, opts)
}

func (c *clientsetIstioResourceClient) GetVirtualService(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istionetworking.VirtualService, error) {
	return c.clientset.NetworkingV1alpha3().VirtualServices(namespace).Get(ctx, name, opts)
}

func (c *clientsetIstioResourceClient) CreateVirtualService(ctx context.Context, namespace string, virtualService *istionetworking.VirtualService, opts k8smetav1.CreateOptions) (*istionetworking.VirtualService, error) {
	start := time.Now()
	result, err := c.clientset.NetworkingV1alpha3().VirtualServices(namespace).Create(ctx, virtualService, opts)
	recordIstioWrite(istioResourceVirtualService, kubernetesWriteVerbCreate, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) DeleteVirtualService(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error {
	start := time.Now()
	err := c.clientset.NetworkingV1alpha3().VirtualServices(namespace).Delete(ctx, name, opts)
	recordIstioWrite(istioResourceVirtualService, kubernetesWriteVerbDelete, start, err)
	return err
}

func (c *clientsetIstioResourceClient) PatchVirtualService(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.VirtualService, error) {
	start := time.Now()
	result, err := c.clientset.NetworkingV1alpha3().VirtualServices(namespace).Patch(ctx, name, patchType, data, opts)
	recordIstioWrite(istioResourceVirtualService, kubernetesWriteVerbPatch, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) GetDestinationRule(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istionetworking.DestinationRule, error) {
	return c.clientset.NetworkingV1alpha3().DestinationRules(namespace).Get(ctx, name, opts)
}

func (c *clientsetIstioResourceClient) CreateDestinationRule(ctx context.Context, namespace string, destinationRule *istionetworking.DestinationRule, opts k8smetav1.CreateOptions) (*istionetworking.DestinationRule, error) {
	start := time.Now()
	result, err := c.clientset.NetworkingV1alpha3().DestinationRules(namespace).Create(ctx, destinationRule, opts)
	recordIstioWrite(istioResourceDestinationRule, kubernetesWriteVerbCreate, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) PatchDestinationRule(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istionetworking.DestinationRule, error) {
	start := time.Now()
	result, err := c.clientset.NetworkingV1alpha3().DestinationRules(namespace).Patch(ctx, name, patchType, data, opts)
	recordIstioWrite(istioResourceDestinationRule, kubernetesWriteVerbPatch, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) ListAuthorizationPolicies(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istiosecurity.AuthorizationPolicyList, error) {
	return c.clientset.SecurityV1beta1().AuthorizationPolicies(namespace).List(ctx, opts)
}

func (c *clientsetIstioResourceClient) GetAuthorizationPolicy(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istiosecurity.AuthorizationPolicy, error) {
	return c.clientset.SecurityV1beta1().AuthorizationPolicies(namespace).Get(ctx, name, opts)
}

func (c *clientsetIstioResourceClient) CreateAuthorizationPolicy(ctx context.Context, namespace string, policy *istiosecurity.AuthorizationPolicy, opts k8smetav1.CreateOptions) (*istiosecurity.AuthorizationPolicy, error) {
	start := time.Now()
	result, err := c.clientset.SecurityV1beta1().AuthorizationPolicies(namespace).Create(ctx, policy, opts)
	recordIstioWrite(istioResourceAuthorizationPolicy, kubernetesWriteVerbCreate, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) DeleteAuthorizationPolicy(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error {
	start := time.Now()
	err := c.clientset.SecurityV1beta1().AuthorizationPolicies(namespace).Delete(ctx, name, opts)
	recordIstioWrite(istioResourceAuthorizationPolicy, kubernetesWriteVerbDelete, start, err)
	return err
}

func (c *clientsetIstioResourceClient) PatchAuthorizationPolicy(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istiosecurity.AuthorizationPolicy, error) {
	start := time.Now()
	result, err := c.clientset.SecurityV1beta1().AuthorizationPolicies(namespace).Patch(ctx, name, patchType, data, opts)
	recordIstioWrite(istioResourceAuthorizationPolicy, kubernetesWriteVerbPatch, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) ListRequestAuthentications(ctx context.Context, namespace string, opts k8smetav1.ListOptions) (*istiosecurity.RequestAuthenticationList, error) {
	return c.clientset.SecurityV1beta1().RequestAuthentications(namespace).List(ctx, opts)
}

func (c *clientsetIstioResourceClient) GetRequestAuthentication(ctx context.Context, namespace string, name string, opts k8smetav1.GetOptions) (*istiosecurity.RequestAuthentication, error) {
	return c.clientset.SecurityV1beta1().RequestAuthentications(namespace).Get(ctx, name, opts)
}

func (c *clientsetIstioResourceClient) CreateRequestAuthentication(ctx context.Context, namespace string, requestAuthentication *istiosecurity.RequestAuthentication, opts k8smetav1.CreateOptions) (*istiosecurity.RequestAuthentication, error) {
	start := time.Now()
	result, err := c.clientset.SecurityV1beta1().RequestAuthentications(namespace).Create(ctx, requestAuthentication, opts)
	recordIstioWrite(istioResourceRequestAuthentication, kubernetesWriteVerbCreate, start, err)
	return result, err
}

func (c *clientsetIstioResourceClient) DeleteRequestAuthentication(ctx context.Context, namespace string, name string, opts k8smetav1.DeleteOptions) error {
	start := time.Now()
	err := c.clientset.SecurityV1beta1().RequestAuthentications(namespace).Delete(ctx, name, opts)
	recordIstioWrite(istioResourceRequestAuthentication, kubernetesWriteVerbDelete, start, err)
	return err
}

func (c *clientsetIstioResourceClient) PatchRequestAuthentication(ctx context.Context, namespace string, name string, patchType k8stypes.PatchType, data []byte, opts k8smetav1.PatchOptions) (*istiosecurity.RequestAuthentication, error) {
	start := time.Now()
	result, err := c.clientset.SecurityV1beta1().RequestAuthentications(namespace).Patch(ctx, name, patchType, data, opts)
	recordIstioWrite(istioResourceRequestAuthentication, kubernetesWriteVerbPatch, start, err)
	return result, err
}
