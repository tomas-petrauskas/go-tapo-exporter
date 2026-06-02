package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	tapo "github.com/tess1o/tapo-go"
)

func (p *PrometheusExporter) handleOnePowerStripMetric(deviceId string, key string, v float64, labels map[string]string) {
	metricName := p.Config.Prefix + "_" + key
	cacheKey := deviceId + "_" + metricName
	p.mu.Lock()
	gauge, exists := p.metrics[cacheKey]
	p.mu.Unlock()
	if !exists {
		slog.Debug("Adding new gauge metric for power strip socket", "metric", metricName, "deviceId", deviceId)
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        metricName,
			ConstLabels: labels,
		})
		prometheus.MustRegister(gauge)
		p.mu.Lock()
		p.metrics[cacheKey] = gauge
		p.mu.Unlock()
	}
	gauge.Set(v)
}

func (p *PrometheusExporter) HandlePowerStripSocket(_ context.Context, host string, name string, socket tapo.PowerStripSocket) {
	slog.Info("Handling prometheus metrics for power strip socket",
		"host", host,
		"name", name,
		"position", socket.Position,
		"socket_name", socket.Nickname,
	)

	labels := map[string]string{
		"host":            host,
		"name":            name,
		"socket_position": fmt.Sprintf("%d", socket.Position),
		"socket_name":     socket.Nickname,
	}

	onVal := 0.0
	if socket.DeviceOn {
		onVal = 1.0
	}

	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_on", onVal, labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_on_time_seconds", float64(socket.OnTime), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_today_runtime_minutes", float64(socket.TodayRuntime), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_month_runtime_minutes", float64(socket.MonthRuntime), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_today_energy_wh", float64(socket.TodayEnergy), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_month_energy_wh", float64(socket.MonthEnergy), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_current_power_w", float64(socket.CurrentPower), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_power_mw", float64(socket.PowerMw), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_current_ma", float64(socket.CurrentMa), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_voltage_mv", float64(socket.VoltageMv), labels)
	p.handleOnePowerStripMetric(socket.DeviceID, "power_strip_socket_cumulative_energy_wh", float64(socket.EnergyWh), labels)
}
