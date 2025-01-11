package main

import (
	"context"
	"github.com/tess1o/tapo-go"
	"log/slog"
)

func initTSeries(devices []TSeries, username string, password string) {
	for i := range devices {
		client, err := initTSeriesClient(devices[i].HubHost, username, password)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "device", devices[i].HubHost)
			continue
		}
		devices[i].Client = client
	}
}

func initTSeriesClient(host, username, password string) (*tapo.TSeries, error) {
	hub, err := tapo.NewHub(host, username, password, tapo.Options{})
	if err != nil {
		return nil, err
	}
	return tapo.NewTSeriesDevices(hub), nil
}

func handleTSeries(devices TSeries, username string, password string, exporter *PrometheusExporter) {
	if devices.Client == nil {
		client, err := initTSeriesClient(devices.HubHost, username, password)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "hub", devices.HubHost)
			return
		}
		devices.Client = client
	}
	r, err := devices.Client.GetTSeriesDevices()
	if err != nil {
		slog.Error("Error getting t-series device parameters", "hub", devices.HubHost, "error", err)
	} else {
		slog.Info("Successfully received metrics", "device", devices.HubHost)
		for _, params := range r {
			exporter.HandleTSeries(context.Background(), devices.HubHost, params)
		}
	}
}
