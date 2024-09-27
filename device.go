package main

import (
	"encoding/json"
	"github.com/tess1o/go-tapo/pkg/tapogo"
)

type Config struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	Name      string       `json:"name"`
	IPAddress string       `json:"ip_address"`
	Client    *tapogo.Tapo `json:"-"`
}

func ReadDevices(configData []byte) ([]Device, error) {
	var config Config
	err := json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	return config.Devices, nil
}
