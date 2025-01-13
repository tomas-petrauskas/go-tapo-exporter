package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

func (p *PrometheusExporter) HandleSmartPlug(_ context.Context, device *SmartPlug, rawParameters map[string]interface{}) {
	slog.Debug("Handling prometheus metrics for device", "sn", device.Name)
	for field, val := range rawParameters {
		p.handleOneMetricSmartPlug(device, field, val)
	}
}

func (p *PrometheusExporter) handleOneMetricSmartPlug(device *SmartPlug, field string, val interface{}) {
	metricName := p.Config.Prefix + "_" + field
	deviceMetricName := device.Host + "_" + metricName
	p.mu.Lock()
	gauge, ok := p.metrics[deviceMetricName]
	p.mu.Unlock()
	if !ok {
		slog.Debug("Adding new metric", "metric", metricName, "host", device.Host, "device_name", device.Name)
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: metricName,
			ConstLabels: map[string]string{
				"device": device.Name,
				"host":   device.Host,
			},
		})
		prometheus.MustRegister(gauge)
		p.mu.Lock()
		p.metrics[deviceMetricName] = gauge
		p.mu.Unlock()
	} else {
		slog.Debug("Updating metric", "metric", metricName, "value", val, "ip_address", device.Host, "device_name", device.Name)
	}
	floatVal, ok := val.(float64)
	if ok {
		gauge.Set(floatVal)
	} else {
		if arrVal, arrValOk := val.([]interface{}); arrValOk {
			for i, v := range arrVal {
				arrFloatVal, okVal := v.(float64)
				if okVal {
					p.handleOneMetricSmartPlug(device, fmt.Sprintf("%s_%d", field, i), arrFloatVal)
				} else {
					slog.Debug("Metric value is not a float", "metric", metricName, "value", v, "host", device.Host, "device_name", device.Name)
				}
			}
		} else {
			slog.Debug("Metric value is not a float or an array", "metric", metricName, "value", val, "host", device.Host, "device_name", device.Name)
		}
		slog.Debug("Metric value is not a float", "metric", metricName, "value", val, "host", device.Host, "device_name", device.Name)
	}
}
