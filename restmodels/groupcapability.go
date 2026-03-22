package restmodels

import "time"

type GroupCapability struct {
	Name          string         `json:"name"`
	ArgumentSpecs []ArgumentSpec `json:"argument-specs"`
	Updated       time.Time      `json:"updated"`
}

type GroupCapabilityArgs map[string]any
