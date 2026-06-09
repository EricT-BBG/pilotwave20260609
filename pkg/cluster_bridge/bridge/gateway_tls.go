package cluster_bridge

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"math"
	"strings"
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "k8s.io/client-go/kubernetes"
)

const (
	gatewayTLSStatusHealthy  = "healthy"
	gatewayTLSStatusWarning  = "warning"
	gatewayTLSStatusCritical = "critical"
	gatewayTLSStatusExpired  = "expired"
	gatewayTLSStatusMissing  = "missing"
	gatewayTLSStatusInvalid  = "invalid"
	gatewayTLSStatusUnknown  = "unknown"
)

type certificateSummary struct {
	notBefore         time.Time
	notAfter          time.Time
	subject           string
	issuer            string
	dnsNames          []string
	fingerprintSHA256 string
}

func (bridge *ClusterBridge) getGatewayTLSCertificates(ctx context.Context, name string, namespace string) ([]gateway_manager.GatewayTLSCertificateResponse, error) {
	return collectGatewayTLSCertificates(ctx, bridge.istioClientset, bridge.k8sClientset, name, namespace, time.Now)
}

func collectGatewayTLSCertificates(ctx context.Context, istioClientset istioclient.Interface, k8sClientset k8sclient.Interface, name string, namespace string, nowFunc func() time.Time) ([]gateway_manager.GatewayTLSCertificateResponse, error) {
	gateway, err := istioClientset.NetworkingV1alpha3().Gateways(namespace).Get(ctx, name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	now := nowFunc().UTC()
	results := make([]gateway_manager.GatewayTLSCertificateResponse, 0)
	for serverIndex, server := range gateway.Spec.GetServers() {
		tlsSettings := server.GetTls()
		if tlsSettings == nil {
			continue
		}

		port := uint32(0)
		protocol := ""
		if server.GetPort() != nil {
			port = server.GetPort().GetNumber()
			protocol = server.GetPort().GetProtocol()
		}

		credentialName := strings.TrimSpace(tlsSettings.GetCredentialName())
		item := gateway_manager.GatewayTLSCertificateResponse{
			ServerIndex:     serverIndex,
			Port:            port,
			Protocol:        protocol,
			Hosts:           server.GetHosts(),
			CredentialName:  credentialName,
			Status:          gatewayTLSStatusUnknown,
			Reason:          "missing_credential_name",
			DaysUntilExpiry: 0,
		}

		if credentialName == "" {
			results = append(results, item)
			continue
		}

		secretNamespace, secretName := resolveCredentialSecret(defaultGatewayTLSSecretNamespace(), credentialName)
		item.SecretNamespace = secretNamespace
		item.SecretName = secretName

		secret, err := k8sClientset.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, k8smetav1.GetOptions{})
		if err != nil {
			item.Status = gatewayTLSStatusMissing
			item.Reason = "read_error"
			if k8serrors.IsNotFound(err) {
				item.Reason = "not_found"
			}
			results = append(results, item)
			continue
		}

		summary, err := parseTLSCertificateSummary(secret)
		if err != nil {
			item.Status = gatewayTLSStatusInvalid
			item.Reason = err.Error()
			results = append(results, item)
			continue
		}

		daysUntilExpiry := int(math.Floor(summary.notAfter.Sub(now).Hours() / 24))
		item.Status = gatewayTLSCertificateStatus(summary.notAfter, daysUntilExpiry, now)
		item.Reason = ""
		item.DaysUntilExpiry = daysUntilExpiry
		item.NotBefore = summary.notBefore.UTC().Format(time.RFC3339)
		item.NotAfter = summary.notAfter.UTC().Format(time.RFC3339)
		item.Subject = summary.subject
		item.Issuer = summary.issuer
		item.DNSNames = summary.dnsNames
		item.FingerprintSHA256 = summary.fingerprintSHA256
		results = append(results, item)
	}

	return results, nil
}

func gatewayTLSCertificateStatus(notAfter time.Time, daysUntilExpiry int, now time.Time) string {
	if !notAfter.After(now) {
		return gatewayTLSStatusExpired
	}
	if daysUntilExpiry <= 7 {
		return gatewayTLSStatusCritical
	}
	if daysUntilExpiry <= 30 {
		return gatewayTLSStatusWarning
	}
	return gatewayTLSStatusHealthy
}

func parseTLSCertificateSummary(secret *k8scorev1.Secret) (certificateSummary, error) {
	certData := secret.Data[k8scorev1.TLSCertKey]
	if len(certData) == 0 {
		certData = secret.Data[istioSecretCertKey]
	}
	if len(certData) == 0 {
		return certificateSummary{}, metricReasonError("missing_certificate")
	}

	var selected *x509.Certificate
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
			return certificateSummary{}, metricReasonError("parse_error")
		}
		if selected == nil || cert.NotAfter.Before(selected.NotAfter) {
			selected = cert
		}
	}

	if selected == nil {
		return certificateSummary{}, metricReasonError("parse_error")
	}

	fingerprint := sha256.Sum256(selected.Raw)
	return certificateSummary{
		notBefore:         selected.NotBefore,
		notAfter:          selected.NotAfter,
		subject:           selected.Subject.String(),
		issuer:            selected.Issuer.String(),
		dnsNames:          append([]string{}, selected.DNSNames...),
		fingerprintSHA256: colonHex(fingerprint[:]),
	}, nil
}

func colonHex(data []byte) string {
	encoded := strings.ToUpper(hex.EncodeToString(data))
	if encoded == "" {
		return ""
	}
	parts := make([]string, 0, len(encoded)/2)
	for i := 0; i < len(encoded); i += 2 {
		parts = append(parts, encoded[i:i+2])
	}
	return strings.Join(parts, ":")
}
