package main

import (
	"io"
	"log/slog"
	"os"
	"time"
)

const (
	tapoEmail            = "TAPO_EMAIL"
	tapoPassword         = "TAPO_PASSWORD"
	tapoConfigLocation   = "TAPO_CONFIG_LOCATION"
	prometheusPort       = "8086"
	metricPrefix         = "tapo"
	fetchIntervalSeconds = 30
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

	config := PrometheusConfig{
		ServerPort: prometheusPort,
		Prefix:     metricPrefix,
		Devices:    devices,
	}
	exporter := NewPrometheusExporter(&config)

	initSmartPlugs(devices.SmartPlugs, username, password, exporter)
	initTSeries(devices.TSeries, username, password, exporter)
	initPowerStrips(devices.PowerStrips, username, password, exporter)

	ticker := time.NewTicker(fetchIntervalSeconds * time.Second)
	for _ = time.Now(); ; _ = <-ticker.C {
		for _, device := range devices.SmartPlugs {
			go handleSmartPlug(device)
		}
		for _, device := range devices.TSeries {
			go handleTSeries(device)
		}
		for _, device := range devices.PowerStrips {
			go handlePowerStrip(device)
		}
	}
}

func getDevices(configLocation string) (*Devices, error) {
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
