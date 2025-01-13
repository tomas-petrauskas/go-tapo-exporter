package main

import (
	"context"
	"github.com/tess1o/tapo-go"
	"log/slog"
)

type TSeries struct {
	HubHost  string              `json:"hub_host"`
	Client   *tapo.TSeries       `json:"-"`
	Exporter *PrometheusExporter `json:"-"`
	Username string              `json:"-"`
	Password string              `json:"-"`
}

func initTSeries(devices []*TSeries, username string, password string, exporter *PrometheusExporter) {
	for i := range devices {
		devices[i].Exporter = exporter
		devices[i].Username = username
		devices[i].Password = password
		err := initTSeriesClient(devices[i])
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "device", devices[i].HubHost)
			continue
		}
	}
}

func initTSeriesClient(device *TSeries) error {
	slog.Info("Creating Tapo client", "host", device.HubHost)
	hub, err := tapo.NewHub(device.HubHost, device.Username, device.Password, tapo.Options{RetryConfig: tapo.DefaultRetryConfig})
	if err != nil {
		return err
	}
	tSeriesDevices := tapo.NewTSeriesDevices(hub)
	device.Client = tSeriesDevices
	return nil
}

func handleTSeries(devices *TSeries) {
	if devices.Client == nil {
		err := initTSeriesClient(devices)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "hub", devices.HubHost)
			return
		}
	}
	r, err := devices.Client.GetTSeriesDevices()
	if err != nil {
		slog.Error("Error getting t-series device parameters, will try to handshake again", "hub", devices.HubHost, "error", err)
		initErr := initTSeriesClient(devices)
		if initErr != nil {
			slog.Error("Failed to reinitialize Tapo client", "error", initErr, "hub", devices.HubHost)
			return
		}
	} else {
		for _, params := range r {
			devices.Exporter.HandleTSeries(context.Background(), devices.HubHost, params)
		}
	}
}
