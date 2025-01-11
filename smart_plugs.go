package main

import (
	"context"
	"encoding/json"
	"github.com/tess1o/tapo-go"
	"log/slog"
	"time"
)

func initSmartPlugs(devices []SmartPlug, username string, password string) {
	for i := range devices {
		client, err := initSmartPlugClient(devices[i].Host, username, password)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "device", devices[i].Name)
			continue
		}
		devices[i].Client = client
	}
}

func initSmartPlugClient(host, username, password string) (*tapo.SmartPlug, error) {
	return tapo.NewSmartPlug(host, username, password, tapo.Options{})
}

func handleSmartPlug(device SmartPlug, username string, password string, exporter *PrometheusExporter) {
	if device.Client == nil {
		client, err := initSmartPlugClient(device.Host, username, password)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "device", device.Name)
			return
		}
		device.Client = client
	}
	r, err := getEnergyUsage(device)
	if err != nil {
		slog.Error("Error getting energy usage", "device", device.Name, "error", err)
	} else {
		slog.Info("Successfully received metrics", "device", device.Name)
		d, _ := json.Marshal(r.Result)
		var params map[string]interface{}
		json.Unmarshal(d, &params)
		exporter.HandleSmartPlug(context.Background(), device, params)
	}
}

func getEnergyUsage(device SmartPlug) (*tapo.EnergyUsageResponse, error) {
	var energyUsageResponse *tapo.EnergyUsageResponse
	var energyError error
	for i := 0; i < maxRetries; i++ {
		r, err := device.Client.GetEnergyUsage()
		if err == nil {
			energyUsageResponse = r
			break
		} else {
			energyError = err
			slog.Error("Error getting energy usage", "attempt", i+1, "device", device.Name, "error", energyError)
			time.Sleep(time.Second * delayBetweenRetries)
		}
	}

	return energyUsageResponse, energyError
}
