package ingestmodels

type IngestGroupCapability struct {
	Name                string                    `json:"name"`
	ArgumentsJsonSchema IngestArgumentsJsonSchema `json:"arguments-json-schema"`
}

type IngestGroupCapabilityArgs map[string]any
