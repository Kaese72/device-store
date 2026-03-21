package restmodels

import "time"

type DeviceCapability struct {
	Name          string         `json:"name"`
	ArgumentSpecs []ArgumentSpec `json:"argument-specs"`
	Updated       time.Time      `json:"updated"`
}

type DeviceCapabilityArgs map[string]any
