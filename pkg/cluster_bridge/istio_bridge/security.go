package istio_bridge

import (
	"context"
	"errors"
	"fmt"
	"git.brobridge.com/pilotwave/pilotwave/pkg/security_manager"
	istioServerSecurityv1beta1 "istio.io/api/security/v1beta1"
	istioV1beta1 "istio.io/api/type/v1beta1"
	istioClientSecurityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

// Common function
func composeAuthorizationPolicyBaseData(name string, namespace string) *istioClientSecurityv1beta1.AuthorizationPolicy {

	return &istioClientSecurityv1beta1.AuthorizationPolicy{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "AuthorizationPolicy",
			APIVersion: "security.istio.io/v1beta1",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func composeRequestAuthenticationBaseData(name string, namespace string) *istioClientSecurityv1beta1.RequestAuthentication {

	return &istioClientSecurityv1beta1.RequestAuthentication{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "RequestAuthentication",
			APIVersion: "security.istio.io/v1beta1",
		},

		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// Authentication Policy

func (br *IstioBridge) GetAuthorizationPoliciesFull(namespace string, search string, searchIsFuzzy bool) ([]*security_manager.AuthorizationPolicyResponse, int, error) {

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	opts := k8smetav1.ListOptions{}

	if !searchIsFuzzy && (len(search) > 0) {
		opts.FieldSelector = "metadata.name=" + search
	}

	requestPolicyList, err := br.securityResources.ListAuthorizationPolicies(ctx, namespace, opts)

	if err != nil {
		return []*security_manager.AuthorizationPolicyResponse{}, 0, err
	}

	if requestPolicyList.Items == nil {
		return []*security_manager.AuthorizationPolicyResponse{}, 0, nil
	}

	responseData := make([]*security_manager.AuthorizationPolicyResponse, 0, len(requestPolicyList.Items))

	for _, requestPolicyItem := range requestPolicyList.Items {

		// Fuzzy search
		if search != "" {
			if searchIsFuzzy && !strings.Contains(requestPolicyItem.Name, search) {
				continue
			}
		}

		// Base
		eachData := security_manager.AuthorizationPolicyResponse{
			Name:            requestPolicyItem.GetName(),
			Namespace:       requestPolicyItem.GetNamespace(),
			ResourceVersion: requestPolicyItem.GetResourceVersion(),
			Action:          requestPolicyItem.Spec.GetAction().String(),
			CreatedAt:       requestPolicyItem.GetCreationTimestamp().Unix(),
		}

		// MatchLabels Item
		labels := make([]security_manager.LabelDtata, 0)
		if requestPolicyItem.Spec.Selector != nil {
			for k, v := range requestPolicyItem.Spec.Selector.GetMatchLabels() {
				labels = append(labels, security_manager.LabelDtata{
					Key:   k,
					Value: v,
				})
			}
		}
		eachData.SelectorMatchLabels = labels

		ruleResults := make([]security_manager.AuthorizationPolicyRuleData, 0, len(requestPolicyItem.Spec.GetRules()))

		if rulesItems := requestPolicyItem.Spec.GetRules(); rulesItems != nil {

			for _, eachRuleItem := range rulesItems {

				fromResult := make([]*security_manager.AuthorizationPolicyFromRuleData, 0, len(eachRuleItem.GetFrom()))
				toResult := make([]*security_manager.AuthorizationPolicyToOperationData, 0, len(eachRuleItem.GetTo()))
				whenResult := make([]*security_manager.RulesWhenData, 0, len(eachRuleItem.GetWhen()))

				if ruleFromItems := eachRuleItem.GetFrom(); ruleFromItems != nil {
					for _, eachFromItem := range ruleFromItems {
						if sourceItem := eachFromItem.GetSource(); sourceItem != nil {
							IpBlocks := make([]string, 0)
							if sourceItem.GetIpBlocks() != nil {
								IpBlocks = sourceItem.GetIpBlocks()
							}
							NotIpBlocks := make([]string, 0)
							if sourceItem.GetIpBlocks() != nil {
								NotIpBlocks = sourceItem.GetNotIpBlocks()
							}
							RemoteIpBlocks := make([]string, 0)
							if sourceItem.GetRemoteIpBlocks() != nil {
								RemoteIpBlocks = sourceItem.GetRemoteIpBlocks()
							}
							NotRemoteIpBlocks := make([]string, 0)
							if sourceItem.GetNotRemoteIpBlocks() != nil {
								NotRemoteIpBlocks = sourceItem.GetNotRemoteIpBlocks()
							}
							Principals := make([]string, 0)
							if sourceItem.GetPrincipals() != nil {
								Principals = sourceItem.GetPrincipals()
							}
							NotPrincipals := make([]string, 0)
							if sourceItem.GetNotPrincipals() != nil {
								NotPrincipals = sourceItem.GetNotPrincipals()
							}
							RequestPrincipals := make([]string, 0)
							if sourceItem.GetRequestPrincipals() != nil {
								RequestPrincipals = sourceItem.GetRequestPrincipals()
							}
							NotRequestPrincipals := make([]string, 0)
							if sourceItem.GetNotRequestPrincipals() != nil {
								NotRequestPrincipals = sourceItem.GetNotRequestPrincipals()
							}
							Namespaces := make([]string, 0)
							if sourceItem.GetNamespaces() != nil {
								Namespaces = sourceItem.GetNamespaces()
							}
							NotNamespaces := make([]string, 0)
							if sourceItem.GetNotNamespaces() != nil {
								NotNamespaces = sourceItem.GetNotNamespaces()
							}

							sourceResult := &security_manager.AuthorizationPolicyRuleSourceData{
								IpBlocks:             IpBlocks,
								NotIpBlocks:          NotIpBlocks,
								RemoteIpBlocks:       RemoteIpBlocks,
								NotRemoteIpBlocks:    NotRemoteIpBlocks,
								Principals:           Principals,
								NotPrincipals:        NotPrincipals,
								RequestPrincipals:    RequestPrincipals,
								NotRequestPrincipals: NotRequestPrincipals,
								Namespaces:           Namespaces,
								NotNamespaces:        NotNamespaces}

							fromResult = append(fromResult, &security_manager.AuthorizationPolicyFromRuleData{
								Source: sourceResult,
							})
						}
					}
				}

				if ruleToItems := eachRuleItem.GetTo(); ruleToItems != nil {
					for _, eachToItem := range ruleToItems {
						if toItem := eachToItem.GetOperation(); toItem != nil {
							Paths := make([]string, 0)
							if toItem.GetPaths() != nil {
								Paths = toItem.GetPaths()
							}
							NotPaths := make([]string, 0)
							if toItem.GetNotPaths() != nil {
								NotPaths = toItem.GetNotPaths()
							}
							Ports := make([]string, 0)
							if toItem.GetPorts() != nil {
								Ports = toItem.GetPorts()
							}
							NotPorts := make([]string, 0)
							if toItem.GetNotPorts() != nil {
								NotPorts = toItem.GetNotPorts()
							}
							Hosts := make([]string, 0)
							if toItem.GetHosts() != nil {
								Hosts = toItem.GetHosts()
							}
							NotHosts := make([]string, 0)
							if toItem.GetNotHosts() != nil {
								NotHosts = toItem.GetNotHosts()
							}
							Methods := make([]string, 0)
							if toItem.GetMethods() != nil {
								Methods = toItem.GetMethods()
							}
							NotMethods := make([]string, 0)
							if toItem.GetNotMethods() != nil {
								Hosts = toItem.GetNotMethods()
							}
							destResult := &security_manager.AuthorizationPolicyRuleDestinationData{
								Paths:      Paths,
								NotPaths:   NotPaths,
								Ports:      Ports,
								NotPorts:   NotPorts,
								Hosts:      Hosts,
								NotHosts:   NotHosts,
								Methods:    Methods,
								NotMethods: NotMethods}

							toResult = append(toResult, &security_manager.AuthorizationPolicyToOperationData{
								Operation: destResult,
							})
						}
					}
				}

				if ruleWhenItems := eachRuleItem.GetWhen(); ruleWhenItems != nil {
					for _, eachWhenItem := range ruleWhenItems {
						if eachWhenItem.GetKey() != "" {
							Values := make([]string, 0)
							if eachWhenItem.GetValues() != nil {
								Values = eachWhenItem.GetValues()
							}
							NotValues := make([]string, 0)
							if eachWhenItem.GetNotValues() != nil {
								NotValues = eachWhenItem.GetNotValues()
							}
							condResult := &security_manager.RulesWhenData{
								Key:       eachWhenItem.GetKey(),
								Values:    Values,
								NotValues: NotValues,
							}
							whenResult = append(whenResult, condResult)
						}
					}
				}

				v := security_manager.AuthorizationPolicyRuleData{
					From: fromResult,
					To:   toResult,
					When: whenResult}
				ruleResults = append(ruleResults, v)
			}

		}

		eachData.Rules = ruleResults
		responseData = append(responseData, &eachData)
	}

	return responseData, len(responseData), nil
}

func (br *IstioBridge) GetAuthorizationPolicies(page int, perPage int, search string, namespace string) ([]*security_manager.AuthorizationPolicyResponse, int, error) {

	data, total, err := br.GetAuthorizationPoliciesFull(namespace, search, true)
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

func composeCreateAuthorizationPolicyData(name string, namespace string, data *security_manager.AuthorizationPolicyRequest) (*istioClientSecurityv1beta1.AuthorizationPolicy, error) {

	// 建立基本結構
	composeData := composeAuthorizationPolicyBaseData(name, namespace)

	if data == nil {
		return composeData, nil
	}

	// 先處理 Action
	currentAction := strings.ToLower(data.Action)
	if currentAction == "" {
		return nil, fmt.Errorf("AuthorizationPolicy error, required action parameter for name: %s, namespace: %s", name, namespace)
	}

	// 判斷 Action 是否正確
	validActionTable := map[string]istioServerSecurityv1beta1.AuthorizationPolicy_Action{
		"allow": istioServerSecurityv1beta1.AuthorizationPolicy_ALLOW,
		"deny":  istioServerSecurityv1beta1.AuthorizationPolicy_DENY,
		"audit": istioServerSecurityv1beta1.AuthorizationPolicy_AUDIT,
	}

	newAction, ok := validActionTable[currentAction]
	if !ok {
		return nil, fmt.Errorf("unknown action parameter: %s", data.Action)
	} else {
		composeData.Spec.Action = newAction
	}

	// MatchLabels
	if data.SelectorMatchLabels != nil {
		labelData := make(map[string]string)
		if len(data.SelectorMatchLabels) != 0 {
			for _, label := range data.SelectorMatchLabels {
				labelData[label.Key] = label.Value
			}
			composeData.Spec.Selector = &istioV1beta1.WorkloadSelector{
				MatchLabels: labelData,
			}
		}
	}

	// 先產生 Rule 項目
	composeDataRuleResult := make([]*istioServerSecurityv1beta1.Rule, 0, len(data.Rules))

	for _, ruleItem := range data.Rules {
		ruleFromResults := make([]*istioServerSecurityv1beta1.Rule_From, 0, len(ruleItem.From))
		ruleToResults := make([]*istioServerSecurityv1beta1.Rule_To, 0, len(ruleItem.To))
		ruleWhenResults := make([]*istioServerSecurityv1beta1.Condition, 0, len(ruleItem.When))

		// 處理 From
		if len(ruleItem.From) != 0 {
			for _, fromItem := range ruleItem.From {
				ruleFromResult := istioServerSecurityv1beta1.Rule_From{}

				if fromItem != nil {
					if fromItem.Source != nil {
						ruleFromResult.Source = &istioServerSecurityv1beta1.Source{}

						if len(fromItem.Source.Principals) != 0 {
							ruleFromResult.Source.Principals = fromItem.Source.Principals
						}
						if len(fromItem.Source.NotPrincipals) != 0 {
							ruleFromResult.Source.NotPrincipals = fromItem.Source.NotPrincipals
						}
						if len(fromItem.Source.RequestPrincipals) != 0 {
							ruleFromResult.Source.RequestPrincipals = fromItem.Source.RequestPrincipals
						}
						if len(fromItem.Source.NotRequestPrincipals) != 0 {
							ruleFromResult.Source.NotRequestPrincipals = fromItem.Source.NotRequestPrincipals
						}
						if len(fromItem.Source.Namespaces) != 0 {
							ruleFromResult.Source.Namespaces = fromItem.Source.Namespaces
						}
						if len(fromItem.Source.NotNamespaces) != 0 {
							ruleFromResult.Source.NotNamespaces = fromItem.Source.NotNamespaces
						}
						if len(fromItem.Source.IpBlocks) != 0 {
							ruleFromResult.Source.IpBlocks = fromItem.Source.IpBlocks
						}
						if len(fromItem.Source.NotIpBlocks) != 0 {
							ruleFromResult.Source.NotIpBlocks = fromItem.Source.NotIpBlocks
						}
						if len(fromItem.Source.RemoteIpBlocks) != 0 {
							ruleFromResult.Source.RemoteIpBlocks = fromItem.Source.RemoteIpBlocks
						}
						if len(fromItem.Source.NotRemoteIpBlocks) != 0 {
							ruleFromResult.Source.NotRemoteIpBlocks = fromItem.Source.NotRemoteIpBlocks
						}
					}

					ruleFromResults = append(ruleFromResults, &ruleFromResult)
				}
			}
		}

		// 處理 To
		if len(ruleItem.To) != 0 {
			for _, toItem := range ruleItem.To {
				ruleToResult := istioServerSecurityv1beta1.Rule_To{}

				if toItem.Operation != nil {
					ruleToResult.Operation = &istioServerSecurityv1beta1.Operation{}

					if len(toItem.Operation.Hosts) != 0 {
						ruleToResult.Operation.Hosts = toItem.Operation.Hosts
					}
					if len(toItem.Operation.NotHosts) != 0 {
						ruleToResult.Operation.NotHosts = toItem.Operation.NotHosts
					}
					if len(toItem.Operation.Ports) != 0 {
						ruleToResult.Operation.Ports = toItem.Operation.Ports
					}
					if len(toItem.Operation.NotPorts) != 0 {
						ruleToResult.Operation.NotPorts = toItem.Operation.NotPorts
					}
					if len(toItem.Operation.Methods) != 0 {
						ruleToResult.Operation.Methods = toItem.Operation.Methods
					}
					if len(toItem.Operation.NotMethods) != 0 {
						ruleToResult.Operation.NotMethods = toItem.Operation.NotMethods
					}
					if len(toItem.Operation.Paths) != 0 {
						ruleToResult.Operation.Paths = toItem.Operation.Paths
					}
					if len(toItem.Operation.NotPaths) != 0 {
						ruleToResult.Operation.NotPaths = toItem.Operation.NotPaths
					}

					ruleToResults = append(ruleToResults, &ruleToResult)
				}
			}
		}

		// 處理 When
		if len(ruleItem.When) != 0 {
			for _, whenItem := range ruleItem.When {
				ruleWhenResult := istioServerSecurityv1beta1.Condition{}

				modified := false

				if whenItem.Key != "" {
					ruleWhenResult.Key = whenItem.Key
					modified = true
				}
				if len(whenItem.Values) != 0 {
					ruleWhenResult.Values = whenItem.Values
					modified = true
				}
				if len(whenItem.NotValues) != 0 {
					ruleWhenResult.NotValues = whenItem.NotValues
					modified = true
				}

				if modified {
					ruleWhenResults = append(ruleWhenResults, &ruleWhenResult)
				}
			}
		}

		composeDataRuleData := &istioServerSecurityv1beta1.Rule{}

		if len(ruleFromResults) != 0 {
			composeDataRuleData.From = ruleFromResults
		}
		if len(ruleToResults) != 0 {
			composeDataRuleData.To = ruleToResults
		}
		if len(ruleWhenResults) != 0 {
			composeDataRuleData.When = ruleWhenResults
		}

		// 合併
		composeDataRuleResult = append(composeDataRuleResult, composeDataRuleData)
	}

	composeData.Spec.Rules = composeDataRuleResult
	return composeData, nil

}

func (br *IstioBridge) GetAuthorizationPolicy(name string, namespace string) (*security_manager.AuthorizationPolicyResponse, error) {

	res, total, err := br.GetAuthorizationPoliciesFull(namespace, name, false)

	if err != nil || total == 0 {
		return nil, err
	}

	return res[0], nil

}

func (br *IstioBridge) CreateAuthorizationPolicy(name string, namespace string, data *security_manager.AuthorizationPolicyRequest) error {

	// Cancel Handler
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Compose completed data
	resp, err := composeCreateAuthorizationPolicyData(name, namespace, data)
	if err != nil {
		return err
	}

	// Create
	_, err = br.securityResources.CreateAuthorizationPolicy(ctx, namespace, resp, k8smetav1.CreateOptions{})
	return err
}

func (br *IstioBridge) UpdateAuthorizationPolicy(name string, namespace string, data *security_manager.AuthorizationPolicyUpdateRequest) error {

	// Cancel Handler
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	existInstance, errForGet := br.securityResources.GetAuthorizationPolicy(ctx, namespace, name, k8smetav1.GetOptions{})

	if errForGet != nil {
		return fmt.Errorf("Could not retrieve authentication policy data, name: %s, namespace: %s, reason: %s .", name, namespace, errForGet.Error())
	}

	if existInstance == nil {
		return errors.New("Update failure, authentication policy not found")
	}

	dataItem := &security_manager.AuthorizationPolicyRequest{
		Name:                name,
		Namespace:           namespace,
		ResourceVersion:     data.ResourceVersion,
		Rules:               data.Rules,
		SelectorMatchLabels: data.SelectorMatchLabels,
		Action:              data.Action,
	}

	// Compose completed data
	composeData, err := composeCreateAuthorizationPolicyData(name, namespace, dataItem)
	if err != nil {
		return err
	}

	// Check version
	if data.ResourceVersion != nil && *data.ResourceVersion != "" {
		if *data.ResourceVersion != existInstance.GetResourceVersion() {
			recordIstioStaleConflict(istioResourceAuthorizationPolicy, kubernetesWriteVerbUpdate)
			return k8serrors.NewConflict(schema.GroupResource{Group: "security.istio.io", Resource: "authorizationpolicies"}, name, fmt.Errorf("resource version changed"))
		}
	}

	metadata := map[string]interface{}{}
	if existInstance.GetResourceVersion() != "" {
		metadata["resourceVersion"] = existInstance.GetResourceVersion()
	}

	return br.patchAuthorizationPolicy(ctx, namespace, name, metadata, map[string]interface{}{
		"selector": composeData.Spec.Selector,
		"rules":    composeData.Spec.Rules,
		"action":   strings.ToUpper(strings.TrimSpace(data.Action)),
	})
}

func (br *IstioBridge) DeleteAuthorizationPolicy(name string, namespace string) error {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	return br.securityResources.DeleteAuthorizationPolicy(ctx, namespace, name, k8smetav1.DeleteOptions{})
}

// Request Authentication

func (br *IstioBridge) GetRequestAuthenticationsFull(namespace string, search string, searchIsFuzzy bool) ([]*security_manager.RequestAuthenticationResponse, int, error) {

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	opts := k8smetav1.ListOptions{}

	if !searchIsFuzzy && len(search) > 0 {
		opts.FieldSelector = "metadata.name=" + search
	}

	requestAuthList, err := br.securityResources.ListRequestAuthentications(ctx, namespace, opts)

	if err != nil {
		return []*security_manager.RequestAuthenticationResponse{}, 0, err
	}

	if requestAuthList.Items == nil {
		return []*security_manager.RequestAuthenticationResponse{}, 0, nil
	}

	responseData := make([]*security_manager.RequestAuthenticationResponse, 0, len(requestAuthList.Items))

	for _, requestAuthItem := range requestAuthList.Items {

		// Fuzzy search
		if search != "" {
			if searchIsFuzzy && !strings.Contains(requestAuthItem.Name, search) {
				continue
			}
		}

		// Base
		eachData := security_manager.RequestAuthenticationResponse{
			Name:            requestAuthItem.GetName(),
			Namespace:       requestAuthItem.GetNamespace(),
			ResourceVersion: requestAuthItem.GetResourceVersion(),
			CreatedAt:       requestAuthItem.GetCreationTimestamp().Unix(),
		}

		// MatchLabels Item
		labels := make([]security_manager.LabelDtata, 0)
		if requestAuthItem.Spec.Selector != nil {
			for k, v := range requestAuthItem.Spec.Selector.GetMatchLabels() {
				labels = append(labels, security_manager.LabelDtata{
					Key:   k,
					Value: v,
				})
			}

			if len(labels) != 0 {
				eachData.SelectorMatchLabels = labels
			}
		}

		// JWTRules Item
		jwtRulesResult := make([]security_manager.RequestAuthenticationJWTData, 0, len(requestAuthItem.Spec.GetJwtRules()))

		for _, jwtRulesItem := range requestAuthItem.Spec.GetJwtRules() {
			jwtRulesResult = append(jwtRulesResult, security_manager.RequestAuthenticationJWTData{
				JwksUri:   jwtRulesItem.GetJwksUri(),
				Issuer:    jwtRulesItem.GetIssuer(),
				Audiences: jwtRulesItem.GetAudiences()})
		}
		eachData.JWTRules = jwtRulesResult
		responseData = append(responseData, &eachData)
	}

	return responseData, len(responseData), nil

}

func (br *IstioBridge) GetRequestAuthentications(page int, perPage int, search string, namespace string) ([]*security_manager.RequestAuthenticationResponse, int, error) {

	data, total, err := br.GetRequestAuthenticationsFull(namespace, search, true)

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

func (br *IstioBridge) GetRequestAuthentication(name string, namespace string) (*security_manager.RequestAuthenticationResponse, error) {

	res, total, err := br.GetRequestAuthenticationsFull(namespace, name, false)

	if err != nil {
		return nil, err
	} else if total == 1 {
		return res[0], nil
	} else if total > 1 {
		return nil, fmt.Errorf("too many matched records for namespace: %s, name: %s", namespace, name)
	}
	return nil, nil
}

func composeCreateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationRequest) (*istioClientSecurityv1beta1.RequestAuthentication, error) {

	// 建立基本結構
	composeData := composeRequestAuthenticationBaseData(name, namespace)

	if data == nil {
		return composeData, nil
	}

	// MatchLabels
	if data.SelectorMatchLabels != nil {
		labelData := make(map[string]string)
		if len(data.SelectorMatchLabels) != 0 {
			for _, label := range data.SelectorMatchLabels {
				labelData[label.Key] = label.Value
			}
			composeData.Spec.Selector = &istioV1beta1.WorkloadSelector{
				MatchLabels: labelData,
			}
		}
	}

	// JWTRule
	jwtResults := make([]*istioServerSecurityv1beta1.JWTRule, 0, len(data.JWTRules))
	for _, jwtRuleItem := range data.JWTRules {

		jwtRuleItemData := &istioServerSecurityv1beta1.JWTRule{}
		modified := false

		if len(jwtRuleItem.Issuer) != 0 {
			jwtRuleItemData.Issuer = jwtRuleItem.Issuer
			modified = true
		}

		if len(jwtRuleItem.Audiences) != 0 {
			jwtRuleItemData.Audiences = jwtRuleItem.Audiences
			modified = true
		}

		if len(jwtRuleItem.JwksUri) != 0 {
			jwtRuleItemData.JwksUri = jwtRuleItem.JwksUri
			modified = true
		}

		if modified {
			jwtResults = append(jwtResults, jwtRuleItemData)
		}
	}

	composeData.Spec.JwtRules = jwtResults
	return composeData, nil
}

func (br *IstioBridge) CreateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationRequest) error {

	// Cancel Handler
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	// Compose completed data
	resp, err := composeCreateRequestAuthentication(name, namespace, data)
	if err != nil {
		return err
	}

	// Create
	_, err = br.securityResources.CreateRequestAuthentication(ctx, namespace, resp, k8smetav1.CreateOptions{})

	return err
}

func (br *IstioBridge) UpdateRequestAuthentication(name string, namespace string, data *security_manager.RequestAuthenticationUpdateRequest) error {

	// Cancel Handler
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	existInstance, err := br.securityResources.GetRequestAuthentication(ctx, namespace, name, k8smetav1.GetOptions{})

	if err != nil {
		return fmt.Errorf("Could not retrieve request authentication, name: %s, namespace: %s .", name, namespace)
	}

	if existInstance == nil {
		return fmt.Errorf("Specified request authentication not found for update, name: %s, namespace: %s .", name, namespace)
	}

	dataItem := &security_manager.RequestAuthenticationRequest{
		Name:                name,
		Namespace:           namespace,
		ResourceVersion:     data.ResourceVersion,
		JWTRules:            data.JWTRules,
		SelectorMatchLabels: data.SelectorMatchLabels,
	}

	// Compose completed data
	composedData, err := composeCreateRequestAuthentication(name, namespace, dataItem)
	if err != nil {
		return err
	}

	// Check version
	if data.ResourceVersion != nil && *data.ResourceVersion != "" {
		if *data.ResourceVersion != existInstance.GetResourceVersion() {
			recordIstioStaleConflict(istioResourceRequestAuthentication, kubernetesWriteVerbUpdate)
			return k8serrors.NewConflict(schema.GroupResource{Group: "security.istio.io", Resource: "requestauthentications"}, name, fmt.Errorf("resource version changed"))
		}
	}

	metadata := map[string]interface{}{}
	if existInstance.GetResourceVersion() != "" {
		metadata["resourceVersion"] = existInstance.GetResourceVersion()
	}

	return br.patchRequestAuthentication(ctx, namespace, name, metadata, map[string]interface{}{
		"selector": composedData.Spec.Selector,
		"jwtRules": composedData.Spec.JwtRules,
	})
}

func (br *IstioBridge) DeleteRequestAuthentication(name string, namespace string) error {

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	return br.securityResources.DeleteRequestAuthentication(ctx, namespace, name, k8smetav1.DeleteOptions{})

}
