package cluster_bridge

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"math"
	"strings"
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/metrics"
	"github.com/spf13/viper"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
)

const (
	defaultIstioMetricsRefreshInterval = time.Minute

	istioInjectionLabel = "istio-injection"
	istioRevisionLabel  = "istio.io/rev"

	namespaceInjectionDisabled = "disabled"
	namespaceInjectionEnabled  = "enabled"
	namespaceInjectionRevision = "revision"

	istioSecretCertKey          = "cert"
	defaultIstioSecretNamespace = "istio-system"
)

type istioMetricsRefresher struct {
	istioClientset istioclient.Interface
	k8sClientset   k8sclient.Interface
	now            func() time.Time
}

func newIstioMetricsRefresher(istioClientset istioclient.Interface, k8sClientset k8sclient.Interface) *istioMetricsRefresher {
	return &istioMetricsRefresher{
		istioClientset: istioClientset,
		k8sClientset:   k8sClientset,
		now:            time.Now,
	}
}

func (r *istioMetricsRefresher) refresh(ctx context.Context) error {
	snapshot, err := r.snapshot(ctx)
	if err != nil {
		return err
	}
	metrics.SetIstioClusterSnapshot(snapshot)
	return nil
}

func (r *istioMetricsRefresher) snapshot(ctx context.Context) (metrics.IstioClusterSnapshot, error) {
	var snapshot metrics.IstioClusterSnapshot

	if err := r.collectResources(ctx, &snapshot); err != nil {
		return snapshot, err
	}
	if err := r.collectNamespaceInjection(ctx, &snapshot); err != nil {
		return snapshot, err
	}
	if err := r.collectGatewayTLS(ctx, &snapshot); err != nil {
		return snapshot, err
	}

	return snapshot, nil
}

func (r *istioMetricsRefresher) collectResources(ctx context.Context, snapshot *metrics.IstioClusterSnapshot) error {
	listOpts := k8smetav1.ListOptions{}

	gateways, err := r.istioClientset.NetworkingV1alpha3().Gateways(k8smetav1.NamespaceAll).List(ctx, listOpts)
	if err != nil {
		return err
	}
	for _, resource := range gateways.Items {
		appendIstioResource(snapshot, "Gateway", resource.Namespace, resource.Name, resource.Generation)
	}

	virtualServices, err := r.istioClientset.NetworkingV1alpha3().VirtualServices(k8smetav1.NamespaceAll).List(ctx, listOpts)
	if err != nil {
		return err
	}
	for _, resource := range virtualServices.Items {
		appendIstioResource(snapshot, "VirtualService", resource.Namespace, resource.Name, resource.Generation)
	}

	destinationRules, err := r.istioClientset.NetworkingV1alpha3().DestinationRules(k8smetav1.NamespaceAll).List(ctx, listOpts)
	if err != nil {
		return err
	}
	for _, resource := range destinationRules.Items {
		appendIstioResource(snapshot, "DestinationRule", resource.Namespace, resource.Name, resource.Generation)
	}

	authorizationPolicies, err := r.istioClientset.SecurityV1beta1().AuthorizationPolicies(k8smetav1.NamespaceAll).List(ctx, listOpts)
	if err != nil {
		return err
	}
	for _, resource := range authorizationPolicies.Items {
		appendIstioResource(snapshot, "AuthorizationPolicy", resource.Namespace, resource.Name, resource.Generation)
	}

	requestAuthentications, err := r.istioClientset.SecurityV1beta1().RequestAuthentications(k8smetav1.NamespaceAll).List(ctx, listOpts)
	if err != nil {
		return err
	}
	for _, resource := range requestAuthentications.Items {
		appendIstioResource(snapshot, "RequestAuthentication", resource.Namespace, resource.Name, resource.Generation)
	}

	return nil
}

func appendIstioResource(snapshot *metrics.IstioClusterSnapshot, resource string, namespace string, name string, generation int64) {
	snapshot.Resources = append(snapshot.Resources, metrics.IstioResourceMetric{
		Resource:   resource,
		Namespace:  namespace,
		Name:       name,
		Generation: generation,
	})
}

func (r *istioMetricsRefresher) collectNamespaceInjection(ctx context.Context, snapshot *metrics.IstioClusterSnapshot) error {
	namespaces, err := r.k8sClientset.CoreV1().Namespaces().List(ctx, k8smetav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, namespace := range namespaces.Items {
		mode, revision := namespaceInjectionState(namespace.Labels)
		snapshot.NamespaceInjections = append(snapshot.NamespaceInjections, metrics.IstioNamespaceInjectionMetric{
			Namespace: namespace.Name,
			Mode:      mode,
			Revision:  revision,
		})
	}

	return nil
}

func namespaceInjectionState(labels map[string]string) (string, string) {
	if labels == nil {
		return namespaceInjectionDisabled, ""
	}
	if revision := strings.TrimSpace(labels[istioRevisionLabel]); revision != "" {
		return namespaceInjectionRevision, revision
	}
	if labels[istioInjectionLabel] == namespaceInjectionEnabled {
		return namespaceInjectionEnabled, ""
	}
	return namespaceInjectionDisabled, ""
}

func (r *istioMetricsRefresher) collectGatewayTLS(ctx context.Context, snapshot *metrics.IstioClusterSnapshot) error {
	gateways, err := r.istioClientset.NetworkingV1alpha3().Gateways(k8smetav1.NamespaceAll).List(ctx, k8smetav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, gateway := range gateways.Items {
		for _, server := range gateway.Spec.GetServers() {
			tlsSettings := server.GetTls()
			credentialName := strings.TrimSpace(tlsSettings.GetCredentialName())
			if credentialName == "" {
				continue
			}

			secretNamespace, secretName := resolveCredentialSecret(defaultGatewayTLSSecretNamespace(), credentialName)
			secret, err := r.k8sClientset.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, k8smetav1.GetOptions{})
			if err != nil {
				reason := "read_error"
				if k8serrors.IsNotFound(err) {
					reason = "not_found"
				}
				snapshot.GatewaySecretMissing = append(snapshot.GatewaySecretMissing, metrics.IstioGatewayTLSSecretIssueMetric{
					Namespace: secretNamespace,
					Gateway:   gateway.Name,
					Secret:    secretName,
					Reason:    reason,
				})
				continue
			}

			notAfter, err := certificateNotAfter(secret)
			if err != nil {
				snapshot.GatewaySecretInvalid = append(snapshot.GatewaySecretInvalid, metrics.IstioGatewayTLSSecretIssueMetric{
					Namespace: secretNamespace,
					Gateway:   gateway.Name,
					Secret:    secretName,
					Reason:    err.Error(),
				})
				continue
			}

			daysUntilExpiry := notAfter.Sub(r.now()).Hours() / 24
			snapshot.TLSCertificates = append(snapshot.TLSCertificates, metrics.IstioTLSCertificateMetric{
				Namespace:       secretNamespace,
				Gateway:         gateway.Name,
				Secret:          secretName,
				NotAfterUnix:    float64(notAfter.Unix()),
				DaysUntilExpiry: math.Floor(daysUntilExpiry),
				Expired:         !notAfter.After(r.now()),
			})
		}
	}

	return nil
}

func defaultGatewayTLSSecretNamespace() string {
	namespace := strings.TrimSpace(viper.GetString("gateway.tls_secret_namespace"))
	if namespace == "" {
		return defaultIstioSecretNamespace
	}
	return namespace
}

func resolveCredentialSecret(defaultNamespace string, credentialName string) (string, string) {
	parts := strings.SplitN(credentialName, "/", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return defaultNamespace, credentialName
}

func certificateNotAfter(secret *k8scorev1.Secret) (time.Time, error) {
	certData := secret.Data[k8scorev1.TLSCertKey]
	if len(certData) == 0 {
		certData = secret.Data[istioSecretCertKey]
	}
	if len(certData) == 0 {
		return time.Time{}, metricReasonError("missing_certificate")
	}

	var earliest time.Time
	for {
		block, remaining := pem.Decode(certData)
		if block == nil {
			break
		}
		certData = remaining
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return time.Time{}, metricReasonError("parse_error")
		}
		if earliest.IsZero() || cert.NotAfter.Before(earliest) {
			earliest = cert.NotAfter
		}
	}

	if earliest.IsZero() {
		return time.Time{}, metricReasonError("parse_error")
	}

	return earliest, nil
}

type metricReasonError string

func (e metricReasonError) Error() string {
	return string(e)
}
