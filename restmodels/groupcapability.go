package restmodels

type GroupCapability struct {
	Name          string         `json:"name"`
	ArgumentSpecs []ArgumentSpec `json:"argument-specs"`
}

type GroupCapabilityArgs map[string]any
