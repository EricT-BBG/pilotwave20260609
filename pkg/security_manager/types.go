package security_manager

// RequestAuthentication

type RequestAuthenticationJWTData struct {
	Issuer    string   `json:"issuer"`
	JwksUri   string   `json:"jwksUri"`
	Audiences []string `json:"audiences"`
}

type LabelDtata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type RequestAuthenticationRequest struct {
	Name                string                         `json:"name"`
	Namespace           string                         `json:"namespace"`
	ResourceVersion     *string                        `json:"resourceversion"`
	JWTRules            []RequestAuthenticationJWTData `json:"jwtRules"`
	SelectorMatchLabels []LabelDtata                   `json:"selectorMatchLabels"`
}

type RequestAuthenticationUpdateRequest struct {
	ResourceVersion     *string                        `json:"resourceversion"`
	JWTRules            []RequestAuthenticationJWTData `json:"jwtRules"`
	SelectorMatchLabels []LabelDtata                   `json:"selectorMatchLabels"`
}

type RequestAuthenticationResponse struct {
	Name                string                         `json:"name"`
	Namespace           string                         `json:"namespace"`
	ResourceVersion     string                         `json:"resourceversion"`
	JWTRules            []RequestAuthenticationJWTData `json:"jwtRules"`
	SelectorMatchLabels []LabelDtata                   `json:"selectorMatchLabels"`
	CreatedAt           int64                          `json:"createdAt"`
}

// AuthorizationPolicy

type RulesWhenData struct {
	Key       string   `json:"key"`
	Values    []string `json:"values"`
	NotValues []string `json:"notValues"`
}

type AuthorizationPolicyRuleSourceData struct {
	Principals           []string `json:"principals"`
	NotPrincipals        []string `json:"notPrincipals"`
	RequestPrincipals    []string `json:"requestPrincipals"`
	NotRequestPrincipals []string `json:"notRequestPrincipals"`
	Namespaces           []string `json:"namespaces"`
	NotNamespaces        []string `json:"notNamespaces"`
	IpBlocks             []string `json:"ipBlocks"`
	NotIpBlocks          []string `json:"notIpBlocks"`
	RemoteIpBlocks       []string `json:"remoteIpBlocks"`
	NotRemoteIpBlocks    []string `json:"notRemoteIpBlocks"`
}

type AuthorizationPolicyRuleDestinationData struct {
	Hosts      []string `json:"hosts"`
	NotHosts   []string `json:"notHosts"`
	Ports      []string `json:"ports"`
	NotPorts   []string `json:"notPorts"`
	Methods    []string `json:"methods"`
	NotMethods []string `json:"notMethods"`
	Paths      []string `json:"paths"`
	NotPaths   []string `json:"notPaths"`
}

type AuthorizationPolicyFromRuleData struct {
	Source *AuthorizationPolicyRuleSourceData `json:"source"`
}

type AuthorizationPolicyToOperationData struct {
	Operation *AuthorizationPolicyRuleDestinationData `json:"operation"`
}

type AuthorizationPolicyRuleData struct {
	From []*AuthorizationPolicyFromRuleData    `json:"from"`
	To   []*AuthorizationPolicyToOperationData `json:"to"`
	When []*RulesWhenData                      `json:"when"`
}

type AuthorizationPolicyUpdateRequest struct {
	ResourceVersion     *string                        `json:"resourceversion"`
	Rules               []*AuthorizationPolicyRuleData `json:"rules"`
	SelectorMatchLabels []LabelDtata                   `json:"selectorMatchLabels"`
	Action              string                         `json:"action"`
}

type AuthorizationPolicyRequest struct {
	Name                string                         `json:"name"`
	Namespace           string                         `json:"namespace"`
	ResourceVersion     *string                        `json:"resourceversion"`
	Rules               []*AuthorizationPolicyRuleData `json:"rules"`
	SelectorMatchLabels []LabelDtata                   `json:"selectorMatchLabels"`
	Action              string                         `json:"action"`
}

type AuthorizationPolicyResponse struct {
	Name                string                        `json:"name"`
	Namespace           string                        `json:"namespace"`
	ResourceVersion     string                        `json:"resourceVersion"`
	Rules               []AuthorizationPolicyRuleData `json:"rules"`
	SelectorMatchLabels []LabelDtata                  `json:"selectorMatchLabels"`
	Action              string                        `json:"action"`
	CreatedAt           int64                         `json:"createdAt"`
}

type Security interface {
}
