package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func counterValue(t *testing.T, collector prometheus.Collector, labels map[string]string) float64 {
	t.Helper()

	for _, metric := range collectMetrics(t, collector) {
		if labelsMatch(metric, labels) && metric.GetCounter() != nil {
			return metric.GetCounter().GetValue()
		}
	}

	t.Fatalf("counter with labels %v was not collected", labels)
	return 0
}

func gaugeValue(t *testing.T, collector prometheus.Collector, labels map[string]string) float64 {
	t.Helper()

	for _, metric := range collectMetrics(t, collector) {
		if labelsMatch(metric, labels) && metric.GetGauge() != nil {
			return metric.GetGauge().GetValue()
		}
	}

	t.Fatalf("gauge with labels %v was not collected", labels)
	return 0
}

func histogramCount(t *testing.T, collector prometheus.Collector, labels map[string]string) uint64 {
	t.Helper()

	for _, metric := range collectMetrics(t, collector) {
		if labelsMatch(metric, labels) && metric.GetHistogram() != nil {
			return metric.GetHistogram().GetSampleCount()
		}
	}

	t.Fatalf("histogram with labels %v was not collected", labels)
	return 0
}

func collectMetrics(t *testing.T, collector prometheus.Collector) []*dto.Metric {
	t.Helper()

	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []*dto.Metric
	for metric := range ch {
		dtoMetric := &dto.Metric{}
		if err := metric.Write(dtoMetric); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		metrics = append(metrics, dtoMetric)
	}

	return metrics
}

func labelsMatch(metric *dto.Metric, expected map[string]string) bool {
	seen := make(map[string]string, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		seen[label.GetName()] = label.GetValue()
	}

	for name, value := range expected {
		if seen[name] != value {
			return false
		}
	}

	return len(seen) == len(expected)
}
