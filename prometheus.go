package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"sync"
)

// PrometheusConfig represents the configuration for recording Prometheus metrics.
type PrometheusConfig struct {
	ServerPort string
	Prefix     string
	Devices    []Device
}

type PrometheusExporter struct {
	Config  *PrometheusConfig
	metrics map[string]prometheus.Gauge
	mu      sync.RWMutex
	Server  *http.Server
}

func NewPrometheusExporter(config *PrometheusConfig) *PrometheusExporter {
	slog.Debug("Creating prometheus exporter")

	// Set up HTTP server for Prometheus metrics
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: mux,
	}

	go func() {
		// Start the HTTP server
		slog.Debug("Starting HTTP server", "port", config.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server ListenAndServe error", "error", err)
		}
	}()
	return &PrometheusExporter{
		Config:  config,
		Server:  server,
		mu:      sync.RWMutex{},
		metrics: make(map[string]prometheus.Gauge),
	}
}

func (p *PrometheusExporter) Handle(_ context.Context, device Device, rawParameters map[string]interface{}) {
	slog.Debug("Handling prometheus metrics for device", "sn", device.Name)
	for field, val := range rawParameters {
		p.handleOneMetric(device, field, val)
	}
}

func (p *PrometheusExporter) handleOneMetric(device Device, field string, val interface{}) {
	metricName := p.Config.Prefix + "_" + field
	deviceMetricName := device.IPAddress + "_" + metricName
	p.mu.Lock()
	gauge, ok := p.metrics[deviceMetricName]
	p.mu.Unlock()
	if !ok {
		slog.Debug("Adding new metric", "metric", metricName, "ip_address", device.IPAddress, "device_name", device.Name)
		gauge = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: metricName,
			ConstLabels: map[string]string{
				"device":     device.Name,
				"ip_address": device.IPAddress,
			},
		})
		prometheus.MustRegister(gauge)
		p.mu.Lock()
		p.metrics[deviceMetricName] = gauge
		p.mu.Unlock()
	} else {
		slog.Debug("Updating metric", "metric", metricName, "value", val, "ip_address", device.IPAddress, "device_name", device.Name)
	}
	floatVal, ok := val.(float64)
	if ok {
		gauge.Set(floatVal)
	} else {
		if arrVal, arrValOk := val.([]interface{}); arrValOk {
			for i, v := range arrVal {
				arrFloatVal, okVal := v.(float64)
				if okVal {
					p.handleOneMetric(device, fmt.Sprintf("%s_%d", field, i), arrFloatVal)
				} else {
					slog.Debug("Metric value is not a float", "metric", metricName, "value", v, "ip_address", device.IPAddress, "device_name", device.Name)
				}
			}
		} else {
			slog.Debug("Metric value is not a float or an array", "metric", metricName, "value", val, "ip_address", device.IPAddress, "device_name", device.Name)
		}
		slog.Debug("Metric value is not a float", "metric", metricName, "value", val, "ip_address", device.IPAddress, "device_name", device.Name)
	}
}

func (p *PrometheusExporter) Close(ctx context.Context) {
	// Shutdown HTTP server
	slog.Debug("Shutting down HTTP server...")
	if err := p.Server.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	} else {
		slog.Debug("HTTP server gracefully stopped")
	}
}
