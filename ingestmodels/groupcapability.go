package ingestmodels

type IngestGroupCapability struct {
	Name                string                    `json:"name"`
	ArgumentsJsonSchema IngestArgumentsJsonSchema `json:"arguments-json-schema"`
}

type IngestAttributeGroupCapabilityArgs map[string]any
