package istio_bridge

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	istioapi "istio.io/api/networking/v1alpha3"
	istionetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stesting "k8s.io/client-go/testing"
)

type recordingSecretBridge struct {
	cluster_bridge.Bridge

	existingSecrets map[string]bool
	deletedSecrets  []string
	createdSecrets  []string
	secretData      map[string]recordedSecretData
}

type recordedSecretData struct {
	certificate string
	privateKey  string
	ca          string
}

func (b *recordingSecretBridge) secretKey(name string, namespace string) string {
	return namespace + "/" + name
}

func (b *recordingSecretBridge) SecretsExist(name string, namespace string) (bool, error) {
	return b.existingSecrets[b.secretKey(name, namespace)], nil
}

func (b *recordingSecretBridge) DeleteSecrets(name string, namespace string) error {
	b.deletedSecrets = append(b.deletedSecrets, b.secretKey(name, namespace))
	return nil
}

func (b *recordingSecretBridge) CreateSecrets(name string, namespace string, certificate string, privateKey string, caCertificate string) error {
	key := b.secretKey(name, namespace)
	b.createdSecrets = append(b.createdSecrets, key)
	if b.secretData == nil {
		b.secretData = map[string]recordedSecretData{}
	}
	b.secretData[key] = recordedSecretData{certificate: certificate, privateKey: privateKey, ca: caCertificate}
	return nil
}

type testApp struct {
	app.App

	clusterBridge cluster_bridge.Bridge
}

func (a *testApp) GetClusterBridge() cluster_bridge.Bridge {
	return a.clusterBridge
}

func TestUpdateGatewayTLSCertificateReplacementCreatesNewManagedSecretAndDeletesOld(t *testing.T) {
	gateway := &istionetworking.Gateway{
		TypeMeta: k8smetav1.TypeMeta{Kind: "Gateway", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:            "gw-tls",
			Namespace:       "default",
			ResourceVersion: "9",
		},
		Spec: istioapi.Gateway{
			Selector: map[string]string{"istio": "ingressgateway"},
			Servers: []*istioapi.Server{{
				Hosts: []string{"old.example.local"},
				Port:  &istioapi.Port{Name: "https", Number: 443, Protocol: "HTTPS"},
				Tls:   &istioapi.ServerTLSSettings{Mode: istioapi.ServerTLSSettings_SIMPLE, CredentialName: "pilotwave-old-cert"},
			}},
		},
	}
	br, client := newTestBridge(gateway)
	secretBridge := &recordingSecretBridge{existingSecrets: map[string]bool{}}
	br.app = &testApp{clusterBridge: secretBridge}
	cert, key := newBase64TLSCertificate(t, "new.example.local")

	err := br.UpdateGateway("gw-tls", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"new.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Cert:     cert,
				Pkey:     key,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("UpdateGateway returned error: %v", err)
	}

	assertPatchFirst(t, client, "gateways")
	if len(secretBridge.createdSecrets) != 1 {
		t.Fatalf("expected one generated TLS secret, got %v", secretBridge.createdSecrets)
	}
	newSecretKey := secretBridge.createdSecrets[0]
	if !strings.HasPrefix(newSecretKey, "istio-system/pilotwave-gw-tls-istio-system-port-443") {
		t.Fatalf("generated TLS secret key has unexpected name: %s", newSecretKey)
	}
	assertGatewayPatchTLSCredential(t, client, strings.TrimPrefix(newSecretKey, "istio-system/"))

	wantDeleted := []string{newSecretKey, "istio-system/pilotwave-old-cert"}
	if !reflect.DeepEqual(secretBridge.deletedSecrets, wantDeleted) {
		t.Fatalf("deleted secrets mismatch: got %v want %v", secretBridge.deletedSecrets, wantDeleted)
	}
}

func TestCreateGatewayMutualTLSCreatesSecretWithCABundleAndMode(t *testing.T) {
	br, client := newTestBridge()
	secretBridge := &recordingSecretBridge{existingSecrets: map[string]bool{}}
	br.app = &testApp{clusterBridge: secretBridge}
	cert, key := newBase64TLSCertificate(t, "mtls.example.local")
	ca, _ := newBase64TLSCertificate(t, "client-ca.example.local")

	err := br.CreateGateway("gw-mtls", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"mtls.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Mode:     "MUTUAL",
				Cert:     cert,
				Pkey:     key,
				Cacert:   ca,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("CreateGateway returned error: %v", err)
	}

	if len(secretBridge.createdSecrets) != 1 {
		t.Fatalf("expected one generated TLS secret, got %v", secretBridge.createdSecrets)
	}
	secret := secretBridge.secretData[secretBridge.createdSecrets[0]]
	if secret.certificate == "" || secret.privateKey == "" || secret.ca == "" {
		t.Fatalf("expected TLS secret certificate, private key, and CA bundle, got %#v", secret)
	}
	assertGatewayCreateTLSMode(t, client, "MUTUAL")
}

func TestCreateGatewayCombinedPEMSplitsCertificateAndKey(t *testing.T) {
	br, client := newTestBridge()
	secretBridge := &recordingSecretBridge{existingSecrets: map[string]bool{}}
	br.app = &testApp{clusterBridge: secretBridge}
	cert, key := newTLSCertificate(t, "combined.example.local")
	combined := base64.StdEncoding.EncodeToString([]byte(cert + key))

	err := br.CreateGateway("gw-combined", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"combined.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Cert:     combined,
			}},
		}},
	})
	if err != nil {
		t.Fatalf("CreateGateway returned error: %v", err)
	}

	secret := secretBridge.secretData[secretBridge.createdSecrets[0]]
	if secret.certificate != cert {
		t.Fatalf("combined PEM certificate was not split: got %q want %q", secret.certificate, cert)
	}
	if secret.privateKey != key {
		t.Fatalf("combined PEM key was not split")
	}
	assertGatewayCreateTLSMode(t, client, "SIMPLE")
}

func TestCreateGatewayMutualTLSRequiresCABundle(t *testing.T) {
	br, _ := newTestBridge()
	br.app = &testApp{clusterBridge: &recordingSecretBridge{existingSecrets: map[string]bool{}}}
	cert, key := newBase64TLSCertificate(t, "missing-ca.example.local")

	err := br.CreateGateway("gw-missing-ca", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"missing-ca.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Mode:     "MUTUAL",
				Cert:     cert,
				Pkey:     key,
			}},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "required CA certificate bundle") {
		t.Fatalf("expected missing CA bundle error, got %v", err)
	}
}

func TestCreateGatewayTLSRejectsInvalidBase64(t *testing.T) {
	br, _ := newTestBridge()
	br.app = &testApp{clusterBridge: &recordingSecretBridge{existingSecrets: map[string]bool{}}}
	_, key := newBase64TLSCertificate(t, "invalid-base64.example.local")

	err := br.CreateGateway("gw-invalid-base64", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"invalid-base64.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Cert:     "not-base64-%%%/",
				Pkey:     key,
			}},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "Invalid base64 data for cert") {
		t.Fatalf("expected invalid base64 error, got %v", err)
	}
}

func TestCreateGatewayTLSRequiresPrivateKey(t *testing.T) {
	br, _ := newTestBridge()
	br.app = &testApp{clusterBridge: &recordingSecretBridge{existingSecrets: map[string]bool{}}}
	cert, _ := newBase64TLSCertificate(t, "missing-key.example.local")

	err := br.CreateGateway("gw-missing-key", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"missing-key.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Cert:     cert,
			}},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "required private key PEM") {
		t.Fatalf("expected missing private key error, got %v", err)
	}
}

func TestCreateGatewayTLSRejectsMismatchedCertificateAndKey(t *testing.T) {
	br, _ := newTestBridge()
	br.app = &testApp{clusterBridge: &recordingSecretBridge{existingSecrets: map[string]bool{}}}
	cert, _ := newBase64TLSCertificate(t, "mismatch.example.local")
	_, otherKey := newBase64TLSCertificate(t, "other.example.local")

	err := br.CreateGateway("gw-mismatch", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"mismatch.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:     443,
				Protocol: "HTTPS",
				Cert:     cert,
				Pkey:     otherKey,
			}},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "Invalid certification or private key") {
		t.Fatalf("expected certificate/key mismatch error, got %v", err)
	}
}

func TestCreateGatewayExistingSecretPreservesRequestedMode(t *testing.T) {
	br, client := newTestBridge()
	secretBridge := &recordingSecretBridge{
		existingSecrets: map[string]bool{"istio-system/existing-mtls": true},
	}
	br.app = &testApp{clusterBridge: secretBridge}

	err := br.CreateGateway("gw-existing", "default", &gateway_manager.GatewayRequest{
		Servers: []gateway_manager.GatewayRequestServersData{{
			Hosts: []string{"existing.example.local"},
			Ports: []gateway_manager.PortsRequest{{
				Port:           443,
				Protocol:       "HTTPS",
				Mode:           "MUTUAL",
				CredentialName: "existing-mtls",
			}},
		}},
	})
	if err != nil {
		t.Fatalf("CreateGateway returned error: %v", err)
	}

	assertGatewayCreateTLSMode(t, client, "MUTUAL")
	if len(secretBridge.createdSecrets) != 0 || len(secretBridge.deletedSecrets) != 0 {
		t.Fatalf("existing Secret mode should not create or delete secrets: created=%v deleted=%v", secretBridge.createdSecrets, secretBridge.deletedSecrets)
	}
}

func newBase64TLSCertificate(t *testing.T, host string) (string, string) {
	t.Helper()

	certPEM, keyPEM := newTLSCertificate(t, host)
	return base64.StdEncoding.EncodeToString([]byte(certPEM)), base64.StdEncoding.EncodeToString([]byte(keyPEM))
}

func newTLSCertificate(t *testing.T, host string) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate test TLS key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{host},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("generate test TLS certificate: %v", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return string(certPEM), string(keyPEM)
}

func assertGatewayPatchTLSCredential(t *testing.T, client *istiofake.Clientset, expected string) {
	t.Helper()

	data := patchDataForResource(t, client, "gateways")
	var patch struct {
		Spec struct {
			Servers []struct {
				TLS struct {
					CredentialName string `json:"credentialName"`
				} `json:"tls"`
			} `json:"servers"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(data, &patch); err != nil {
		t.Fatalf("invalid gateway patch json: %v", err)
	}
	if len(patch.Spec.Servers) != 1 {
		t.Fatalf("expected one patched gateway server, got %d in %s", len(patch.Spec.Servers), string(data))
	}
	if patch.Spec.Servers[0].TLS.CredentialName != expected {
		t.Fatalf("expected TLS credentialName %q, got %q in %s", expected, patch.Spec.Servers[0].TLS.CredentialName, string(data))
	}
}

func assertGatewayCreateTLSMode(t *testing.T, client *istiofake.Clientset, expected string) {
	t.Helper()

	var created *istionetworking.Gateway
	for _, action := range client.Actions() {
		if action.GetVerb() != "create" || action.GetResource().Resource != "gateways" {
			continue
		}
		createAction, ok := action.(k8stesting.CreateAction)
		if !ok {
			t.Fatalf("create action has unexpected type %T", action)
		}
		var okGateway bool
		created, okGateway = createAction.GetObject().(*istionetworking.Gateway)
		if !okGateway {
			t.Fatalf("create object has unexpected type %T", createAction.GetObject())
		}
		break
	}
	if created == nil || len(created.Spec.GetServers()) != 1 || created.Spec.GetServers()[0].GetTls() == nil {
		t.Fatalf("expected created gateway with one TLS server; actions: %s", actionSummary(client.Actions()))
	}
	if got := created.Spec.GetServers()[0].GetTls().GetMode().String(); got != expected {
		t.Fatalf("created gateway TLS mode mismatch: got %s want %s", got, expected)
	}
}

func TestDeleteRouterGatewayMappingPatchesOnlyGatewaysAndPreservesVirtualServiceCustomFields(t *testing.T) {
	vs := &istionetworking.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:            "rt-a",
			Namespace:       "default",
			ResourceVersion: "2",
		},
		Spec: istioapi.VirtualService{
			Hosts:    []string{"app.example.local"},
			Gateways: []string{"gw-a", "other-ns/gw-b"},
			Http: []*istioapi.HTTPRoute{{
				Name: "route-a",
				Route: []*istioapi.HTTPRouteDestination{{
					Destination: &istioapi.Destination{Host: "svc.default.svc.cluster.local"},
					Weight:      100,
				}},
			}},
		},
	}
	br, client := newTestBridge(vs)

	if err := br.DeleteRouterGatewayMapping("rt-a", "default"); err != nil {
		t.Fatalf("DeleteRouterGatewayMapping returned error: %v", err)
	}

	assertPatchFirst(t, client, "virtualservices")
	assertPatchResourceVersion(t, client, "virtualservices", "2")
	assertVirtualServicePatchGateways(t, client, []string{})
	assertVirtualServicePatchPreservesCustomFields(t, client)
}

func TestDeleteGatewayRouterMappingPatchesMappedVirtualServicesAndPreservesCustomFields(t *testing.T) {
	mapped := &istionetworking.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:            "rt-mapped",
			Namespace:       "default",
			ResourceVersion: "5",
		},
		Spec: istioapi.VirtualService{
			Hosts:    []string{"mapped.example.local"},
			Gateways: []string{"gw-a", "other-ns/gw-b"},
		},
	}
	unmapped := &istionetworking.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{Kind: "VirtualService", APIVersion: "networking.istio.io/v1alpha3"},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:            "rt-unmapped",
			Namespace:       "default",
			ResourceVersion: "6",
		},
		Spec: istioapi.VirtualService{
			Hosts:    []string{"unmapped.example.local"},
			Gateways: []string{"other-ns/gw-b"},
		},
	}
	br, client := newTestBridge(mapped, unmapped)

	if err := br.DeleteGatewayRouterMapping("gw-a", "default"); err != nil {
		t.Fatalf("DeleteGatewayRouterMapping returned error: %v", err)
	}

	assertPatchFirst(t, client, "virtualservices")
	assertOnlyPatchedVirtualServices(t, client, []string{"rt-mapped"})
	assertPatchResourceVersion(t, client, "virtualservices", "5")
	assertVirtualServicePatchGateways(t, client, []string{"other-ns/gw-b"})
	assertVirtualServicePatchPreservesCustomFields(t, client)
}

func assertVirtualServicePatchGateways(t *testing.T, client *istiofake.Clientset, expected []string) {
	t.Helper()

	data := patchDataForResource(t, client, "virtualservices")
	var patch struct {
		Spec struct {
			Gateways []string `json:"gateways"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(data, &patch); err != nil {
		t.Fatalf("invalid virtualservice patch json: %v", err)
	}
	if !reflect.DeepEqual(patch.Spec.Gateways, expected) {
		t.Fatalf("patched gateways mismatch: got %v want %v in %s", patch.Spec.Gateways, expected, string(data))
	}
}

func assertVirtualServicePatchPreservesCustomFields(t *testing.T, client *istiofake.Clientset) {
	t.Helper()

	data := patchDataForResource(t, client, "virtualservices")
	var patch map[string]interface{}
	if err := json.Unmarshal(data, &patch); err != nil {
		t.Fatalf("invalid virtualservice patch json: %v", err)
	}

	object := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels":      map[string]interface{}{"pilotwave.io/custom-label": "keep"},
			"annotations": map[string]interface{}{"pilotwave.io/custom-annotation": "keep"},
		},
		"spec": map[string]interface{}{
			"hosts":                     []interface{}{"app.example.local"},
			"http":                      []interface{}{map[string]interface{}{"name": "route-a"}},
			"pilotwave.io/customObject": map[string]interface{}{"nested": "keep"},
			"pilotwave.io/customList":   []interface{}{"keep"},
		},
	}
	merged := applyJSONMergePatch(object, patch)
	spec := merged["spec"].(map[string]interface{})
	if _, ok := spec["http"]; !ok {
		t.Fatalf("http routes were not preserved by virtualservice patch: %v", spec)
	}
	customObject := spec["pilotwave.io/customObject"].(map[string]interface{})
	if customObject["nested"] != "keep" {
		t.Fatalf("custom object was not preserved by virtualservice patch: %v", spec)
	}
	if _, ok := spec["pilotwave.io/customList"]; !ok {
		t.Fatalf("custom list was not preserved by virtualservice patch: %v", spec)
	}
}

func assertOnlyPatchedVirtualServices(t *testing.T, client *istiofake.Clientset, expectedNames []string) {
	t.Helper()

	names := make([]string, 0, len(expectedNames))
	for _, action := range client.Actions() {
		if action.GetVerb() != "patch" || action.GetResource().Resource != "virtualservices" {
			continue
		}
		patchAction, ok := action.(k8stesting.PatchAction)
		if !ok {
			t.Fatalf("patch action has unexpected type %T", action)
		}
		names = append(names, patchAction.GetName())
	}
	if !reflect.DeepEqual(names, expectedNames) {
		t.Fatalf("patched virtualservices mismatch: got %v want %v; actions: %s", names, expectedNames, actionSummary(client.Actions()))
	}
}

func actionSummary(actions []k8stesting.Action) string {
	result := ""
	for i, action := range actions {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s/%s/%s", action.GetVerb(), action.GetResource().Resource, action.GetSubresource())
	}
	return result
}
