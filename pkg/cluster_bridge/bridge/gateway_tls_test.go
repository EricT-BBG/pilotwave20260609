package cluster_bridge

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	istionetworkingapi "istio.io/api/networking/v1alpha3"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestGetGatewayTLSCertificatesReturnsMetadataOnly(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()
	now := time.Now().UTC()
	notAfter := now.Add(18 * 24 * time.Hour)
	istioClient := istiofake.NewSimpleClientset()
	_, err := istioClient.NetworkingV1alpha3().Gateways("edge").Create(context.Background(), &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{APIVersion: "networking.istio.io/v1alpha3", Kind: "Gateway"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "edge-gateway", Namespace: "edge"},
		Spec: istionetworkingapi.Gateway{
			Servers: []*istionetworkingapi.Server{
				{
					Hosts: []string{"app.example.local"},
					Port:  &istionetworkingapi.Port{Name: "https", Number: 443, Protocol: "HTTPS"},
					Tls:   &istionetworkingapi.ServerTLSSettings{CredentialName: "wildcard-cert"},
				},
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset(&k8scorev1.Secret{
		ObjectMeta: k8smetav1.ObjectMeta{Name: "wildcard-cert", Namespace: "istio-system"},
		Type:       k8scorev1.SecretTypeTLS,
		Data: map[string][]byte{
			k8scorev1.TLSCertKey:       testCertificatePEM(t, now.Add(-time.Hour), notAfter),
			k8scorev1.TLSPrivateKeyKey: []byte("do-not-return-this-key"),
		},
	})

	items, err := collectGatewayTLSCertificates(context.Background(), istioClient, k8sClient, "edge-gateway", "edge", func() time.Time {
		return now
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one certificate, got %d", len(items))
	}

	item := items[0]
	if item.Status != gatewayTLSStatusWarning {
		t.Fatalf("expected warning status, got %#v", item)
	}
	if item.DaysUntilExpiry < 17 || item.DaysUntilExpiry > 18 {
		t.Fatalf("unexpected days until expiry: %#v", item)
	}
	if item.SecretNamespace != "istio-system" || item.SecretName != "wildcard-cert" {
		t.Fatalf("unexpected secret reference: %#v", item)
	}
	if item.NotAfter == "" || item.Subject == "" || item.Issuer == "" || item.FingerprintSHA256 == "" {
		t.Fatalf("expected certificate metadata, got %#v", item)
	}
	if strings.Contains(item.Subject+item.Issuer+item.FingerprintSHA256, "do-not-return-this-key") {
		t.Fatalf("private key leaked in certificate metadata: %#v", item)
	}
}

func TestGetGatewayTLSCertificatesUsesExplicitCredentialNamespace(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()
	now := time.Now().UTC()
	notAfter := now.Add(45 * 24 * time.Hour)
	istioClient := istiofake.NewSimpleClientset()
	_, err := istioClient.NetworkingV1alpha3().Gateways("edge").Create(context.Background(), &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{APIVersion: "networking.istio.io/v1alpha3", Kind: "Gateway"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "edge-gateway", Namespace: "edge"},
		Spec: istionetworkingapi.Gateway{
			Servers: []*istionetworkingapi.Server{
				{
					Port: &istionetworkingapi.Port{Name: "https", Number: 443, Protocol: "HTTPS"},
					Tls:  &istionetworkingapi.ServerTLSSettings{CredentialName: "custom-ns/wildcard-cert"},
				},
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset(&k8scorev1.Secret{
		ObjectMeta: k8smetav1.ObjectMeta{Name: "wildcard-cert", Namespace: "custom-ns"},
		Type:       k8scorev1.SecretTypeTLS,
		Data: map[string][]byte{
			k8scorev1.TLSCertKey: testCertificatePEM(t, now.Add(-time.Hour), notAfter),
		},
	})

	items, err := collectGatewayTLSCertificates(context.Background(), istioClient, k8sClient, "edge-gateway", "edge", func() time.Time {
		return now
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one certificate, got %d", len(items))
	}
	if items[0].Status != gatewayTLSStatusHealthy {
		t.Fatalf("expected healthy status, got %#v", items[0])
	}
	if items[0].SecretNamespace != "custom-ns" || items[0].SecretName != "wildcard-cert" {
		t.Fatalf("unexpected secret reference: %#v", items[0])
	}
}

func TestGetGatewayTLSCertificatesReportsMissingAndInvalidSecrets(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()
	istioClient := istiofake.NewSimpleClientset()
	_, err := istioClient.NetworkingV1alpha3().Gateways("edge").Create(context.Background(), &istionetworking.Gateway{
		TypeMeta:   k8smetav1.TypeMeta{APIVersion: "networking.istio.io/v1alpha3", Kind: "Gateway"},
		ObjectMeta: k8smetav1.ObjectMeta{Name: "edge-gateway", Namespace: "edge"},
		Spec: istionetworkingapi.Gateway{
			Servers: []*istionetworkingapi.Server{
				{
					Port: &istionetworkingapi.Port{Name: "https", Number: 443, Protocol: "HTTPS"},
					Tls:  &istionetworkingapi.ServerTLSSettings{CredentialName: "missing-cert"},
				},
				{
					Port: &istionetworkingapi.Port{Name: "https-alt", Number: 8443, Protocol: "HTTPS"},
					Tls:  &istionetworkingapi.ServerTLSSettings{CredentialName: "invalid-cert"},
				},
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	k8sClient := k8sfake.NewSimpleClientset(&k8scorev1.Secret{
		ObjectMeta: k8smetav1.ObjectMeta{Name: "invalid-cert", Namespace: "istio-system"},
		Data: map[string][]byte{
			k8scorev1.TLSCertKey: []byte("not a certificate"),
		},
	})
	items, err := collectGatewayTLSCertificates(context.Background(), istioClient, k8sClient, "edge-gateway", "edge", time.Now)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two certificate statuses, got %d", len(items))
	}
	if items[0].Status != gatewayTLSStatusMissing || items[0].Reason != "not_found" {
		t.Fatalf("expected missing status, got %#v", items[0])
	}
	if items[0].SecretNamespace != "istio-system" || items[0].SecretName != "missing-cert" {
		t.Fatalf("expected missing status to report configured secret reference, got %#v", items[0])
	}
	if items[1].Status != gatewayTLSStatusInvalid || items[1].Reason != "parse_error" {
		t.Fatalf("expected invalid status, got %#v", items[1])
	}
	if items[1].SecretNamespace != "istio-system" || items[1].SecretName != "invalid-cert" {
		t.Fatalf("expected invalid status to report configured secret reference, got %#v", items[1])
	}
}

func TestDefaultGatewayTLSSecretNamespaceFallsBackToIstioSystem(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "")
	defer viper.Reset()

	if got := defaultGatewayTLSSecretNamespace(); got != "istio-system" {
		t.Fatalf("expected istio-system fallback, got %q", got)
	}
}
