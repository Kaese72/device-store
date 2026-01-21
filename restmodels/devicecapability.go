package restmodels

type DeviceCapability struct {
	Name          string         `json:"name"`
	ArgumentSpecs []ArgumentSpec `json:"argument-specs"`
}

type DeviceCapabilityArgs map[string]any
