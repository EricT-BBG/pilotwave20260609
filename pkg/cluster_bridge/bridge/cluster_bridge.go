package cluster_bridge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	app "git.brobridge.com/pilotwave/pilotwave/pkg/app"
	cluster_bridge_types "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	istio_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge/istio_bridge"
	k8s_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge/k8s_bridge"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/metrics"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"

	//log "github.com/sirupsen/logrus"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	k8sclient "k8s.io/client-go/kubernetes"

	k8srestapi "k8s.io/client-go/rest"
	k8sclientcmd "k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
)

type ClusterBridge struct {
	app               app.App
	istioClientset    *istioclient.Clientset
	k8sClientset      *k8sclient.Clientset
	istioBridge       *istio_bridge.IstioBridge
	k8sBridge         *k8s_bridge.K8sBridge
	disabled          bool
	istioCapabilities cluster_bridge_types.IstioCapabilities
	metricsCancel     context.CancelFunc
}

func NewClusterBridge(a app.App) *ClusterBridge {
	return &ClusterBridge{
		app: a,
	}
}

func (bridge *ClusterBridge) Init() error {
	if viper.GetBool("cluster.disabled") {
		bridge.disabled = true
		bridge.istioCapabilities = disabledIstioCapabilities()
		metrics.SetIstioClusterSnapshot(metrics.IstioClusterSnapshot{})
		log.Warn("Cluster bridge disabled; Kubernetes and Istio APIs will return empty read results or clear errors")
		return nil
	}

	// Setup configuration for when runs inside/outside
	// the cluster and the API client for making requests.
	var cfg *k8srestapi.Config
	var err error
	//if kubeconfig := os.Getenv("KUBERNETES_CONFIG_FILE"); kubeconfig != "" {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		cfg, err = k8sclientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		cfg, err = k8srestapi.InClusterConfig()
	}
	if err != nil {
		return err
	}

	//configuration and prepares the istio rest clients that will be used to interact with the cluster objects.
	istioClientset, err := istioclient.NewForConfig(cfg)
	if err != nil {
		return err
	}

	bridge.istioClientset = istioClientset

	//configuration and prepares the k8s rest clients that will be used to interact with the cluster objects.
	k8sClientset, err := k8sclient.NewForConfig(cfg)
	if err != nil {
		return err
	}

	bridge.k8sClientset = k8sClientset
	bridge.istioCapabilities = detectIstioCapabilitiesWithClient(bridge.k8sClientset)

	// Init istio bridge
	bridge.istioBridge = istio_bridge.NewIstioBridge(bridge.app, bridge.istioClientset)

	// Init k8s bridge
	bridge.k8sBridge = k8s_bridge.NewK8sBridge(bridge.app, bridge.k8sClientset)

	bridge.startIstioMetricsRefresh()

	return nil
}

func (bridge *ClusterBridge) GetIstioCapabilities() cluster_bridge_types.IstioCapabilities {
	return bridge.istioCapabilities
}

func (bridge *ClusterBridge) startIstioMetricsRefresh() {
	if bridge.metricsCancel != nil {
		bridge.metricsCancel()
	}

	interval := viper.GetDuration("cluster.metrics_refresh_interval")
	if interval <= 0 {
		interval = defaultIstioMetricsRefreshInterval
	}

	ctx, cancel := context.WithCancel(context.Background())
	bridge.metricsCancel = cancel
	refresher := newIstioMetricsRefresher(bridge.istioClientset, bridge.k8sClientset)

	go func() {
		if err := refresher.refresh(ctx); err != nil {
			log.WithError(err).Warn("Failed to refresh Istio resource metrics")
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := refresher.refresh(ctx); err != nil {
					log.WithError(err).Warn("Failed to refresh Istio resource metrics")
				}
			}
		}
	}()
}

func (bridge *ClusterBridge) clusterUnavailableError() error {
	return cluster_bridge_types.NewIstioUnavailableError(disabledIstioCapabilities())
}

func (bridge *ClusterBridge) istioUnavailableError() error {
	capabilities := bridge.istioCapabilities
	if bridge.disabled || capabilities.Disabled {
		capabilities = disabledIstioCapabilities()
	}
	if strings.TrimSpace(capabilities.Message) == "" {
		capabilities.Message = "Istio CRDs are not installed or incomplete in this cluster"
	}

	return cluster_bridge_types.NewIstioUnavailableError(capabilities)
}

func (bridge *ClusterBridge) ensureIstioAvailable() error {
	if bridge.disabled || !bridge.istioCapabilities.Installed {
		return bridge.istioUnavailableError()
	}

	return nil
}

// ---------------------------------------------------
// Router
// ---------------------------------------------------

func (bridge *ClusterBridge) GetRouters(page int, perPage int, search string, namespace string) ([]router_manager.RouterResponse, int, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return []router_manager.RouterResponse{}, 0, err
	}

	// List Istio Router (VirtualService)
	return bridge.istioBridge.GetRouters(page, perPage, search, namespace)
}

func (bridge *ClusterBridge) GetRouter(name string, namespace string) (*router_manager.RouterResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return nil, err
	}

	// List Istio Router (VirtualService)
	return bridge.istioBridge.GetRouter(name, namespace)
}

func (bridge *ClusterBridge) GetRouterServices(name string, namespace string) ([]string, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return []string{}, err
	}

	// List Istio Router (VirtualService)
	return bridge.istioBridge.GetRouterServices(name, namespace)
}

func (bridge *ClusterBridge) CreateRouter(name string, namespace string, protocol string, hosts []string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Create Istio Router (VirtualService)
	return bridge.istioBridge.CreateRouter(name, namespace, protocol, hosts)
}

func (bridge *ClusterBridge) UpdateRouter(name string, namespace string, protocol string, hosts []string, resourceVersion string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Update Istio Router (VirtualService)
	return bridge.istioBridge.UpdateRouter(name, namespace, protocol, hosts, resourceVersion)
}

func (bridge *ClusterBridge) DeleteRouter(name string, namespace string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Delete Istio Router (VirtualService)
	return bridge.istioBridge.DeleteRouter(name, namespace)
}

func (bridge *ClusterBridge) UpdateRouterRule(name string, namespace string, rule router_manager.RouterRuleRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.UpdateRouterRule(name, namespace, rule)
}

func (bridge *ClusterBridge) GetRouterRule(name string, namespace string) (router_manager.RouterRuleResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return router_manager.RouterRuleResponse{}, err
	}

	return bridge.istioBridge.GetRouterRule(name, namespace)
}

// Namespaces
func (bridge *ClusterBridge) GetNamespaces() ([]string, error) {
	if bridge.disabled {
		return []string{"default"}, nil
	}

	// List Namespace from kubernetes
	return bridge.k8sBridge.GetNamespaces()
}

func (bridge *ClusterBridge) GetNamespaceMetadata() ([]cluster_bridge_types.NamespaceMetadata, error) {
	if bridge.disabled {
		return []cluster_bridge_types.NamespaceMetadata{
			{
				Name:           "default",
				IstioInjection: "",
				IstioRevision:  "",
				Labels:         map[string]string{},
			},
		}, nil
	}

	// List Namespace metadata from kubernetes
	return bridge.k8sBridge.GetNamespaceMetadata()
}

func (bridge *ClusterBridge) PatchNamespaceIstioInjection(name string, mode string, revision string) (cluster_bridge_types.NamespaceMetadata, error) {
	name = strings.TrimSpace(name)
	mode = strings.TrimSpace(mode)
	revision = strings.TrimSpace(revision)
	if err := validateNamespaceIstioInjectionPatch(name, mode, revision); err != nil {
		return cluster_bridge_types.NamespaceMetadata{}, err
	}

	if err := bridge.ensureIstioAvailable(); err != nil {
		return cluster_bridge_types.NamespaceMetadata{}, err
	}

	return bridge.k8sBridge.PatchNamespaceIstioInjection(name, mode, revision)
}

func validateNamespaceIstioInjectionPatch(name string, mode string, revision string) error {
	if name == "" {
		return errors.New("namespace name is required")
	}

	if errs := k8svalidation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid namespace name: %s", strings.Join(errs, ", "))
	}

	switch mode {
	case cluster_bridge_types.NamespaceIstioInjectionModeDisabled, cluster_bridge_types.NamespaceIstioInjectionModeEnabled:
		if revision != "" {
			return errors.New("revision must be empty unless mode is revision")
		}
	case cluster_bridge_types.NamespaceIstioInjectionModeRevision:
		if revision == "" {
			return errors.New("revision is required when mode is revision")
		}
		if errs := k8svalidation.IsValidLabelValue(revision); len(errs) > 0 {
			return fmt.Errorf("invalid revision: %s", strings.Join(errs, ", "))
		}
	default:
		return errors.New("mode must be disabled, enabled, or revision")
	}

	return nil
}

// Secrets
func (bridge *ClusterBridge) CreateSecrets(name string, namespace string, certificate string, privateKey string, caCertificate string) error {
	if bridge.disabled {
		return bridge.clusterUnavailableError()
	}

	// Create Secrets
	return bridge.k8sBridge.CreateSecrets(name, namespace, certificate, privateKey, caCertificate)
}

func (bridge *ClusterBridge) DeleteSecrets(name string, namespace string) error {
	if bridge.disabled {
		return bridge.clusterUnavailableError()
	}

	// Delete Secrets
	return bridge.k8sBridge.DeleteSecrets(name, namespace)
}

func (bridge *ClusterBridge) SecretsExist(name string, namespace string) (bool, error) {
	if bridge.disabled {
		return false, nil
	}

	// Secrets Exist
	return bridge.k8sBridge.SecretsExist(name, namespace)
}

func (bridge *ClusterBridge) GetRouterGatewayMapping(name string, namespace string) (routers router_manager.RouterMappingResponse, err error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return router_manager.RouterMappingResponse{}, err
	}

	// Get Gateways for router
	return bridge.istioBridge.GetRouterGatewayMapping(name, namespace)
}

func (bridge *ClusterBridge) CreateRouterGatewayMapping(name string, namespace string, routers []router_manager.RouterMappingGatewayData, resourceVersion string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Create Gateways for router
	return bridge.istioBridge.CreateRouterGatewayMapping(name, namespace, routers, resourceVersion)
}
func (bridge *ClusterBridge) UpdateRouterGatewayMapping(name string, namespace string, routers []router_manager.RouterMappingGatewayData, resourceVersion string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Update Gateways for router
	return bridge.istioBridge.UpdateRouterGatewayMapping(name, namespace, routers, resourceVersion)
}
func (bridge *ClusterBridge) DeleteRouterGatewayMapping(name string, namespace string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Delete Gateways for router
	return bridge.istioBridge.DeleteRouterGatewayMapping(name, namespace)
}

// ---------------------------------------------------
// Gateway
// ---------------------------------------------------

func (bridge *ClusterBridge) GetGateways(page int, perPage int, search string, namespace string) ([]gateway_manager.GatewayResponse, int, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return []gateway_manager.GatewayResponse{}, 0, err
	}

	// list Istio Gateway via namespace
	return bridge.istioBridge.GetGateways(page, perPage, search, namespace)
}
func (bridge *ClusterBridge) GetGateway(name string, namespace string) (*gateway_manager.GatewayResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return nil, err
	}

	// Retrieve via Istio Gateway via namespace
	return bridge.istioBridge.GetGateway(name, namespace)
}

func (bridge *ClusterBridge) GetGatewayTLSCertificates(name string, namespace string) ([]gateway_manager.GatewayTLSCertificateResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return []gateway_manager.GatewayTLSCertificateResponse{}, err
	}

	return bridge.getGatewayTLSCertificates(context.Background(), name, namespace)
}

func (bridge *ClusterBridge) CreateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Create Istio Gateway via namespace
	return bridge.istioBridge.CreateGateway(name, namespace, requestdata)
}

func (bridge *ClusterBridge) UpdateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Update Istio Gateway via namespace
	return bridge.istioBridge.UpdateGateway(name, namespace, requestdata)
}

func (bridge *ClusterBridge) DeleteGateway(name string, namespace string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Delete Istio Gateway via namespace
	return bridge.istioBridge.DeleteGateway(name, namespace)
}

func (bridge *ClusterBridge) GetGatewayRouterMapping(name string, namespace string) (gateway_manager.RouterMappingResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return gateway_manager.RouterMappingResponse{}, err
	}

	// Get Router mapping for via gateway
	return bridge.istioBridge.GetGatewayRouterMapping(name, namespace)
}

func (bridge *ClusterBridge) CreateGatewayRouterMapping(name string, namespace string, routers []gateway_manager.GatewayMappinRouterData, resourceVersions map[string]string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Create Router mapping for via gateway
	return bridge.istioBridge.CreateGatewayRouterMapping(name, namespace, routers, resourceVersions)
}

func (bridge *ClusterBridge) UpdateGatewayRouterMapping(name string, namespace string, routers []gateway_manager.GatewayMappinRouterData, resourceVersions map[string]string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Update Router mapping for via gateway
	return bridge.istioBridge.UpdateGatewayRouterMapping(name, namespace, routers, resourceVersions)
}

func (bridge *ClusterBridge) DeleteGatewayRouterMapping(name string, namespace string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	// Delete Router mapping for via gateway
	return bridge.istioBridge.DeleteGatewayRouterMapping(name, namespace)
}

// ---------------------------------------------------
// AuthorizationPolicy
// ---------------------------------------------------

func (bridge *ClusterBridge) GetAuthorizationPolicies(page int, perPage int, search string, namespace string) ([]*security_manager.AuthorizationPolicyResponse, int, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return []*security_manager.AuthorizationPolicyResponse{}, 0, err
	}

	return bridge.istioBridge.GetAuthorizationPolicies(page, perPage, search, namespace)
}

func (bridge *ClusterBridge) CreateAuthorizationPolicy(name string, namespace string, data *security_manager.AuthorizationPolicyRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.CreateAuthorizationPolicy(name, namespace, data)
}

func (bridge *ClusterBridge) GetAuthorizationPolicy(name string, namespace string) (*security_manager.AuthorizationPolicyResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return nil, err
	}

	return bridge.istioBridge.GetAuthorizationPolicy(name, namespace)
}

func (bridge *ClusterBridge) UpdateAuthorizationPolicy(name string, namespace string, data *security_manager.AuthorizationPolicyUpdateRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.UpdateAuthorizationPolicy(name, namespace, data)
}

func (bridge *ClusterBridge) DeleteAuthorizationPolicy(name string, namespace string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.DeleteAuthorizationPolicy(name, namespace)
}

// ---------------------------------------------------
// RequestAuthentications
// ---------------------------------------------------

func (bridge *ClusterBridge) GetRequestAuthentications(page int, perPage int, search string, namespace string) ([]*security_manager.RequestAuthenticationResponse, int, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return []*security_manager.RequestAuthenticationResponse{}, 0, err
	}

	return bridge.istioBridge.GetRequestAuthentications(page, perPage, search, namespace)
}

func (bridge *ClusterBridge) CreateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.CreateRequestAuthentication(name, namespace, data)
}

func (bridge *ClusterBridge) GetRequestAuthentication(name string, namespace string) (*security_manager.RequestAuthenticationResponse, error) {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return nil, err
	}

	return bridge.istioBridge.GetRequestAuthentication(name, namespace)
}

func (bridge *ClusterBridge) UpdateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationUpdateRequest) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.UpdateRequestAuthentication(name, namespace, data)
}

func (bridge *ClusterBridge) DeleteRequestAuthentication(name string, namespace string) error {
	if err := bridge.ensureIstioAvailable(); err != nil {
		return err
	}

	return bridge.istioBridge.DeleteRequestAuthentication(name, namespace)
}
