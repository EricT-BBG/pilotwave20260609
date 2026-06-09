package gateway_manager

type PortData struct {
	Name           string `json:"name"`
	Port           uint32 `json:"port"`
	Protocol       string `json:"protocol"`
	CredentialName string `json:"credentialname"`
	Mode           string `json:"mode"`
	Cert           string `json:"cert"`
	Pkey           string `json:"pkey"`
	Cacert         string `json:"cacert"`
}

type GatewayServersData struct {
	Hosts []string   `json:"hosts"`
	Ports []PortData `json:"ports"`
}

type GatewayResponse struct {
	Name                string               `json:"name"`
	Namespace           string               `json:"namespace"`
	Servers             []GatewayServersData `json:"servers"`
	SelectorMatchLabels map[string]string    `json:"selectormatchlabels"`
	CreatedAt           int64                `json:"createdAt"`
	ResourceVersion     string               `json:"resourceversion"`
}

type GatewayTLSCertificateResponse struct {
	ServerIndex       int      `json:"serverIndex"`
	Port              uint32   `json:"port"`
	Protocol          string   `json:"protocol"`
	Hosts             []string `json:"hosts"`
	CredentialName    string   `json:"credentialName"`
	SecretNamespace   string   `json:"secretNamespace"`
	SecretName        string   `json:"secretName"`
	Status            string   `json:"status"`
	Reason            string   `json:"reason,omitempty"`
	DaysUntilExpiry   int      `json:"daysUntilExpiry"`
	NotBefore         string   `json:"notBefore,omitempty"`
	NotAfter          string   `json:"notAfter,omitempty"`
	Subject           string   `json:"subject,omitempty"`
	Issuer            string   `json:"issuer,omitempty"`
	DNSNames          []string `json:"dnsNames,omitempty"`
	FingerprintSHA256 string   `json:"fingerprintSHA256,omitempty"`
}

type RouterData struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type GatewayMappingResponse struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Routers   []RouterData `json:"routers"`
}

type PortsRequest struct {
	Name           string `json:"name"`
	Port           uint32 `json:"port"`
	Protocol       string `json:"protocol"`
	CredentialName string `json:"credentialname"`
	Mode           string `json:"mode"`
	Cert           string `json:"cert"`
	Pkey           string `json:"pkey"`
	Cacert         string `json:"cacert"`
}

type GatewayRequestServersData struct {
	Hosts []string       `json:"hosts"`
	Ports []PortsRequest `json:"ports"`
}

type GatewayRequest struct {
	Name                string                      `json:"name"`
	Namespace           string                      `json:"namespace"`
	Servers             []GatewayRequestServersData `json:"servers"`
	SelectorMatchLabels map[string]string           `json:"selectormatchlabels"`
	ResourceVersion     *string                     `json:"resourceversion"`
}

type GatewayUpdateRequest struct {
	Servers []GatewayRequestServersData `json:"servers"`
}

type GatewayEnableRequest struct {
	Enabled bool `json:"enabled"`
}

type BlackWhiteListResponse struct {
	ID          string `json:"id"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
	Category    string `json:"category"`
	UpdatedAt   int64  `json:"updatedAt"`
	CreatedAt   int64  `json:"createdAt"`
}

type BlackWhiteListRequest struct {
	Domain      string `json:"domain"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type GatewayMappinRouterData struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type GayewayMappingRequest struct {
	Routers          []GatewayMappinRouterData `json:"routers"`
	ResourceVersions map[string]string         `json:"resourceversions"`
}

type RouterMappingResponse struct {
	Name             string                    `json:"name"`
	Namespace        string                    `json:"namespace"`
	Routers          []GatewayMappinRouterData `json:"routers"`
	ResourceVersions map[string]string         `json:"resourceversions"`
}

type Gateway interface {
}
