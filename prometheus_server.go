package main

import (
	"context"
	"errors"
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
	Devices    *Devices
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

func (p *PrometheusExporter) Close(ctx context.Context) {
	// Shutdown HTTP server
	slog.Debug("Shutting down HTTP server...")
	if err := p.Server.Shutdown(ctx); err != nil {
		slog.Error("HTTP server shutdown error", "error", err)
	} else {
		slog.Debug("HTTP server gracefully stopped")
	}
}
