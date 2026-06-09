package istio_bridge

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"istio.io/api/networking/v1alpha3"
)

type gatewayTLSMaterial struct {
	cert string
	key  string
	ca   string
}

func gatewayTLSMode(mode string) (v1alpha3.ServerTLSSettings_TLSmode, string, error) {
	switch strings.ToUpper(strings.TrimSpace(mode)) {
	case "", "SIMPLE":
		return v1alpha3.ServerTLSSettings_SIMPLE, "SIMPLE", nil
	case "MUTUAL":
		return v1alpha3.ServerTLSSettings_MUTUAL, "MUTUAL", nil
	default:
		return v1alpha3.ServerTLSSettings_SIMPLE, "", fmt.Errorf("Unsupported TLS mode: %s", mode)
	}
}

func decodeGatewayTLSField(value string, field string, port uint32) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("Invalid base64 data for %s for port: %d", field, port)
	}

	return string(decoded), nil
}

func splitGatewayPEM(value string) (cert string, key string, err error) {
	rest := []byte(value)
	certBlocks := make([][]byte, 0)
	keyBlocks := make([][]byte, 0)

	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		encoded := pem.EncodeToMemory(block)
		if encoded == nil {
			continue
		}

		switch block.Type {
		case "CERTIFICATE":
			certBlocks = append(certBlocks, encoded)
		case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
			keyBlocks = append(keyBlocks, encoded)
		}
	}

	if len(bytes.TrimSpace(rest)) > 0 {
		return "", "", fmt.Errorf("Invalid PEM data")
	}

	return string(bytes.Join(certBlocks, nil)), string(bytes.Join(keyBlocks, nil)), nil
}

func normalizeGatewayTLSMaterial(port gateway_manager.PortsRequest) (gatewayTLSMaterial, error) {
	_, modeName, err := gatewayTLSMode(port.Mode)
	if err != nil {
		return gatewayTLSMaterial{}, err
	}

	certValue, err := decodeGatewayTLSField(port.Cert, "cert", port.Port)
	if err != nil {
		return gatewayTLSMaterial{}, err
	}
	keyValue, err := decodeGatewayTLSField(port.Pkey, "pkey", port.Port)
	if err != nil {
		return gatewayTLSMaterial{}, err
	}
	caValue, err := decodeGatewayTLSField(port.Cacert, "cacert", port.Port)
	if err != nil {
		return gatewayTLSMaterial{}, err
	}

	certsFromCert, keyFromCert, err := splitGatewayPEM(certValue)
	if err != nil {
		return gatewayTLSMaterial{}, fmt.Errorf("Invalid cert PEM for port %d: %s", port.Port, err.Error())
	}
	certsFromKey, keyFromKey, err := splitGatewayPEM(keyValue)
	if err != nil {
		return gatewayTLSMaterial{}, fmt.Errorf("Invalid pkey PEM for port %d: %s", port.Port, err.Error())
	}
	certsFromCA, keyFromCA, err := splitGatewayPEM(caValue)
	if err != nil {
		return gatewayTLSMaterial{}, fmt.Errorf("Invalid cacert PEM for port %d: %s", port.Port, err.Error())
	}
	if keyFromCA != "" {
		return gatewayTLSMaterial{}, fmt.Errorf("Invalid cacert PEM for port %d: private key is not allowed in CA bundle", port.Port)
	}

	certPEM := certsFromCert
	if certPEM == "" {
		certPEM = certsFromKey
	}
	keyPEM := keyFromKey
	if keyPEM == "" {
		keyPEM = keyFromCert
	}

	if strings.TrimSpace(certPEM) == "" {
		return gatewayTLSMaterial{}, fmt.Errorf("Port %d required certificate PEM", port.Port)
	}
	if strings.TrimSpace(keyPEM) == "" {
		return gatewayTLSMaterial{}, fmt.Errorf("Port %d required private key PEM", port.Port)
	}
	if modeName == "MUTUAL" && strings.TrimSpace(certsFromCA) == "" {
		return gatewayTLSMaterial{}, fmt.Errorf("Port %d required CA certificate bundle for mTLS", port.Port)
	}

	if _, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM)); err != nil {
		return gatewayTLSMaterial{}, fmt.Errorf("Invalid certification or private key, %s", err.Error())
	}

	return gatewayTLSMaterial{
		cert: certPEM,
		key:  keyPEM,
		ca:   certsFromCA,
	}, nil
}
