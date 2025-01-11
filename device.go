package main

import (
	"encoding/json"
	"github.com/tess1o/tapo-go"
)

type Devices struct {
	SmartPlugs []SmartPlug `json:"smart_plugs"`
	TSeries    []TSeries   `json:"t_series"`
}

type SmartPlug struct {
	Name   string          `json:"name"`
	Host   string          `json:"host"`
	Client *tapo.SmartPlug `json:"-"`
}

type TSeries struct {
	HubHost string        `json:"hub_host"`
	Client  *tapo.TSeries `json:"-"`
}

func ReadDevices(configData []byte) (*Devices, error) {
	var devices *Devices
	err := json.Unmarshal(configData, &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
