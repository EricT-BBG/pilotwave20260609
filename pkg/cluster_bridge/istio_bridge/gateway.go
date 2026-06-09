package istio_bridge

import (
	"context"
	"errors"
	"fmt"
	// "log"
	"reflect"
	"strconv"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/gateway_manager"
	"istio.io/api/networking/v1alpha3"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/spf13/viper"
)

const (
	DEFAULT_ISTIO_SECRET_NAMESPACE = "istio-system"
	MANAGED_GATEWAY_SECRET_PREFIX  = "pilotwave-"
)

func composeGatewayBaseData(name string, namespace string) *istioNetworking.Gateway {

	virtualGateway := &istioNetworking.Gateway{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "Gateway", // Gateway
			APIVersion: "networking.istio.io/v1alpha3",
		},

		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return virtualGateway
}

func composeGatewayTLSBaseName(name string, namespace string, port uint32) string {
	return "pilotwave-" + name + "-" + namespace + "-port-" + strconv.FormatInt(int64(port), 10) + strings.ToLower(RandomString(5, ""))
}

func isManagedGatewaySecret(name string) bool {
	return strings.HasPrefix(name, MANAGED_GATEWAY_SECRET_PREFIX)
}

func appendUniqueData(data []gateway_manager.GatewayMappinRouterData, name string, namespace string) []gateway_manager.GatewayMappinRouterData {
	for _, item := range data {
		if item.Name == name && item.Namespace == namespace {
			return data
		}
	}
	data = append(data, gateway_manager.GatewayMappinRouterData{Name: name, Namespace: namespace})
	return data
}

func virtualServiceResourceKey(namespace string, name string) string {
	return namespace + "/" + name
}

func parseGatewayReference(ref string, defaultNamespace string) (string, string) {
	spiltResult := strings.Split(ref, "/")
	if len(spiltResult) == 2 {
		return spiltResult[0], spiltResult[1]
	}

	return defaultNamespace, ref
}

func composeGatewayReference(gatewayName string, gatewayNamespace string, virtualServiceNamespace string) string {
	if gatewayNamespace == virtualServiceNamespace {
		return gatewayName
	}

	return gatewayNamespace + "/" + gatewayName
}

func virtualServiceHasGateway(gateways []string, defaultNamespace string, gatewayName string, gatewayNamespace string) bool {
	for _, item := range gateways {
		currentNamespace, currentName := parseGatewayReference(item, defaultNamespace)
		if currentName == gatewayName && currentNamespace == gatewayNamespace {
			return true
		}
	}

	return false
}

func removeGatewayReference(gateways []string, defaultNamespace string, gatewayName string, gatewayNamespace string) []string {
	results := make([]string, 0, len(gateways))
	for _, item := range gateways {
		currentNamespace, currentName := parseGatewayReference(item, defaultNamespace)
		if currentName == gatewayName && currentNamespace == gatewayNamespace {
			continue
		}
		results = append(results, composeGatewayReference(currentName, currentNamespace, defaultNamespace))
	}

	return results
}

func isRequireTLSProtocol(protocol string) bool {
	table := map[string]bool{
		"https": true,
		"tls":   true,
	}

	_, ok := table[strings.ToLower(protocol)]
	if ok {
		return true
	}
	return false
}

func gatewayTLSSecretNamespace() string {
	namespace := strings.TrimSpace(viper.GetString("gateway.tls_secret_namespace"))
	if namespace == "" {
		return DEFAULT_ISTIO_SECRET_NAMESPACE
	}

	return namespace
}

func (br *IstioBridge) GetGatewaysFull(namespace string, name string, nameIsfuzzy bool) ([]gateway_manager.GatewayResponse, int, error) {
	ctx, cancelFn := context.WithCancel(context.Background())

	// List gateways

	opts := k8smetav1.ListOptions{}

	if !nameIsfuzzy && (len(name) > 0) {
		opts.FieldSelector = "metadata.name=" + name
	}

	gwsList, err := br.gatewayResources.ListGateways(ctx, namespace, opts)

	defer cancelFn()
	if err != nil {
		return []gateway_manager.GatewayResponse{}, 0, err
	}

	items := make([]gateway_manager.GatewayResponse, 0, len(gwsList.Items))

	for _, gwItem := range gwsList.Items {
		if name != "" {
			if nameIsfuzzy && !strings.Contains(gwItem.Name, name) {
				continue
			}
		}

		serversResult := make([]gateway_manager.GatewayServersData, 0, len(gwItem.Spec.GetServers()))

		hosts := make([]string, 0)
		ports := make([]gateway_manager.PortData, 0)
		servers := gwItem.Spec.GetServers()
		for i, serverItem := range servers {
			var portdata gateway_manager.PortData
			if serverItem.Port != nil {
				portdata = gateway_manager.PortData{
					Protocol: serverItem.Port.Protocol,
					Name:     serverItem.Port.Name,
					Port:     serverItem.Port.Number,
				}
			}

			if serverItem.Tls != nil {
				portdata.Pkey = serverItem.Tls.PrivateKey
				portdata.CredentialName = serverItem.Tls.CredentialName
				portdata.Cert = serverItem.Tls.ServerCertificate
				portdata.Mode = serverItem.Tls.Mode.String()
			}

			index := i - 1
			previous := serverItem
			if index < 0 {
				previous = nil
			} else {
				previous = servers[index]
			}

			result := gateway_manager.GatewayServersData{}

			hostIndex := 0
			if !reflect.DeepEqual(previous.GetHosts(), serverItem.GetHosts()) {
				hosts = serverItem.GetHosts()
				result.Hosts = hosts
				hostIndex = i // record hosts index in servers at first time
				ports = make([]gateway_manager.PortData, 0)
			}

			if serverItem.Port != nil {
				ports = append(ports, portdata)
				result.Ports = ports
			}

			if result.Hosts == nil {
				serversResult[hostIndex].Ports = ports // replace exists host's ports
			} else {
				serversResult = append(serversResult, result)
			}
		}

		v := gateway_manager.GatewayResponse{
			Name:                gwItem.Name,
			Namespace:           gwItem.Namespace,
			CreatedAt:           gwItem.CreationTimestamp.Unix(),
			Servers:             serversResult,
			SelectorMatchLabels: gwItem.Spec.Selector,
			ResourceVersion:     gwItem.ObjectMeta.ResourceVersion,
		}

		items = append(items, v)

	}

	return items, len(items), nil
}

func (br *IstioBridge) GetGateways(page int, perPage int, search string, namespace string) ([]gateway_manager.GatewayResponse, int, error) {
	data, total, err := br.GetGatewaysFull(namespace, search, true)
	if err != nil {
		return nil, 0, err
	}

	start := (page - 1) * perPage
	end := page * perPage

	if start > total || page < 1 {
		return nil, 0, nil
	}

	if perPage < 0 {
		return data[0:total], total, nil
	}

	if end > total {
		end = total
	}

	return data[start:end], total, nil
}

func (br *IstioBridge) GetGateway(name string, namespace string) (*gateway_manager.GatewayResponse, error) {
	res, total, err := br.GetGatewaysFull(namespace, name, false)

	if err != nil || total == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (br *IstioBridge) CreateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error {
	return br.CreateGatewayInternal(name, namespace, false, requestdata)
}

func (br *IstioBridge) UpdateGateway(name string, namespace string, requestdata *gateway_manager.GatewayRequest) error {
	return br.CreateGatewayInternal(name, namespace, true, requestdata)
}

func (br *IstioBridge) DeleteGateway(name string, namespace string) error {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	err := br.gatewayResources.DeleteGateway(ctx, namespace, name, k8smetav1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (br *IstioBridge) GetGatewayRouterMapping(name string, namespace string) (gateway_manager.RouterMappingResponse, error) {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// List virtual service
	vsList, err := br.routerResources.ListVirtualServices(ctx, "", k8smetav1.ListOptions{})
	if err != nil {
		return gateway_manager.RouterMappingResponse{}, err
	}

	routerItems := make([]gateway_manager.GatewayMappinRouterData, 0, len(vsList.Items))
	resourceVersions := make(map[string]string, len(vsList.Items))

	for _, item := range vsList.Items {
		resourceVersions[virtualServiceResourceKey(item.GetObjectMeta().GetNamespace(), item.GetObjectMeta().GetName())] = item.GetObjectMeta().GetResourceVersion()
		for _, gwItem := range item.Spec.GetGateways() {
			gatewayNamespace, gatewayName := parseGatewayReference(gwItem, item.GetObjectMeta().GetNamespace())

			if gatewayName == name && gatewayNamespace == namespace {
				routerItems = appendUniqueData(routerItems, item.GetObjectMeta().GetName(), item.GetObjectMeta().GetNamespace())
			}
		}
	}

	return gateway_manager.RouterMappingResponse{
			Name:             name,
			Namespace:        namespace,
			Routers:          routerItems,
			ResourceVersions: resourceVersions},
		nil
}

func (br *IstioBridge) CreateGatewayRouterMapping(name string, namespace string, routers []gateway_manager.GatewayMappinRouterData, resourceVersions map[string]string) error {

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	selectedRouters := make(map[string]bool, len(routers))
	for _, routeritem := range routers {
		if routeritem.Name == "" {
			return errors.New("Error in Router mapping, router name is empty")
		}

		vsNamespace := routeritem.Namespace
		if vsNamespace == "" {
			vsNamespace = namespace
		}

		selectedRouters[virtualServiceResourceKey(vsNamespace, routeritem.Name)] = true
	}

	vsList, err := br.routerResources.ListVirtualServices(ctx, "", k8smetav1.ListOptions{})
	if err != nil {
		return err
	}

	virtualServices := make(map[string]*istioNetworking.VirtualService, len(vsList.Items))
	for i := range vsList.Items {
		vsItem := &vsList.Items[i]
		key := virtualServiceResourceKey(vsItem.GetObjectMeta().GetNamespace(), vsItem.GetObjectMeta().GetName())
		virtualServices[key] = vsItem
	}

	for key := range selectedRouters {
		if _, ok := virtualServices[key]; !ok {
			return fmt.Errorf("Router for gateway mapping is not found - %s", key)
		}
	}

	for key, vsItem := range virtualServices {
		currentlyMapped := virtualServiceHasGateway(vsItem.Spec.GetGateways(), vsItem.GetObjectMeta().GetNamespace(), name, namespace)
		desiredMapped := selectedRouters[key]
		if currentlyMapped == desiredMapped {
			continue
		}

		if len(resourceVersions) > 0 && resourceVersions[key] != vsItem.GetObjectMeta().GetResourceVersion() {
			recordIstioStaleConflict(istioResourceVirtualService, kubernetesWriteVerbUpdate)
			return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "virtualservices"}, vsItem.GetObjectMeta().GetName(), fmt.Errorf("resource version changed"))
		}
	}

	for key, vsItem := range virtualServices {
		currentlyMapped := virtualServiceHasGateway(vsItem.Spec.GetGateways(), vsItem.GetObjectMeta().GetNamespace(), name, namespace)
		desiredMapped := selectedRouters[key]
		if currentlyMapped == desiredMapped {
			continue
		}

		if desiredMapped {
			vsItem.Spec.Gateways = append(vsItem.Spec.GetGateways(), composeGatewayReference(name, namespace, vsItem.GetObjectMeta().GetNamespace()))
		} else {
			vsItem.Spec.Gateways = removeGatewayReference(vsItem.Spec.GetGateways(), vsItem.GetObjectMeta().GetNamespace(), name, namespace)
		}

		ns := vsItem.GetObjectMeta().GetNamespace()
		metadata := map[string]interface{}{}
		if vsItem.GetObjectMeta().GetResourceVersion() != "" {
			metadata["resourceVersion"] = vsItem.GetObjectMeta().GetResourceVersion()
		}
		err = br.patchVirtualService(ctx, ns, vsItem.GetObjectMeta().GetName(), metadata, map[string]interface{}{
			"gateways": vsItem.Spec.GetGateways(),
		})
		if err != nil {
			return fmt.Errorf("Update router mapping failure for gateway, gateway name: %s, namespace: %s, router: %s. failure message: %s", name, namespace, key, err.Error())
		}
	}

	return nil

}

func (br *IstioBridge) UpdateGatewayRouterMapping(name string, namespace string, routers []gateway_manager.GatewayMappinRouterData, resourceVersions map[string]string) error {

	return br.CreateGatewayRouterMapping(name, namespace, routers, resourceVersions)
}

func (br *IstioBridge) DeleteGatewayRouterMapping(name string, namespace string) error {

	ctx, cancelFn := context.WithCancel(context.Background())

	defer cancelFn()

	vsList, err := br.routerResources.ListVirtualServices(context.TODO(), "", k8smetav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, vsItem := range vsList.Items {

		doModify := false
		ns := vsItem.GetObjectMeta().GetNamespace()
		newGatewayResult := make([]string, 0, len(vsItem.Spec.GetGateways()))

		for _, gwItem := range vsItem.Spec.GetGateways() {
			var gatewayNamespace string
			var gatewayName string
			var composedName string

			spiltResult := strings.Split(gwItem, "/")

			if len(spiltResult) == 2 {
				gatewayNamespace = spiltResult[0]
				gatewayName = spiltResult[1]
			} else {
				gatewayNamespace = vsItem.GetObjectMeta().GetNamespace()
				gatewayName = gwItem
			}

			if gatewayNamespace == ns {
				composedName = gatewayName
			} else {
				composedName = gatewayNamespace + "/" + gatewayName
			}

			if gatewayName == name && gatewayNamespace == namespace {
				doModify = true
			} else {
				newGatewayResult = append(newGatewayResult, composedName)
			}
		}

		if !doModify {
			continue
		}

		vsItem.Spec.Gateways = newGatewayResult
		metadata := map[string]interface{}{}
		if vsItem.GetObjectMeta().GetResourceVersion() != "" {
			metadata["resourceVersion"] = vsItem.GetObjectMeta().GetResourceVersion()
		}
		err = br.patchVirtualService(ctx, ns, vsItem.GetObjectMeta().GetName(), metadata, map[string]interface{}{
			"gateways": vsItem.Spec.GetGateways(),
		})
		if err != nil {
			return err
		}
	}

	return nil

}

func (br *IstioBridge) CreateGatewayInternal(name string, namespace string, forUpdateOnly bool, requestdata *gateway_manager.GatewayRequest) error {

	ctx, cancelFn := context.WithCancel(context.Background())

	defer cancelFn()

	currentSecretResults := make(map[string]bool)
	var currentGatewayItem *istioNetworking.Gateway
	var err error
	tlsSecretNamespace := gatewayTLSSecretNamespace()

	// 更新存在的 Gateway
	if forUpdateOnly {
		currentGatewayItem, err = br.gatewayResources.GetGateway(ctx, namespace, name, k8smetav1.GetOptions{})

		if err != nil {
			return err
		}

		if currentGatewayItem == nil {
			return fmt.Errorf("Gateway not found. name: %s, namespace: %s", name, namespace)
		}

		if requestdata.ResourceVersion != nil && *requestdata.ResourceVersion != "" && *requestdata.ResourceVersion != currentGatewayItem.GetResourceVersion() {
			recordIstioStaleConflict(istioResourceGateway, kubernetesWriteVerbUpdate)
			return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "gateways"}, name, fmt.Errorf("resource version changed"))
		}

		// 先掃描一下目前使用中的 secret key
		for _, sItem := range currentGatewayItem.Spec.GetServers() {
			if cItem := sItem.GetTls().GetCredentialName(); cItem != "" {
				currentSecretResults[cItem] = true
			}
		}
	}

	newSecretResults := make(map[string]bool)

	gateway := composeGatewayBaseData(name, namespace)

	serverItems := make([]*v1alpha3.Server, 0, len(requestdata.Servers))

	for _, serverDataItem := range requestdata.Servers {

		if len(serverDataItem.Hosts) == 0 {
			return errors.New("Server Hosts is empty")
		}

		if len(serverDataItem.Ports) == 0 {
			return errors.New("Server Ports is empty")
		}

		for _, hostItem := range serverDataItem.Hosts {
			if hostItem == "" {
				return errors.New("Hostname or domain name is empty")
			}
		}

		for _, portItem := range serverDataItem.Ports {

			// 產生基本所需要半成品
			server := v1alpha3.Server{
				Hosts: serverDataItem.Hosts,
			}

			if portItem.Port < 0 || portItem.Port > 65535 {
				return fmt.Errorf("Incorrect port number - %d", portItem.Port)
			}

			useTLS := false
			useTLSInFile := false

			portData := v1alpha3.Port{
				Name:     "Port-" + strconv.FormatInt(int64(portItem.Port), 10) + "-" + RandomString(5, ""),
				Number:   portItem.Port,
				Protocol: portItem.Protocol,
			}

			// 有指定 https or tls ?
			if isRequireTLSProtocol(portItem.Protocol) {
				useTLS = true

				// 檢查 secret key 避免重複建立
				var tlsBaseName string

				if portItem.CredentialName == "" && portItem.Cert == "" {
					return fmt.Errorf("Port %d required cert/pkey or credentialname", portItem.Port)
				}

				// 有指定 CredentialName 資料直接使用
				if portItem.CredentialName != "" {

					found, err := br.app.GetClusterBridge().SecretsExist(portItem.CredentialName, tlsSecretNamespace)

					if err != nil {
						return fmt.Errorf("Could not retrieve secret data. name: %s, namespace: %s", portItem.CredentialName, tlsSecretNamespace)

					} else if found == false {
						return fmt.Errorf("Credential Name not found, name: %s, namespace: %s", portItem.CredentialName, tlsSecretNamespace)

					} else {
						tlsBaseName = portItem.CredentialName
						newSecretResults[tlsBaseName] = true
					}

				} else if strings.HasPrefix(portItem.Cert, "/") && strings.HasPrefix(portItem.Pkey, "/") {
					// 這是用路徑指定？
					useTLSInFile = true
				} else {
					tlsMaterial, err := normalizeGatewayTLSMaterial(portItem)
					if err != nil {
						return err
					}

					for i := 0; ; i++ {
						tlsBaseName = composeGatewayTLSBaseName(name, tlsSecretNamespace, portItem.Port)
						found, _ := br.app.GetClusterBridge().SecretsExist(tlsBaseName, tlsSecretNamespace)
						if i > 100 {
							return errors.New("Could not generate unique random secret key")
						}
						if found {
							continue
						}

						break
					}
					newSecretResults[tlsBaseName] = true

					// 重建 secret key
					_ = br.app.GetClusterBridge().DeleteSecrets(tlsBaseName, tlsSecretNamespace)
					err = br.app.GetClusterBridge().CreateSecrets(tlsBaseName, tlsSecretNamespace, tlsMaterial.cert, tlsMaterial.key, tlsMaterial.ca)
					if err != nil {
						return fmt.Errorf("Could not retrieve secret key, name: %s, namespace: %s, reason: %s", tlsBaseName, tlsSecretNamespace, err.Error())
					}
				}

				// 組合採用 TLS 驗證方式
				if useTLS && !useTLSInFile {
					mode, _, err := gatewayTLSMode(portItem.Mode)
					if err != nil {
						return err
					}
					server.Tls = &v1alpha3.ServerTLSSettings{
						Mode:           mode,
						CredentialName: tlsBaseName,
					}
				} else if useTLS && useTLSInFile {
					server.Tls = &v1alpha3.ServerTLSSettings{
						PrivateKey:        portItem.Pkey,
						ServerCertificate: portItem.Cert,
					}
				}
			}

			server.Port = &portData
			serverItems = append(serverItems, &server)
		}
	}

	// 處理 Selector 項目
	selectorData := make(map[string]string)
	if currentGatewayItem != nil {
		for k, v := range currentGatewayItem.Spec.Selector {
			selectorData[k] = v
		}
	}
	for k, v := range requestdata.SelectorMatchLabels {
		selectorData[k] = v
	}

	if len(selectorData) == 0 {
		selectorData["istio"] = "ingressgateway"
	}

	// 最後資料套用
	gateway.Spec = v1alpha3.Gateway{
		Servers:  serverItems,
		Selector: selectorData,
	}

	// 更新用所以 Resource 版本重新指定
	if forUpdateOnly {

		if requestdata.ResourceVersion != nil && *requestdata.ResourceVersion != "" {
			gateway.SetResourceVersion(*requestdata.ResourceVersion)
		} else {
			gateway.SetResourceVersion(currentGatewayItem.GetResourceVersion())
		}
	}

	// 決定要建立還是更新
	if forUpdateOnly {
		metadata := map[string]interface{}{}
		if gateway.GetResourceVersion() != "" {
			metadata["resourceVersion"] = gateway.GetResourceVersion()
		}
		err = br.patchGateway(ctx, namespace, name, metadata, map[string]interface{}{
			"servers":  gateway.Spec.Servers,
			"selector": gateway.Spec.Selector,
		})
	} else {
		_, err = br.gatewayResources.CreateGateway(ctx, namespace, gateway, k8smetav1.CreateOptions{})
	}

	if err != nil {
		for k := range newSecretResults {
			if _, existed := currentSecretResults[k]; !existed && isManagedGatewaySecret(k) {
				_ = br.app.GetClusterBridge().DeleteSecrets(k, tlsSecretNamespace)
			}
		}

		return err
	}

	// 更新的話要刪除過時的 secret
	if forUpdateOnly {
		for k := range currentSecretResults {
			if _, ok := newSecretResults[k]; !ok && isManagedGatewaySecret(k) {
				_ = br.app.GetClusterBridge().DeleteSecrets(k, tlsSecretNamespace)
			}
		}
	}

	return nil

}
