package ingestmodels

type IngestDeviceCapability struct {
	Name          string               `json:"name"`
	ArgumentSpecs []IngestArgumentSpec `json:"argument-specs"`
}

type IngestDeviceCapabilityArgs map[string]any
