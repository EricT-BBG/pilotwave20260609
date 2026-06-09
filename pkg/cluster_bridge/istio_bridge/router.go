package istio_bridge

import (
	"context"
	"fmt"
	"strings"

	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	duration_type "github.com/gogo/protobuf/types"
	istio_networking "istio.io/api/networking/v1alpha3"
	istio_clientgo_networking "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (br *IstioBridge) createDestinationRule(name string, namespace string, versionList []string) error {

	virtualServiceDestinationRule := &istio_clientgo_networking.DestinationRule{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "v1alpha3",
		},
	}

	virtualServiceDestinationRule.Name = name
	virtualServiceDestinationRule.Namespace = namespace
	virtualServiceDestinationRule.Spec.Host = name

	Subsets := make([]*istio_networking.Subset, 0)

	for _, value := range versionList {
		Subsets = append(Subsets,
			&istio_networking.Subset{
				Name:   value,
				Labels: map[string]string{"version": value},
				TrafficPolicy: &istio_networking.TrafficPolicy{LoadBalancer: &istio_networking.LoadBalancerSettings{
					LbPolicy: &istio_networking.LoadBalancerSettings_Simple{Simple: istio_networking.LoadBalancerSettings_ROUND_ROBIN},
				}}},
		)
	}

	virtualServiceDestinationRule.Spec.Subsets = Subsets
	_, err := br.routerResources.CreateDestinationRule(context.TODO(), namespace, virtualServiceDestinationRule, k8smetav1.CreateOptions{})
	return err
}

func (br *IstioBridge) updateDestinationRule(name string, namespace string, versionList []string) error {

	ruleData, err := br.routerResources.GetDestinationRule(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return br.createDestinationRule(name, namespace, versionList)
	}

	existSubsets := ruleData.Spec.GetSubsets()

	for _, item := range versionList {
		found := false
		for _, subnetValue := range existSubsets {
			value, ok := subnetValue.Labels["version"]
			if ok && value == item {
				found = true
				break
			}
		}

		if !found {
			existSubsets = append(existSubsets, &istio_networking.Subset{
				Name:   item,
				Labels: map[string]string{"version": item},
				TrafficPolicy: &istio_networking.TrafficPolicy{LoadBalancer: &istio_networking.LoadBalancerSettings{
					LbPolicy: &istio_networking.LoadBalancerSettings_Simple{Simple: istio_networking.LoadBalancerSettings_ROUND_ROBIN},
				}}},
			)
			break
		}
	}

	ruleData.Spec.Subsets = existSubsets
	return br.patchDestinationRule(context.TODO(), namespace, name, nil, map[string]interface{}{
		"subsets": ruleData.Spec.Subsets,
	})

}

func (br *IstioBridge) newVirtualService(name string, protocol string, hosts []string) *istio_clientgo_networking.VirtualService {
	virtualService := &istio_clientgo_networking.VirtualService{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "v1alpha3",
		},
	}
	virtualService.Name = name
	virtualService.Spec.Hosts = hosts
	if protocol != "" {
		labels := map[string]string{
			"protocol": protocol,
		}
		virtualService.Labels = labels
	}

	// default destination
	httpRouteDestination := make([]*istio_networking.HTTPRouteDestination, 0)
	httpRouteDestination = append(httpRouteDestination, &istio_networking.HTTPRouteDestination{
		Destination: &istio_networking.Destination{
			Host: "localhost",
		},
	})
	httpRoute := &istio_networking.HTTPRoute{
		Name:  "http" + "-" + RandomString(5, ""),
		Route: httpRouteDestination,
	}
	var router []*istio_networking.HTTPRoute
	router = append(router, httpRoute)
	virtualService.Spec.Http = router

	return virtualService
}

func (br *IstioBridge) GetRouters(page int, perPage int, search string, namespace string) ([]router_manager.RouterResponse, int, error) {

	// List virtual service
	vsList, err := br.routerResources.ListVirtualServices(context.TODO(), namespace, k8smetav1.ListOptions{})
	if err != nil {
		return []router_manager.RouterResponse{}, 0, err
	}

	items := make([]router_manager.RouterResponse, 0)
	if perPage < 0 {
		for i := range vsList.Items {
			vs := vsList.Items[i]
			hosts := make([]string, 0)
			if vs.Spec.Hosts != nil {
				hosts = vs.Spec.Hosts
			}

			destinationCount := 0
			for _, http := range vs.Spec.Http {
				if http.Route != nil {
					for _, route := range http.Route {
						if route.Destination != nil {
							if route.Destination.Host != "" {
								destinationCount = destinationCount + 1
							}
						}
					}
				}
			}

			items = append(items, router_manager.RouterResponse{
				Name:             vs.ObjectMeta.Name,
				Hosts:            hosts,
				HttpCount:        len(vs.Spec.Http),
				DestinationCount: destinationCount,
				Namespace:        vs.ObjectMeta.Namespace,
				Protocol:         vs.ObjectMeta.Labels["protocol"],
				CreatedAt:        vs.ObjectMeta.CreationTimestamp.Unix(),
				ResourceVersion:  vs.ObjectMeta.ResourceVersion,
			})
		}
	} else {
		for i := (page - 1) * perPage; i <= (page*perPage)-1; i++ {
			if i == len(vsList.Items) {
				break
			}
			vs := vsList.Items[i]
			hosts := make([]string, 0)
			if vs.Spec.Hosts != nil {
				hosts = vs.Spec.Hosts
			}

			destinationCount := 0
			for _, http := range vs.Spec.Http {
				if http.Route != nil {
					for _, route := range http.Route {
						if route.Destination != nil {
							if route.Destination.Host != "" {
								destinationCount = destinationCount + 1
							}
						}
					}
				}
			}

			items = append(items, router_manager.RouterResponse{
				Name:             vs.ObjectMeta.Name,
				Hosts:            hosts,
				HttpCount:        len(vs.Spec.Http),
				DestinationCount: destinationCount,
				Namespace:        vs.ObjectMeta.Namespace,
				Protocol:         vs.ObjectMeta.Labels["protocol"],
				CreatedAt:        vs.ObjectMeta.CreationTimestamp.Unix(),
				ResourceVersion:  vs.ObjectMeta.ResourceVersion,
			})
		}
	}

	return items, len(vsList.Items), nil
}

func (br *IstioBridge) GetRouter(name string, namespace string) (*router_manager.RouterResponse, error) {

	// Get virtual service
	vs, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return &router_manager.RouterResponse{}, err
	}

	return &router_manager.RouterResponse{
		Name: vs.ObjectMeta.Name,
		// Description: "",
		Hosts:           vs.Spec.Hosts,
		Namespace:       vs.ObjectMeta.Namespace,
		Protocol:        vs.ObjectMeta.Labels["protocol"],
		CreatedAt:       vs.ObjectMeta.CreationTimestamp.Unix(),
		ResourceVersion: vs.ObjectMeta.ResourceVersion,
	}, nil
}

func (br *IstioBridge) GetRouterServices(name string, namespace string) ([]string, error) {

	services := make([]string, 0)

	// Get virtual service
	vs, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return services, err
	}

	// services = append(services, vs.Spec.Hosts...)

	destinations := make([]string, 0)
	for _, http := range vs.Spec.Http {
		if http.Route != nil {
			for _, route := range http.Route {
				var host string
				if route.Destination != nil {
					if route.Destination.Host != "" {
						host = route.Destination.Host
					}
				}
				destinations = append(destinations, host)
			}
		}
	}
	services = append(services, destinations...)

	return services, nil
}

func (br *IstioBridge) CreateRouter(name string, namespace string, protocol string, hosts []string) error {

	ctx, cancelFn := context.WithCancel(context.Background())

	virtualService := br.newVirtualService(name, protocol, hosts)

	// Create virtual service
	_, err := br.routerResources.CreateVirtualService(ctx, namespace, virtualService, k8smetav1.CreateOptions{})
	defer cancelFn()
	if err != nil {
		return err
	}

	return nil
}

func (br *IstioBridge) UpdateRouter(name string, namespace string, protocol string, hosts []string, resourceVersion string) error {

	ctx, cancelFn := context.WithCancel(context.Background())

	// Get virtual service
	virtualService, err := br.routerResources.GetVirtualService(ctx, namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return err
	}
	if resourceVersion != "" && resourceVersion != virtualService.GetResourceVersion() {
		recordIstioStaleConflict(istioResourceVirtualService, kubernetesWriteVerbUpdate)
		return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "virtualservices"}, name, fmt.Errorf("resource version changed"))
	}

	virtualService.Name = name
	virtualService.Spec.Hosts = hosts
	metadata := map[string]interface{}{}
	if protocol != "" {
		metadata["labels"] = map[string]string{"protocol": protocol}
	}
	if virtualService.GetResourceVersion() != "" {
		metadata["resourceVersion"] = virtualService.GetResourceVersion()
	}
	if resourceVersion != "" {
		metadata["resourceVersion"] = resourceVersion
	}

	// Update virtual service
	err = br.patchVirtualService(ctx, namespace, name, metadata, map[string]interface{}{
		"hosts": virtualService.Spec.Hosts,
	})
	defer cancelFn()
	if err != nil {
		return err
	}

	return nil
}

func (br *IstioBridge) DeleteRouter(name string, namespace string) error {
	ctx, cancelFn := context.WithCancel(context.Background())

	// Delete virtual service
	err := br.routerResources.DeleteVirtualService(ctx, namespace, name, k8smetav1.DeleteOptions{})
	defer cancelFn()
	if err != nil {
		return err
	}

	return nil
}

func (br *IstioBridge) UpdateRouterRule(name string, namespace string, rule router_manager.RouterRuleRequest) error {

	// Get virtual service
	virtualService, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return err
	}
	if rule.ResourceVersion != "" && rule.ResourceVersion != virtualService.GetResourceVersion() {
		recordIstioStaleConflict(istioResourceVirtualService, kubernetesWriteVerbUpdate)
		return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "virtualservices"}, name, fmt.Errorf("resource version changed"))
	}

	virtualService.Name = name
	if rule.ResourceVersion != "" {
		virtualService.SetResourceVersion(rule.ResourceVersion)
	}
	var router []*istio_networking.HTTPRoute
	for _, routerRule := range rule.Https {
		routerMatch := make([]*istio_networking.HTTPMatchRequest, 0)

		if len(routerRule.Prefixs) != 0 {

			prefixsCheck := make(map[string]bool)

			for _, p := range routerRule.Prefixs {

				if p != "" {
					if prefixsCheck[p] == true {
						return fmt.Errorf("duplicate prefix: %s", p)
					} else {
						prefixsCheck[p] = true
					}
				}

				routerMatch = append(routerMatch, &istio_networking.HTTPMatchRequest{
					Uri: &istio_networking.StringMatch{
						MatchType: &istio_networking.StringMatch_Prefix{
							Prefix: p,
						},
					},
				})
			}
		}

		routerRewrite := &istio_networking.HTTPRewrite{}
		if routerRule.Rewrite != "" {
			routerRewrite = &istio_networking.HTTPRewrite{
				Uri: routerRule.Rewrite,
			}
		}

		timeout := &duration_type.Duration{}
		if routerRule.Timeout != 0 {
			timeout = &duration_type.Duration{
				Seconds: routerRule.Timeout,
			}
		}

		fixedDelay := &istio_networking.HTTPFaultInjection{}
		if routerRule.FixedDelay != 0 {
			fixedDelay = &istio_networking.HTTPFaultInjection{
				Delay: &istio_networking.HTTPFaultInjection_Delay{
					HttpDelayType: &istio_networking.HTTPFaultInjection_Delay_FixedDelay{
						FixedDelay: &duration_type.Duration{
							Seconds: routerRule.FixedDelay,
						},
					},
				},
			}
		}

		headers := &istio_networking.Headers{}
		if len(routerRule.Headers) != 0 {
			headerData := make(map[string]string)
			for _, header := range routerRule.Headers {
				headerData[header.Key] = header.Value
			}
			headers = &istio_networking.Headers{
				Request: &istio_networking.Headers_HeaderOperations{
					Set: headerData,
				},
			}
		}

		httpRouteDestination := make([]*istio_networking.HTTPRouteDestination, 0)

		subsetResult := make([]string, 0)

		for _, destination := range routerRule.Destinations {
			des := &istio_networking.HTTPRouteDestination{
				Destination: &istio_networking.Destination{
					Host:   destination.Host,
					Subset: destination.Subset,
				},
			}

			if destination.Subset != "" {
				subsetResult = append(subsetResult, destination.Subset)
			}

			if destination.Port != 0 {
				des.Destination.Port = &istio_networking.PortSelector{Number: destination.Port}
			}

			if destination.Weight != 0 {
				des.Weight = destination.Weight
			}
			httpRouteDestination = append(httpRouteDestination, des)
		}

		if len(subsetResult) > 0 {
			if err = br.updateDestinationRule(name, namespace, subsetResult); err != nil {
				return err
			}
		}

		httpRoute := &istio_networking.HTTPRoute{
			Name:  "http" + "-" + RandomString(5, ""),
			Route: httpRouteDestination,
		}
		if len(routerRule.Prefixs) != 0 {
			httpRoute.Match = routerMatch
		}
		if routerRule.Rewrite != "" {
			httpRoute.Rewrite = routerRewrite
		}
		if routerRule.Timeout != 0 {
			httpRoute.Timeout = timeout
		}
		if routerRule.FixedDelay != 0 {
			httpRoute.Fault = fixedDelay
		}
		if len(routerRule.Headers) != 0 {
			httpRoute.Headers = headers
		}

		router = append(router, httpRoute)
	}

	virtualService.Spec.Http = router

	// Update virtual service
	metadata := map[string]interface{}{}
	if virtualService.GetResourceVersion() != "" {
		metadata["resourceVersion"] = virtualService.GetResourceVersion()
	}
	if rule.ResourceVersion != "" {
		metadata["resourceVersion"] = rule.ResourceVersion
	}
	err = br.patchVirtualService(context.TODO(), namespace, name, metadata, map[string]interface{}{
		"http": virtualService.Spec.Http,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *IstioBridge) GetRouterRule(name string, namespace string) (router_manager.RouterRuleResponse, error) {
	// Get virtual service
	virtualService, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return router_manager.RouterRuleResponse{}, err
	}

	https := make([]router_manager.HttpsData, 0)

	for _, http := range virtualService.Spec.Http {
		prefixs := make([]string, 0)

		for _, match := range http.Match {
			if match.GetUri().GetPrefix() != "" {
				prefixs = append(prefixs, match.GetUri().GetPrefix())
			}

			if match.GetUri().GetExact() != "" {
				prefixs = append(prefixs, match.GetUri().GetExact())
			}
		}

		headers := make([]router_manager.HeaderData, 0)
		if http.Headers != nil {
			if http.Headers.Request != nil {
				for k, v := range http.Headers.Request.GetSet() {
					headers = append(headers, router_manager.HeaderData{
						Key:   k,
						Value: v,
					})
				}
			}
		}

		destinations := make([]router_manager.DestinationData, 0)

		if http.Route != nil {
			for _, route := range http.Route {
				var host string
				var port uint32
				var subset string
				if route.Destination != nil {
					if route.Destination.Host != "" {
						host = route.Destination.Host
					}
					if route.Destination.Port != nil {
						port = route.Destination.Port.Number
					}
					if route.Destination.Subset != "" {
						subset = route.Destination.Subset
					}
				}

				// Port 不一定要指定 所以判斷處理

				destdata := router_manager.DestinationData{
					Weight: route.Weight,
					Host:   host,
					Subset: subset,
				}

				if port != 0 {
					destdata.Port = port
				}

				destinations = append(destinations, destdata)
			}
		}

		var rewrite string
		if http.Rewrite != nil {
			rewrite = http.Rewrite.Uri
		}

		var fixedDelay int64
		if http.Fault != nil {
			if http.Fault.Delay != nil {
				fixedDelay = http.Fault.Delay.GetFixedDelay().Seconds
			}
		}

		var timeout int64
		if http.Timeout != nil {
			timeout = http.Timeout.Seconds
		}

		https = append(https, router_manager.HttpsData{
			Prefixs:      prefixs,
			Headers:      headers,
			Destinations: destinations,
			Rewrite:      rewrite,
			FixedDelay:   fixedDelay,
			Timeout:      timeout,
		})
	}
	return router_manager.RouterRuleResponse{
		Name:            name,
		Namespace:       namespace,
		Https:           https,
		ResourceVersion: virtualService.ObjectMeta.ResourceVersion,
	}, nil
}

func (br *IstioBridge) GetRouterGatewayMapping(name string, namespace string) (router_manager.RouterMappingResponse, error) {

	vs, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return router_manager.RouterMappingResponse{
			Name:      name,
			Namespace: namespace,
		}, err
	}

	gatewayItems := make([]router_manager.RouterMappingGatewayData, 0, 10)

	for _, item := range vs.Spec.GetGateways() {
		var gatewayNamespace string
		var gatewayName string

		spiltResult := strings.Split(item, "/")

		if len(spiltResult) == 2 {
			gatewayNamespace = spiltResult[0]
			gatewayName = spiltResult[1]
		} else {
			gatewayNamespace = vs.GetObjectMeta().GetNamespace()
			gatewayName = item
		}

		gatewayItems = append(gatewayItems,
			router_manager.RouterMappingGatewayData{
				Name:      gatewayName,
				Namespace: gatewayNamespace,
			})
	}

	return router_manager.RouterMappingResponse{
		Name:            name,
		Namespace:       namespace,
		Gateways:        gatewayItems,
		ResourceVersion: vs.ObjectMeta.ResourceVersion,
	}, nil
}

func (br *IstioBridge) CreateRouterGatewayMapping(name string, namespace string, gateways []router_manager.RouterMappingGatewayData, resourceVersion string) error {

	ctx, cancelFn := context.WithCancel(context.Background())

	defer cancelFn()

	// 取得目前的 vs
	vs, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return err
	}
	if resourceVersion != "" && resourceVersion != vs.GetResourceVersion() {
		recordIstioStaleConflict(istioResourceVirtualService, kubernetesWriteVerbUpdate)
		return k8serrors.NewConflict(schema.GroupResource{Group: "networking.istio.io", Resource: "virtualservices"}, name, fmt.Errorf("resource version changed"))
	}

	// 先掃描一次確認傳入的 gateway 是否都存在再繼續
	for _, item := range gateways {
		if item.Namespace == "" {
			item.Namespace = namespace
		}
		gwItem, err := br.gatewayResources.GetGateway(ctx, item.Namespace, item.Name, k8smetav1.GetOptions{})
		if err != nil || gwItem == nil {
			errmsg := fmt.Errorf("Gateway (namespace: %s, name: %s) not found for router mapping", item.Namespace, item.Name)
			return errmsg
		}
	}

	// 重新產生結構
	gatewaylist := make([]string, 0, len(gateways))
	for _, item := range gateways {
		var composedName string
		if namespace == item.Namespace {
			composedName = item.Name
		} else {
			composedName = item.Namespace + "/" + item.Name
		}
		gatewaylist = append(gatewaylist, composedName)
	}

	vs.Spec.Gateways = gatewaylist
	metadata := map[string]interface{}{}
	if vs.GetResourceVersion() != "" {
		metadata["resourceVersion"] = vs.GetResourceVersion()
	}
	if resourceVersion != "" {
		metadata["resourceVersion"] = resourceVersion
	}
	err = br.patchVirtualService(ctx, namespace, name, metadata, map[string]interface{}{
		"gateways": vs.Spec.Gateways,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *IstioBridge) UpdateRouterGatewayMapping(name string, namespace string, gateways []router_manager.RouterMappingGatewayData, resourceVersion string) error {

	return br.CreateRouterGatewayMapping(name, namespace, gateways, resourceVersion)
}

func (br *IstioBridge) DeleteRouterGatewayMapping(name string, namespace string) error {

	ctx, cancelFn := context.WithCancel(context.Background())

	defer cancelFn()

	// Get virtual service
	vs, err := br.routerResources.GetVirtualService(context.TODO(), namespace, name, k8smetav1.GetOptions{})
	if err != nil {
		return err
	}

	vs.Spec.Gateways = []string{}
	metadata := map[string]interface{}{}
	if vs.GetResourceVersion() != "" {
		metadata["resourceVersion"] = vs.GetResourceVersion()
	}
	err = br.patchVirtualService(ctx, namespace, name, metadata, map[string]interface{}{
		"gateways": vs.Spec.Gateways,
	})
	if err != nil {
		return err
	}

	return nil
}
