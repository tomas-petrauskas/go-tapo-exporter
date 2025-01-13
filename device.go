package main

import (
	"encoding/json"
)

type Devices struct {
	SmartPlugs []*SmartPlug `json:"smart_plugs"`
	TSeries    []*TSeries   `json:"t_series"`
}

func ReadDevices(configData []byte) (*Devices, error) {
	var devices *Devices
	err := json.Unmarshal(configData, &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}
