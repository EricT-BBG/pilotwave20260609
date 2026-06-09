package router

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.brobridge.com/pilotwave/pilotwave/pkg/app"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager"
	"git.brobridge.com/pilotwave/pilotwave/pkg/router_manager/router/model"
	//log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Router struct {
	app           app.App
	host          string
	port          string
	url           string
	useTLS        bool
	skipTLSVerify bool
	token         string
}

const (
	monitoringProviderGrafana    = "grafana"
	monitoringProviderPrometheus = "prometheus"
	defaultGrafanaDatasourceID   = "1"
	monitoringTestTimeout       = 5 * time.Second
)

func prometheusLabelValue(value string) string {
	return strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`).Replace(value)
}

func parsePrometheusFloat(value string) float64 {
	if value == "NaN" || value == "+Inf" || value == "-Inf" {
		return 0
	}

	val, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}

	return val
}

func buildSuccessRateQuery(service string) string {
	service = prometheusLabelValue(service)
	return fmt.Sprintf(`100 * sum(rate(istio_requests_total{reporter="destination",destination_service="%s",response_code!~"5.*"}[5m])) / sum(rate(istio_requests_total{reporter="destination",destination_service="%s"}[5m]))`, service, service)
}

func buildTotalRequestQuery(service string) string {
	service = prometheusLabelValue(service)
	return fmt.Sprintf(`sum(istio_requests_total{reporter="destination",destination_service="%s"})`, service)
}

func buildTotalSuccessRequestQuery(service string) string {
	service = prometheusLabelValue(service)
	return fmt.Sprintf(`sum(istio_requests_total{reporter="destination",destination_service="%s",response_code!~"5.*"})`, service)
}

func buildLatencyQuery(service string, percentage float64) string {
	service = prometheusLabelValue(service)
	return fmt.Sprintf(`histogram_quantile(%v, sum(rate(istio_request_duration_seconds_bucket{reporter="destination",destination_service="%s"}[5m])) by (le)) or (histogram_quantile(%v, sum(rate(istio_request_duration_milliseconds_bucket{reporter="destination",destination_service="%s"}[5m])) by (le)) / 1000)`, percentage, service, percentage, service)
}

func buildOPSQuery(service string) string {
	service = prometheusLabelValue(service)
	return fmt.Sprintf(`round(sum(rate(istio_requests_total{reporter="destination",destination_service="%s"}[5m])), 0.001)`, service)
}

func normalizeMonitoringProvider(provider string) string {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == monitoringProviderPrometheus {
		return monitoringProviderPrometheus
	}

	return monitoringProviderGrafana
}

func normalizeDatasourceID(datasourceID string) string {
	datasourceID = strings.TrimSpace(datasourceID)
	if datasourceID == "" {
		return defaultGrafanaDatasourceID
	}

	return datasourceID
}

func grafanaPrometheusQueryRangePath(datasourceID string) string {
	return fmt.Sprintf("/api/datasources/proxy/%s/api/v1/query_range", normalizeDatasourceID(datasourceID))
}

func grafanaPrometheusQueryPath(datasourceID string) string {
	return fmt.Sprintf("/api/datasources/proxy/%s/api/v1/query", normalizeDatasourceID(datasourceID))
}

func prometheusQueryRangePath(provider string, datasourceID string) string {
	if normalizeMonitoringProvider(provider) == monitoringProviderPrometheus {
		return "/api/v1/query_range"
	}

	return grafanaPrometheusQueryRangePath(datasourceID)
}

func prometheusQueryPath(provider string, datasourceID string) string {
	if normalizeMonitoringProvider(provider) == monitoringProviderPrometheus {
		return "/api/v1/query"
	}

	return grafanaPrometheusQueryPath(datasourceID)
}

func NewRouter(a app.App) *Router {
	// Get grafana url
	host := viper.GetString("grafana.host")
	port := viper.GetString("grafana.port")
	token := viper.GetString("grafana.token")
	skipTLSVerify := viper.GetBool("grafana.skip_tls_verify")

	protocol := "http"
	useTLS := viper.GetBool("grafana.tls")
	if useTLS {
		protocol = "https"
	}

	// Generate url
	_url := fmt.Sprintf("%s://%s:%v", protocol, host, port)

	return &Router{
		app:           a,
		host:          host,
		port:          port,
		url:           _url,
		useTLS:        useTLS,
		skipTLSVerify: skipTLSVerify,
		token:         token,
	}
}

func (r *Router) FetchGrafana() model.Grafana {

	// Getting data from database
	db := r.app.GetDatabase()
	grafana := model.Grafana{}

	if db.First(&grafana).RecordNotFound() {
		return model.Grafana{}
	}

	return grafana
}

func (r *Router) UpdateGrafana(req router_manager.GrafanaRequest) (string, error) {
	provider := normalizeMonitoringProvider(req.Provider)
	datasourceID := normalizeDatasourceID(req.DatasourceID)
	skipTLSVerify := req.IsTls && req.SkipTLSVerify

	// Getting data from database
	db := r.app.GetDatabase()
	gr := model.Grafana{
		Provider:      provider,
		Host:          req.Host,
		Port:          req.Port,
		Token:         req.Token,
		DatasourceID:  datasourceID,
		Tls:           req.IsTls,
		SkipTLSVerify: skipTLSVerify,
	}
	if req.ID == "" {
		// Create Grafana
		err := db.Create(&gr).Error
		if err != nil {
			log.Println(err)
			return "", err
		}

		return gr.ID, err
	} else {
		// Update Grafana
		err := db.Model(&gr).Where("id = ?", req.ID).Updates(map[string]interface{}{
			"Provider":      provider,
			"Host":          req.Host,
			"Port":          req.Port,
			"Token":         req.Token,
			"DatasourceID":  datasourceID,
			"Tls":           req.IsTls,
			"SkipTLSVerify": skipTLSVerify,
		}).Error

		if err != nil {
			return "", err
		}

		return req.ID, err
	}
}

func (r *Router) GetGrafana() (*router_manager.GrafanaConfig, error) {
	grafana := r.FetchGrafana()

	provider := normalizeMonitoringProvider(viper.GetString("grafana.provider"))
	host := r.host
	port := r.port
	useTLS := r.useTLS
	skipTLSVerify := r.skipTLSVerify
	token := r.token
	datasourceID := normalizeDatasourceID(viper.GetString("grafana.datasource_id"))

	if grafana.ID != "" {
		provider = normalizeMonitoringProvider(grafana.Provider)
		host = grafana.Host
		useTLS = grafana.Tls
		skipTLSVerify = grafana.SkipTLSVerify
		port = grafana.Port
		token = grafana.Token
		datasourceID = normalizeDatasourceID(grafana.DatasourceID)
	}

	return &router_manager.GrafanaConfig{
		ID:            grafana.ID,
		Provider:      provider,
		Host:          host,
		Port:          port,
		Token:         token,
		DatasourceID:  datasourceID,
		Tls:           useTLS,
		SkipTLSVerify: skipTLSVerify,
		CreatedAt:     grafana.CreatedAt.Unix(),
		UpdatedAt:     grafana.UpdatedAt.Unix(),
	}, nil
}

func (r *Router) currentMonitoringConfig() *router_manager.GrafanaConfig {
	config, err := r.GetGrafana()
	if err != nil {
		return &router_manager.GrafanaConfig{
			Provider:      monitoringProviderGrafana,
			Host:          r.host,
			Port:          r.port,
			Token:         r.token,
			DatasourceID:  defaultGrafanaDatasourceID,
			Tls:           r.useTLS,
			SkipTLSVerify: r.skipTLSVerify,
		}
	}

	return config
}

func (r *Router) monitoringBaseURL(config *router_manager.GrafanaConfig) string {
	protocol := "http"
	if config.Tls {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%v", protocol, config.Host, config.Port)
}

func (r *Router) TestGrafana(req router_manager.GrafanaRequest) (*router_manager.MonitoringTestResponse, error) {
	skipTLSVerify := req.IsTls && req.SkipTLSVerify
	config := &router_manager.GrafanaConfig{
		Provider:      normalizeMonitoringProvider(req.Provider),
		Host:          req.Host,
		Port:          req.Port,
		Token:         req.Token,
		DatasourceID:  normalizeDatasourceID(req.DatasourceID),
		Tls:           req.IsTls,
		SkipTLSVerify: skipTLSVerify,
	}
	if config.Host == "" || config.Port == "" {
		return nil, errors.New("monitoring source host and port are required")
	}

	baseURL, _ := url.Parse(r.monitoringBaseURL(config))
	baseURL.Path = baseURL.Path + prometheusQueryPath(config.Provider, config.DatasourceID)

	params := url.Values{}
	params.Add("query", "up")
	baseURL.RawQuery = params.Encode()

	res, err := r.requestWithTimeout("GET", baseURL.String(), config.Tls, config.SkipTLSVerify, config.Token, "", monitoringTestTimeout)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return &router_manager.MonitoringTestResponse{
				OK:      false,
				Message: fmt.Sprintf("monitoring source test timed out after %s", monitoringTestTimeout),
			}, nil
		}
		return nil, err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return &router_manager.MonitoringTestResponse{
			OK:      false,
			Message: fmt.Sprintf("monitoring source returned HTTP %d: %s", res.StatusCode, string(body)),
		}, nil
	}

	var promResp struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(body, &promResp); err != nil {
		return &router_manager.MonitoringTestResponse{
			OK:      false,
			Message: "monitoring source did not return a Prometheus-compatible response",
		}, nil
	}

	if promResp.Status != "success" {
		if promResp.Error == "" {
			promResp.Error = "query status was not success"
		}
		return &router_manager.MonitoringTestResponse{
			OK:      false,
			Message: promResp.Error,
		}, nil
	}

	return &router_manager.MonitoringTestResponse{
		OK:      true,
		Message: "Monitoring source is reachable and returned a Prometheus-compatible response.",
	}, nil
}

func (r *Router) DeleteGrafana(id string) (string, error) {

	// Getting data from database
	db := r.app.GetDatabase()
	ro := model.Grafana{}

	err := db.Unscoped().Delete(&ro, "id = ?", id).Error
	if err != nil {
		return "", err
	}

	return id, err
}

func (r *Router) GetRouterSuccessRate(req router_manager.RouterSuccessRateRequest) (router_manager.RouterSuccessRateResponse, error) {
	config := r.currentMonitoringConfig()
	uri := r.monitoringBaseURL(config)

	// Generate result
	results := []router_manager.RouterSuccessRateResponse{}

	//services := []string{"nginx.jhe-test.svc.cluster.local", "nginx.test.svc.cluster.local"}
	for _, service := range req.Services {
		baseURL, _ := url.Parse(uri)

		// add api path
		baseURL.Path = baseURL.Path + prometheusQueryRangePath(config.Provider, config.DatasourceID)

		// add Query Parameters
		params := url.Values{}

		queryStr := buildSuccessRateQuery(service)

		// queryStr := fmt.Sprintf("sum(istio_requests_total{destination_service=~\"%s\"})", service)

		params.Add("query", queryStr)
		params.Add("start", strconv.Itoa(req.StartTime))
		params.Add("end", strconv.Itoa(req.EndTime))
		params.Add("step", req.Interval)

		baseURL.RawQuery = params.Encode()

		// Connect grafana api and parser data
		res, err := r.Request("GET", baseURL.String(), config.Tls, config.SkipTLSVerify, config.Token, "")
		if err != nil {
			return router_manager.RouterSuccessRateResponse{}, err
		}

		// parsing grafana response
		var grafanaResponse router_manager.GrafanaResponse
		resbody, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(resbody, &grafanaResponse)
		if err != nil {
			return router_manager.RouterSuccessRateResponse{}, err
		}

		// Generate result
		result := router_manager.RouterSuccessRateResponse{}

		valTotal := float64(0)
		valCounter := float64(0)
		for _, r := range grafanaResponse.Data.Result {
			for _, v := range r.Values {
				// parsing value
				valStr := v[1].(string)
				val := parsePrometheusFloat(valStr)

				valTotal += val
				valCounter += 1

				result.Metrics = append(result.Metrics, router_manager.RouterSuccessRateData{
					Timestamp: int(v[0].(float64)),
					Value:     val,
				})

			}

			result.SuccessRate = valTotal / valCounter
			if valCounter == 0 {
				result.SuccessRate = float64(0)
			}
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return router_manager.RouterSuccessRateResponse{}, nil
	}

	// avg
	sr := float64(0)
	metrics := make([]router_manager.RouterSuccessRateData, len(results[0].Metrics))
	for i, res := range results {
		fmt.Println()

		sr = (sr*float64(i) + res.SuccessRate) / float64((i + 1))

		for x, m := range res.Metrics {
			metrics[x].Timestamp = m.Timestamp
			metrics[x].Value = (metrics[x].Value*float64(i) + m.Value) / float64((i + 1))
		}
	}

	startReq, _ := r.GetTotalRequest(req.StartTime, req)
	endReq, _ := r.GetTotalRequest(req.EndTime, req)

	total := int64(0)
	if endReq <= 0 {
		total = startReq
	} else {
		total = endReq - startReq
	}

	startSuccessReq, _ := r.GetTotalSuccessRequest(req.StartTime, req)
	endSuccessReq, _ := r.GetTotalSuccessRequest(req.EndTime, req)

	totalSuccess := int64(0)
	if endSuccessReq <= 0 {
		totalSuccess = startSuccessReq
	} else {
		totalSuccess = endSuccessReq - startSuccessReq
	}

	response := router_manager.RouterSuccessRateResponse{
		SuccessRate:          sr,
		Metrics:              metrics,
		TotalRequests:        total,
		TotalSuccessRequests: totalSuccess,
	}

	return response, nil

}

func (r *Router) GetTotalRequest(moment int, req router_manager.RouterSuccessRateRequest) (int64, error) {
	config := r.currentMonitoringConfig()
	uri := r.monitoringBaseURL(config)

	baseURL, _ := url.Parse(uri)

	// add api path
	baseURL.Path = baseURL.Path + prometheusQueryRangePath(config.Provider, config.DatasourceID)

	// add Query Parameters
	params := url.Values{}

	total := int64(0)
	for _, service := range req.Services {
		queryStr := buildTotalRequestQuery(service)

		params.Add("query", queryStr)
		params.Add("start", strconv.Itoa(moment))
		params.Add("end", strconv.Itoa(moment))
		params.Add("step", req.Interval)

		baseURL.RawQuery = params.Encode()

		// Connect grafana api and parser data
		res, err := r.Request("GET", baseURL.String(), config.Tls, config.SkipTLSVerify, config.Token, "")
		if err != nil {
			return 0, err
		}

		// parsing grafana response
		var grafanaResponse router_manager.GrafanaResponse
		resbody, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(resbody, &grafanaResponse)

		if err != nil {
			return 0, err
		}

		for _, r := range grafanaResponse.Data.Result {
			for _, v := range r.Values {

				// parsing value
				requests := parsePrometheusFloat(v[1].(string))
				total += int64(requests)
			}
		}
	}

	return total, nil
}

func (r *Router) GetTotalSuccessRequest(moment int, req router_manager.RouterSuccessRateRequest) (int64, error) {
	config := r.currentMonitoringConfig()
	uri := r.monitoringBaseURL(config)

	baseURL, _ := url.Parse(uri)

	// add api path
	baseURL.Path = baseURL.Path + prometheusQueryRangePath(config.Provider, config.DatasourceID)

	// add Query Parameters
	params := url.Values{}

	total := int64(0)
	for _, service := range req.Services {
		queryStr := buildTotalSuccessRequestQuery(service)

		params.Add("query", queryStr)
		params.Add("start", strconv.Itoa(moment))
		params.Add("end", strconv.Itoa(moment))
		params.Add("step", req.Interval)

		baseURL.RawQuery = params.Encode()

		// Connect grafana api and parser data
		res, err := r.Request("GET", baseURL.String(), config.Tls, config.SkipTLSVerify, config.Token, "")
		if err != nil {
			return 0, err
		}

		// parsing grafana response
		var grafanaResponse router_manager.GrafanaResponse
		resbody, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(resbody, &grafanaResponse)

		if err != nil {
			return 0, err
		}

		for _, r := range grafanaResponse.Data.Result {
			for _, v := range r.Values {

				// parsing value
				requests := parsePrometheusFloat(v[1].(string))
				total += int64(requests)
			}
		}
	}

	return total, nil
}

func (r *Router) GetRouterLatency(req router_manager.RouterLatencyRequest) (router_manager.RouterLatencyResponse, error) {
	config := r.currentMonitoringConfig()
	uri := r.monitoringBaseURL(config)

	// Generate result
	results := []router_manager.RouterLatencyResponse{}

	//services := []string{"nginx.jhe-test.svc.cluster.local", "nginx.test.svc.cluster.local"}
	for _, service := range req.Services {

		baseURL, _ := url.Parse(uri)

		// add api path
		baseURL.Path = baseURL.Path + prometheusQueryRangePath(config.Provider, config.DatasourceID)

		// add Query Parameters
		params := url.Values{}

		queryStr := buildLatencyQuery(service, req.Percentage)

		params.Add("query", queryStr)
		params.Add("start", strconv.Itoa(req.StartTime))
		params.Add("end", strconv.Itoa(req.EndTime))
		params.Add("step", req.Interval)

		baseURL.RawQuery = params.Encode()

		// Connect grafana api and parser data
		res, err := r.Request("GET", baseURL.String(), config.Tls, config.SkipTLSVerify, config.Token, "")
		if err != nil {
			return router_manager.RouterLatencyResponse{}, err
		}

		// parsing grafana response
		var grafanaResponse router_manager.GrafanaResponse
		resbody, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(resbody, &grafanaResponse)
		if err != nil {
			return router_manager.RouterLatencyResponse{}, err
		}

		// Generate result
		result := router_manager.RouterLatencyResponse{}

		for _, r := range grafanaResponse.Data.Result {
			for _, v := range r.Values {

				// parsing value
				valStr := v[1].(string)
				val := parsePrometheusFloat(valStr)

				result.Metrics = append(result.Metrics, router_manager.RouterLatencyData{
					Timestamp: int(v[0].(float64)),
					Value:     val,
				})

			}

			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return router_manager.RouterLatencyResponse{}, nil
	}

	// avg
	metrics := make([]router_manager.RouterLatencyData, len(results[0].Metrics))
	for i, res := range results {

		for x, m := range res.Metrics {
			metrics[x].Timestamp = m.Timestamp
			metrics[x].Value = (metrics[x].Value*float64(i) + m.Value) / float64((i + 1))
		}
	}

	response := router_manager.RouterLatencyResponse{
		Metrics: metrics,
	}

	return response, nil

}

func (r *Router) GetRouterOPS(req router_manager.RouterOPSRequest) (router_manager.RouterOPSResponse, error) {
	config := r.currentMonitoringConfig()
	uri := r.monitoringBaseURL(config)

	// Generate result
	results := []router_manager.RouterOPSResponse{}

	//services := []string{"nginx.jhe-test.svc.cluster.local", "nginx.test.svc.cluster.local"}
	for _, service := range req.Services {

		baseURL, _ := url.Parse(uri)

		// add api path
		baseURL.Path = baseURL.Path + prometheusQueryRangePath(config.Provider, config.DatasourceID)

		// add Query Parameters
		params := url.Values{}

		queryStr := buildOPSQuery(service)

		params.Add("query", queryStr)
		params.Add("start", strconv.Itoa(req.StartTime))
		params.Add("end", strconv.Itoa(req.EndTime))
		params.Add("step", req.Interval)

		baseURL.RawQuery = params.Encode()

		// Connect grafana api and parser data
		res, err := r.Request("GET", baseURL.String(), config.Tls, config.SkipTLSVerify, config.Token, "")
		if err != nil {
			return router_manager.RouterOPSResponse{}, err
		}

		// parsing grafana response
		var grafanaResponse router_manager.GrafanaResponse
		resbody, _ := ioutil.ReadAll(res.Body)
		err = json.Unmarshal(resbody, &grafanaResponse)
		if err != nil {
			return router_manager.RouterOPSResponse{}, err
		}

		// Generate result
		result := router_manager.RouterOPSResponse{}

		for _, r := range grafanaResponse.Data.Result {
			for _, v := range r.Values {

				// parsing value
				valStr := v[1].(string)
				val := parsePrometheusFloat(valStr)

				result.Metrics = append(result.Metrics, router_manager.RouterOPSData{
					Timestamp: int(v[0].(float64)),
					Value:     val,
				})

			}

			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return router_manager.RouterOPSResponse{}, nil
	}

	// avg
	metrics := make([]router_manager.RouterOPSData, len(results[0].Metrics))
	for i, res := range results {

		for x, m := range res.Metrics {
			metrics[x].Timestamp = m.Timestamp
			metrics[x].Value = (metrics[x].Value*float64(i) + m.Value) / float64((i + 1))
		}
	}

	response := router_manager.RouterOPSResponse{
		Metrics: metrics,
	}

	return response, nil

}

func (r *Router) Request(method string, uri string, useTLS bool, skipTLSVerify bool, token string, payload string) (*http.Response, error) {
	return r.requestWithTimeout(method, uri, useTLS, skipTLSVerify, token, payload, 0)
}

func (r *Router) requestWithTimeout(method string, uri string, useTLS bool, skipTLSVerify bool, token string, payload string, timeout time.Duration) (*http.Response, error) {
	client := &http.Client{Timeout: timeout}
	if useTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
		}
	}
	req, _ := http.NewRequest(method, uri, strings.NewReader(payload))
	//set header
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
