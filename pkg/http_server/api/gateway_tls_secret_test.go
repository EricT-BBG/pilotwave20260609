package api

import (
	"testing"

	"github.com/spf13/viper"
)

func TestResolveGatewayTLSCredentialSecretUsesConfiguredNamespace(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()

	namespace, name := resolveGatewayTLSCredentialSecret("wildcard-cert")
	if namespace != "istio-system" || name != "wildcard-cert" {
		t.Fatalf("unexpected Secret reference %s/%s", namespace, name)
	}
}

func TestResolveGatewayTLSCredentialSecretAllowsExplicitNamespace(t *testing.T) {
	viper.Set("gateway.tls_secret_namespace", "istio-system")
	defer viper.Reset()

	namespace, name := resolveGatewayTLSCredentialSecret("custom-ns/wildcard-cert")
	if namespace != "custom-ns" || name != "wildcard-cert" {
		t.Fatalf("unexpected Secret reference %s/%s", namespace, name)
	}
}
