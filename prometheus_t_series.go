package main

import (
	"context"
	"encoding/base64"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tess1o/tapo-go"
	"log/slog"
	"time"
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
	slog.Info("Handling prometheus metrics for t-series device", "device_id", parameters.DeviceId, "hub_host", hubHost, "nickname", p.getNickName(parameters))

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
	labels["nickname"] = p.getNickName(parameters)
	labels["last_onboarding_timestamp"] = p.convertUnixTime(int64(parameters.LastOnboardingTimestamp))
	if parameters.AtLowBattery {
		labels["at_low_battery"] = "1"
	} else {
		labels["at_low_battery"] = "0"
	}

	p.handleOneTSeriesMetric(parameters.DeviceId, "current_temp", parameters.CurrentTemp, labels)
	p.handleOneTSeriesMetric(parameters.DeviceId, "current_humidity", float64(parameters.CurrentHumidity), labels)
}

func (p *PrometheusExporter) getNickName(parameters tapo.TSeriesResponse) string {
	nickname, err := base64.StdEncoding.DecodeString(parameters.Nickname)
	if err != nil {
		return parameters.Nickname
	} else {
		return string(nickname)
	}
}

func (p *PrometheusExporter) convertUnixTime(unixTime int64) string {
	t := time.Unix(unixTime, 0)
	return t.Format("2006-01-02 15:04:05")
}
