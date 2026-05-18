package main

import (
	"context"
	"log/slog"

	tapo "github.com/tess1o/tapo-go"
)

type PowerStrip struct {
	Name     string              `json:"name"`
	Host     string              `json:"host"`
	Client   *tapo.PowerStrip    `json:"-"`
	Exporter *PrometheusExporter `json:"-"`
	Username string              `json:"-"`
	Password string              `json:"-"`
}

func initPowerStrips(devices []*PowerStrip, username string, password string, exporter *PrometheusExporter) {
	for i := range devices {
		devices[i].Exporter = exporter
		devices[i].Username = username
		devices[i].Password = password
		err := initPowerStripClient(devices[i])
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "device", devices[i].Host)
			continue
		}
	}
}

func initPowerStripClient(device *PowerStrip) error {
	slog.Info("Creating Tapo client", "host", device.Host)
	strip, err := tapo.NewPowerStrip(context.Background(), device.Host, device.Username, device.Password, tapo.Options{RetryConfig: tapo.DefaultRetryConfig})
	if err != nil {
		return err
	}
	device.Client = strip
	return nil
}

func handlePowerStrip(device *PowerStrip) {
	if device.Client == nil {
		err := initPowerStripClient(device)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "host", device.Host)
			return
		}
	}
	sockets, err := device.Client.GetSockets(context.Background())
	if err != nil {
		slog.Error("Error getting power strip socket data, will try to handshake again", "host", device.Host, "error", err)
		initErr := initPowerStripClient(device)
		if initErr != nil {
			slog.Error("Failed to reinitialize Tapo client", "error", initErr, "host", device.Host)
			return
		}
	} else {
		for _, socket := range sockets {
			device.Exporter.HandlePowerStripSocket(context.Background(), device.Host, device.Name, socket)
		}
	}
}
