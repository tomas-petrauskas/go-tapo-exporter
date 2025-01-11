package main

import (
	"context"
	"encoding/base64"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tess1o/tapo-go"
	"log/slog"
)

func (p *PrometheusExporter) handleOneTSeriesMetric(deviceId string, key string, v float64, labels map[string]string) {
	metricName := p.Config.Prefix + "_" + key
	p.mu.Lock()
	gauge, exists := p.metrics[deviceId+"_"+metricName]
	p.mu.Unlock()
	if !exists {
		slog.Debug("Adding new gauge metric for T-series", "metric", metricName, "deviceId", deviceId)
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        metricName,
			ConstLabels: labels,
		})
		prometheus.MustRegister(gauge)
		p.mu.Lock()
		p.metrics[deviceId+"_"+metricName] = gauge
		p.mu.Unlock()
	}
	gauge.Set(v)
}

func (p *PrometheusExporter) HandleTSeries(_ context.Context, hubHost string, parameters tapo.TSeriesResponse) {
	slog.Debug("Handling prometheus metrics for t-series device")

	labels := make(map[string]string)
	labels["parent_device_id"] = parameters.ParentDeviceId
	labels["hw_ver"] = parameters.HwVer
	labels["fw_ver"] = parameters.FwVer
	labels["device_id"] = parameters.DeviceId
	labels["mac"] = parameters.Mac
	labels["model"] = parameters.Model
	labels["status"] = parameters.Status
	labels["temp_unit"] = parameters.TempUnit
	labels["hub_host"] = hubHost
	labels["nickname"] = parameters.Nickname

	nickname, err := base64.StdEncoding.DecodeString(parameters.Nickname)
	if err != nil {
		labels["nickname"] = parameters.Nickname
	} else {
		labels["nickname"] = string(nickname)
	}

	p.handleOneTSeriesMetric(parameters.DeviceId, "lastOnboardingTimestamp", float64(parameters.LastOnboardingTimestamp), labels)
	p.handleOneTSeriesMetric(parameters.DeviceId, "current_temp", parameters.CurrentTemp, labels)
	p.handleOneTSeriesMetric(parameters.DeviceId, "current_humidity", float64(parameters.CurrentHumidity), labels)
}
