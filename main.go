package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/tess1o/go-tapo/api/types"
	"github.com/tess1o/go-tapo/pkg/tapogo"
)

const (
	tapoEmail            = "TAPO_EMAIL"
	tapoPassword         = "TAPO_PASSWORD"
	tapoConfigLocation   = "TAPO_CONFIG_LOCATION"
	prometheusPort       = "8086"
	metricPrefix         = "tapo"
	fetchIntervalSeconds = 3
	maxRetries           = 5
	delayBetweenRetries  = 2
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelInfo)
	username := os.Getenv(tapoEmail)
	password := os.Getenv(tapoPassword)
	configLocation := os.Getenv(tapoConfigLocation)

	if username == "" || password == "" || configLocation == "" {
		slog.Error("TAPO_USERNAME, TAPO_PASSWORD and TAPO_CONFIG_LOCATION must be set")
		return
	}

	devices, err := getDevices(configLocation)
	if err != nil {
		slog.Error("Error getting devices", "error", err)
		return
	}
	if len(devices) == 0 {
		slog.Error("No devices found")
		return
	}

	initTapiClients(devices, username, password)

	config := PrometheusConfig{
		ServerPort: prometheusPort,
		Prefix:     metricPrefix,
		Devices:    devices,
	}
	exporter := NewPrometheusExporter(&config)

	ticker := time.NewTicker(fetchIntervalSeconds * time.Second)
	for _ = time.Now(); ; _ = <-ticker.C {
		for _, device := range devices {
			go handleDevice(device, username, password, exporter)
		}
	}
}

func getDevices(configLocation string) ([]Device, error) {
	configFile, err := os.Open(configLocation)
	if err != nil {
		slog.Error("Error opening config file: ", err)
		return nil, err
	}

	defer configFile.Close()

	configData, err := io.ReadAll(configFile)
	if err != nil {
		slog.Error("Error reading config file: ", err)
		return nil, err
	}

	devices, err := ReadDevices(configData)
	if err != nil {
		slog.Error("Error reading devices: ", err)
		return nil, err
	}
	slog.Info("Devices from config", "devices", devices)
	return devices, nil
}

func initTapiClients(devices []Device, username string, password string) {
	for i := range devices {
		client, err := initClient(devices[i].IPAddress, username, password)
		if err != nil {
			slog.Error("Failed to create Tapo client", "error", err, "device", devices[i].Name)
			continue
		}
		devices[i].Client = client
	}
}

func initClient(host, username, password string) (*tapogo.Tapo, error) {
	return tapogo.NewTapo(host, username, password, &tapogo.TapoOptions{})
}

func handleDevice(device Device, username string, password string, exporter *PrometheusExporter) {
	if device.Client == nil {
		client, err := initClient(device.IPAddress, username, password)
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
		exporter.Handle(context.Background(), device, params)
	}
}

func getEnergyUsage(device Device) (*types.ResponseSpec, error) {
	var energyUsageResponse *types.ResponseSpec
	var energyError error
	for i := 0; i < maxRetries; i++ {
		r, err := device.Client.GetEnergyUsage()
		if err == nil {
			energyUsageResponse = r
			break
		} else {
			energyError = err
			slog.Error("Error getting energy usage", "attempt", i+1, "device", device.Name, "error", energyError)

			if strings.Contains(energyError.Error(), "403") {
				slog.Info("Received 403 error, trying to re-authenticate", "device", device.Name)
				time.Sleep(time.Second * 2)
				err = device.Client.Handshake()
				if err != nil {
					slog.Error("Failed to re-authenticate", "error", err, "device", device.Name)
				}
			}
			time.Sleep(time.Second * delayBetweenRetries)
		}
	}

	return energyUsageResponse, energyError
}
