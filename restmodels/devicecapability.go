package restmodels

type DeviceCapability struct {
	Name          string         `json:"name"`
	ArgumentSpecs []ArgumentSpec `json:"argument-specs"`
	Updated       string         `json:"updated"`
}

type DeviceCapabilityArgs map[string]any
