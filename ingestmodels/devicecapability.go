package ingestmodels

type IngestDeviceCapability struct {
	Name                string                    `json:"name"`
	ArgumentsJsonSchema IngestArgumentsJsonSchema `json:"arguments-json-schema"`
}

type IngestDeviceCapabilityArgs map[string]any
