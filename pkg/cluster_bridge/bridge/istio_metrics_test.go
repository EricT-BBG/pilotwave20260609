package cluster_bridge

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/metrics"
	"github.com/spf13/viper"
	istionetworkingapi "istio.io/api/networking/v1alpha3"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiosecurity "istio.io/client-go/pkg/apis/security/v1beta1"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestIstioMetricsSnapshotCollectsResourcesNamespaceInjectionAndTLS(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()
	now := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	notAfter := now.Add(72 * time.Hour)

	istioClient := istiofake.NewSimpleClientset(
		&istionetworking.VirtualService{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "app-route", Namespace: "app", Generation: 11},
		},
		&istionetworking.DestinationRule{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "app-dr", Namespace: "app", Generation: 13},
		},
		&istiosecurity.AuthorizationPolicy{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "app-authz", Namespace: "app", Generation: 17},
		},
		&istiosecurity.RequestAuthentication{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "app-jwt", Namespace: "app", Generation: 19},
		},
	)
	_, err := istioClient.NetworkingV1alpha3().Gateways("edge").Create(context.Background(), &istionetworking.Gateway{
		ObjectMeta: k8smetav1.ObjectMeta{Name: "edge-gateway", Namespace: "edge", Generation: 7},
		Spec: istionetworkingapi.Gateway{
			Servers: []*istionetworkingapi.Server{
				{Tls: &istionetworkingapi.ServerTLSSettings{CredentialName: "wildcard-cert"}},
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset(
		&k8scorev1.Namespace{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name: "app",
				Labels: map[string]string{
					istioRevisionLabel: "canary",
				},
			},
		},
		&k8scorev1.Namespace{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name: "edge",
				Labels: map[string]string{
					istioInjectionLabel: namespaceInjectionEnabled,
				},
			},
		},
		&k8scorev1.Secret{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "wildcard-cert", Namespace: "istio-system"},
			Type:       k8scorev1.SecretTypeTLS,
			Data: map[string][]byte{
				k8scorev1.TLSCertKey: testCertificatePEM(t, now.Add(-time.Hour), notAfter),
			},
		},
	)

	refresher := newIstioMetricsRefresher(istioClient, k8sClient)
	refresher.now = func() time.Time { return now }

	snapshot, err := refresher.snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assertResourceMetric(t, snapshot.Resources, "Gateway", "edge", "edge-gateway", 7)
	assertResourceMetric(t, snapshot.Resources, "VirtualService", "app", "app-route", 11)
	assertResourceMetric(t, snapshot.Resources, "DestinationRule", "app", "app-dr", 13)
	assertResourceMetric(t, snapshot.Resources, "AuthorizationPolicy", "app", "app-authz", 17)
	assertResourceMetric(t, snapshot.Resources, "RequestAuthentication", "app", "app-jwt", 19)
	assertNamespaceMetric(t, snapshot.NamespaceInjections, "app", namespaceInjectionRevision, "canary")
	assertNamespaceMetric(t, snapshot.NamespaceInjections, "edge", namespaceInjectionEnabled, "")

	if len(snapshot.TLSCertificates) != 1 {
		t.Fatalf("expected one TLS certificate metric, got %d", len(snapshot.TLSCertificates))
	}
	cert := snapshot.TLSCertificates[0]
	if cert.Namespace != "istio-system" || cert.Gateway != "edge-gateway" || cert.Secret != "wildcard-cert" {
		t.Fatalf("unexpected certificate labels: %#v", cert)
	}
	if cert.NotAfterUnix != float64(notAfter.Unix()) {
		t.Fatalf("expected not_after %v, got %v", notAfter.Unix(), cert.NotAfterUnix)
	}
	if cert.DaysUntilExpiry != 3 {
		t.Fatalf("expected 3 days until expiry, got %v", cert.DaysUntilExpiry)
	}
	if cert.Expired {
		t.Fatal("expected certificate to be unexpired")
	}
}

func TestIstioMetricsSnapshotUsesExplicitCredentialNamespace(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()
	now := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	notAfter := now.Add(72 * time.Hour)

	istioClient := istiofake.NewSimpleClientset()
	_, err := istioClient.NetworkingV1alpha3().Gateways("edge").Create(context.Background(), &istionetworking.Gateway{
		ObjectMeta: k8smetav1.ObjectMeta{Name: "edge-gateway", Namespace: "edge", Generation: 1},
		Spec: istionetworkingapi.Gateway{
			Servers: []*istionetworkingapi.Server{
				{Tls: &istionetworkingapi.ServerTLSSettings{CredentialName: "custom-ns/wildcard-cert"}},
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset(
		&k8scorev1.Secret{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "wildcard-cert", Namespace: "custom-ns"},
			Type:       k8scorev1.SecretTypeTLS,
			Data: map[string][]byte{
				k8scorev1.TLSCertKey: testCertificatePEM(t, now.Add(-time.Hour), notAfter),
			},
		},
	)

	refresher := newIstioMetricsRefresher(istioClient, k8sClient)
	refresher.now = func() time.Time { return now }

	snapshot, err := refresher.snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(snapshot.TLSCertificates) != 1 {
		t.Fatalf("expected one TLS certificate metric, got %d", len(snapshot.TLSCertificates))
	}
	cert := snapshot.TLSCertificates[0]
	if cert.Namespace != "custom-ns" || cert.Gateway != "edge-gateway" || cert.Secret != "wildcard-cert" {
		t.Fatalf("unexpected certificate labels: %#v", cert)
	}
}

func TestIstioMetricsSnapshotReportsMissingAndInvalidGatewayTLSSecrets(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()
	now := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	istioClient := istiofake.NewSimpleClientset()
	_, err := istioClient.NetworkingV1alpha3().Gateways("edge").Create(context.Background(), &istionetworking.Gateway{
		ObjectMeta: k8smetav1.ObjectMeta{Name: "edge-gateway", Namespace: "edge", Generation: 1},
		Spec: istionetworkingapi.Gateway{
			Servers: []*istionetworkingapi.Server{
				{Tls: &istionetworkingapi.ServerTLSSettings{CredentialName: "missing-cert"}},
				{Tls: &istionetworkingapi.ServerTLSSettings{CredentialName: "invalid-cert"}},
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset(
		&k8scorev1.Namespace{ObjectMeta: k8smetav1.ObjectMeta{Name: "edge"}},
		&k8scorev1.Secret{
			ObjectMeta: k8smetav1.ObjectMeta{Name: "invalid-cert", Namespace: "istio-system"},
			Data: map[string][]byte{
				k8scorev1.TLSCertKey: []byte("not a certificate"),
			},
		},
	)

	refresher := newIstioMetricsRefresher(istioClient, k8sClient)
	refresher.now = func() time.Time { return now }

	snapshot, err := refresher.snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assertSecretIssueMetric(t, snapshot.GatewaySecretMissing, "istio-system", "edge-gateway", "missing-cert", "not_found")
	assertSecretIssueMetric(t, snapshot.GatewaySecretInvalid, "istio-system", "edge-gateway", "invalid-cert", "parse_error")
}

func assertResourceMetric(t *testing.T, resources []metricsResource, resource string, namespace string, name string, generation int64) {
	t.Helper()
	for _, item := range resources {
		if item.Resource == resource && item.Namespace == namespace && item.Name == name && item.Generation == generation {
			return
		}
	}
	t.Fatalf("missing resource metric %s/%s/%s generation %d in %#v", resource, namespace, name, generation, resources)
}

func assertNamespaceMetric(t *testing.T, namespaces []metricsNamespaceInjection, namespace string, mode string, revision string) {
	t.Helper()
	for _, item := range namespaces {
		if item.Namespace == namespace && item.Mode == mode && item.Revision == revision {
			return
		}
	}
	t.Fatalf("missing namespace metric %s mode %s revision %s in %#v", namespace, mode, revision, namespaces)
}

func assertSecretIssueMetric(t *testing.T, issues []metricsSecretIssue, namespace string, gateway string, secret string, reason string) {
	t.Helper()
	for _, item := range issues {
		if item.Namespace == namespace && item.Gateway == gateway && item.Secret == secret && item.Reason == reason {
			return
		}
	}
	t.Fatalf("missing secret issue metric %s/%s/%s reason %s in %#v", namespace, gateway, secret, reason, issues)
}

func testCertificatePEM(t *testing.T, notBefore time.Time, notAfter time.Time) []byte {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	template := x509.Certificate{
		SerialNumber:          bigOne(),
		Subject:               pkix.Name{CommonName: "example.test"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}

	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func bigOne() *big.Int {
	return big.NewInt(1)
}

type metricsResource = metrics.IstioResourceMetric
type metricsNamespaceInjection = metrics.IstioNamespaceInjectionMetric
type metricsSecretIssue = metrics.IstioGatewayTLSSecretIssueMetric
