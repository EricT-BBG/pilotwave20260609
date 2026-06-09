package router_manager

type RouterResponse struct {
	Name             string   `json:"name"`
	Namespace        string   `json:"namespace"`
	Protocol         string   `json:"protocol"`
	Hosts            []string `json:"hosts"`
	HttpCount        int      `json:"httpCount"`
	DestinationCount int      `json:"destinationCount"`
	CreatedAt        int64    `json:"createdAt"`
	ResourceVersion  string   `json:"resourceversion"`
}

type RouterRequest struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Hosts     []string `json:"hosts"`
	Protocol  string   `json:"protocol"`
}

type RouterUpdateRequest struct {
	Protocol        string   `json:"protocol"`
	Hosts           []string `json:"hosts"`
	ResourceVersion string   `json:"resourceversion"`
}

type GatewayData struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Namespace string `json:"namespace"`
}

type RouterMappingGatewayData struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type RouterMappingRequest struct {
	// Name      string                     `json:"name"`
	// Namespace string                     `json:"namespace"`
	// Protocol  string                     `json:"protocol"`
	Gateways        []RouterMappingGatewayData `json:"gateways"`
	ResourceVersion string                     `json:"resourceversion"`
}

type RouterMappingResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// Protocol  string                     `json:"protocol"`
	Gateways        []RouterMappingGatewayData `json:"gateways"`
	ResourceVersion string                     `json:"resourceversion"`
}

type RouterAuthResponse struct {
	Host      string `json:"host"`
	Issuer    string `json:"issuer"`
	JwksUri   string `json:"jwksUri"`
	CreatedAt int64  `json:"createdAt"`
}

type RouterAuthRequest struct {
	Host    string `json:"host"`
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwksUri"`
}

type HeaderData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type DestinationData struct {
	Host   string `json:"host"`
	Port   uint32 `json:"port"`
	Weight int32  `json:"weight"`
	Subset string `json:"subset"`
}

type HttpsData struct {
	Prefixs      []string          `json:"prefixs"`
	Headers      []HeaderData      `json:"headers"`
	Rewrite      string            `json:"rewrite"`
	Destinations []DestinationData `json:"destinations"`
	FixedDelay   int64             `json:"fixedDelay"`
	Timeout      int64             `json:"timeout"`
}

type RouterRuleRequest struct {
	Https           []HttpsData `json:"https"`
	ResourceVersion string      `json:"resourceversion"`
}

type RouterRuleResponse struct {
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	Https           []HttpsData `json:"https"`
	ResourceVersion string      `json:"resourceversion"`
}

type RouterSuccessRateRequest struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Services  []string `json:"services"`
	StartTime int      `json:"startTime"`
	EndTime   int      `json:"endTime"`
	Interval  string   `json:"interval"`
}

type RouterSuccessRateResponse struct {
	SuccessRate          float64                 `json:"successRate"`
	TotalRequests        int64                   `json:"totalReqests"`
	TotalSuccessRequests int64                   `json:"totalSuccessReqests"`
	Metrics              []RouterSuccessRateData `json:"metrics"`
}

type RouterSuccessRateData struct {
	Timestamp int     `json:"timestamp"`
	Value     float64 `json:"value"`
}

type RouterLatencyRequest struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Percentage float64  `json:"percentage"`
	Services   []string `json:"services"`
	StartTime  int      `json:"startTime"`
	EndTime    int      `json:"endTime"`
	Interval   string   `json:"interval"`
}

type GrafanaResponse struct {
	Data GrafanaData `json:"data"`
}

type RouterLatencyResponse struct {
	Metrics []RouterLatencyData `json:"metrics"`
}

type RouterLatencyData struct {
	Timestamp int     `json:"timestamp"`
	Value     float64 `json:"value"`
}

type RouterOPSRequest struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Services  []string `json:"services"`
	StartTime int      `json:"startTime"`
	EndTime   int      `json:"endTime"`
	Interval  string   `json:"interval"`
}

type RouterOPSResponse struct {
	Metrics []RouterOPSData `json:"metrics"`
}

type RouterOPSData struct {
	Timestamp int     `json:"timestamp"`
	Value     float64 `json:"value"`
}

type GrafanaData struct {
	Result []GrafanaResultObj `json:"result"`
}

type GrafanaResultObj struct {
	Values [][]interface{} `json:"values"`
}

type GrafanaRequest struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	Token         string `json:"token"`
	DatasourceID  string `json:"datasourceId"`
	IsTls         bool   `json:"isTls"`
	SkipTLSVerify bool   `json:"skipTlsVerify"`
}

type GrafanaConfig struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	Host          string `json:"host"`
	Port          string `json:"port"`
	Token         string `json:"token"`
	DatasourceID  string `json:"datasourceId"`
	Tls           bool   `json:"isTls"`
	SkipTLSVerify bool   `json:"skipTlsVerify"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

type MonitoringTestResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type Router interface {
	UpdateGrafana(GrafanaRequest) (string, error)
	GetGrafana() (*GrafanaConfig, error)
	DeleteGrafana(GrafanaId string) (string, error)
	TestGrafana(GrafanaRequest) (*MonitoringTestResponse, error)

	GetRouterSuccessRate(RouterSuccessRateRequest) (RouterSuccessRateResponse, error)
	GetRouterLatency(RouterLatencyRequest) (RouterLatencyResponse, error)
	GetRouterOPS(RouterOPSRequest) (RouterOPSResponse, error)
}
