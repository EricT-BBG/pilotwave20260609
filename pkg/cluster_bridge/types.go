package cluster_bridge

import (
	"errors"
	"fmt"
	"strings"

	//	"github.com/gin-gonic/gin"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"
)

const (
	NamespaceIstioInjectionModeDisabled = "disabled"
	NamespaceIstioInjectionModeEnabled  = "enabled"
	NamespaceIstioInjectionModeRevision = "revision"
)

type NamespaceMetadata struct {
	Name            string            `json:"name"`
	IstioInjection  string            `json:"istioInjection"`
	IstioRevision   string            `json:"istioRevision"`
	SystemNamespace bool              `json:"systemNamespace"`
	Labels          map[string]string `json:"labels,omitempty"`
	ResourceVersion string            `json:"resourceVersion,omitempty"`
}

type IstioCapabilities struct {
	Installed                  bool               `json:"installed"`
	Disabled                   bool               `json:"disabled"`
	MissingCRDs                []string           `json:"missingCRDs"`
	AvailableCRDs              []string           `json:"availableCRDs"`
	DefaultInjectionAvailable  bool               `json:"defaultInjectionAvailable"`
	RevisionInjectionAvailable bool               `json:"revisionInjectionAvailable"`
	Revisions                  []string           `json:"revisions"`
	RevisionTags               []IstioRevisionTag `json:"revisionTags"`
	RevisionDetectionMessage   string             `json:"revisionDetectionMessage,omitempty"`
	Message                    string             `json:"message"`
}

type IstioRevisionTag struct {
	Name     string `json:"name"`
	Revision string `json:"revision,omitempty"`
}

type IstioUnavailableError struct {
	Capabilities IstioCapabilities
}

func (err *IstioUnavailableError) Error() string {
	message := strings.TrimSpace(err.Capabilities.Message)
	if message == "" {
		if err.Capabilities.Disabled {
			message = "Cluster bridge is disabled for local development"
		} else {
			message = "Istio CRDs are not installed or incomplete in this cluster"
		}
	}

	if len(err.Capabilities.MissingCRDs) > 0 {
		return fmt.Sprintf("%s (missing: %s)", message, strings.Join(err.Capabilities.MissingCRDs, ", "))
	}

	return message
}

func NewIstioUnavailableError(capabilities IstioCapabilities) error {
	return &IstioUnavailableError{Capabilities: capabilities}
}

func IsIstioUnavailable(err error) bool {
	var unavailableErr *IstioUnavailableError
	return errors.As(err, &unavailableErr)
}

func IsSystemNamespaceName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "istio-system" {
		return true
	}

	return strings.HasPrefix(name, "kube-") || strings.HasPrefix(name, "openshift-")
}

type Bridge interface {
	Init() error
	GetIstioCapabilities() IstioCapabilities

	// Router
	GetRouters(int, int, string, string) ([]router_manager.RouterResponse, int, error)
	GetRouter(string, string) (*router_manager.RouterResponse, error)
	GetRouterServices(string, string) ([]string, error)
	CreateRouter(string, string, string, []string) error
	UpdateRouter(string, string, string, []string, string) error
	DeleteRouter(string, string) error

	UpdateRouterRule(name string, namespace string, rule router_manager.RouterRuleRequest) error
	GetRouterRule(name string, namespace string) (router_manager.RouterRuleResponse, error)

	GetRouterGatewayMapping(name string, namespace string) (routers router_manager.RouterMappingResponse, err error)
	CreateRouterGatewayMapping(name string, namespace string, gateways []router_manager.RouterMappingGatewayData, resourceVersion string) error
	UpdateRouterGatewayMapping(name string, namespace string, gateways []router_manager.RouterMappingGatewayData, resourceVersion string) error
	DeleteRouterGatewayMapping(name string, namespace string) error

	//Namespace
	GetNamespaces() ([]string, error)
	GetNamespaceMetadata() ([]NamespaceMetadata, error)
	PatchNamespaceIstioInjection(name string, mode string, revision string) (NamespaceMetadata, error)

	//Secrets
	CreateSecrets(string, string, string, string, string) error
	SecretsExist(name string, namespace string) (bool, error)
	DeleteSecrets(string, string) error

	// Gateway
	GetGateways(page int, perPage int, search string, namespace string) ([]gateway_manager.GatewayResponse, int, error)
	GetGateway(name string, namespace string) (*gateway_manager.GatewayResponse, error)
	GetGatewayTLSCertificates(name string, namespace string) ([]gateway_manager.GatewayTLSCertificateResponse, error)
	CreateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error
	UpdateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error
	DeleteGateway(name string, namespace string) error

	GetGatewayRouterMapping(name string, namespace string) (routers gateway_manager.RouterMappingResponse, err error)
	CreateGatewayRouterMapping(name string, namespace string, routers []gateway_manager.GatewayMappinRouterData, resourceVersions map[string]string) error
	UpdateGatewayRouterMapping(name string, namespace string, routers []gateway_manager.GatewayMappinRouterData, resourceVersions map[string]string) error
	DeleteGatewayRouterMapping(name string, namespace string) error

	// Authentication Policy
	GetAuthorizationPolicies(page int, perPage int, search string, namespace string) ([]*security_manager.AuthorizationPolicyResponse, int, error)
	GetAuthorizationPolicy(name string, namespace string) (*security_manager.AuthorizationPolicyResponse, error)
	CreateAuthorizationPolicy(name string, namespace string, data *security_manager.AuthorizationPolicyRequest) error
	UpdateAuthorizationPolicy(name string, namespace string, data *security_manager.AuthorizationPolicyUpdateRequest) error
	DeleteAuthorizationPolicy(name string, namespace string) error

	// Request Authentication
	GetRequestAuthentications(page int, perPage int, search string, namespace string) ([]*security_manager.RequestAuthenticationResponse, int, error)
	GetRequestAuthentication(name string, namespace string) (*security_manager.RequestAuthenticationResponse, error)
	CreateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationRequest) error
	UpdateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationUpdateRequest) error
	DeleteRequestAuthentication(name string, namespace string) error
}
