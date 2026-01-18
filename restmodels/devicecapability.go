package restmodels

type DeviceCapability struct {
	Name                string              `json:"name"`
	ArgumentsJsonSchema ArgumentsJsonSchema `json:"arguments-json-schema"`
}

type DeviceCapabilityArgs map[string]any
