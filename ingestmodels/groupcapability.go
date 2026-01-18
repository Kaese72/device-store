package ingestmodels

type IngestGroupCapability struct {
	Name          string               `json:"name"`
	ArgumentSpecs []IngestArgumentSpec `json:"argument-specs"`
}

type IngestGroupCapabilityArgs map[string]any
