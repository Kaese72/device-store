package restmodels

type GroupCapability struct {
	Name                string              `json:"name"`
	ArgumentsJsonSchema ArgumentsJsonSchema `json:"arguments-json-schema"`
}

type GroupCapabilityArgs map[string]any
