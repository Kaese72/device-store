package restmodels

type ArgumentProperty struct {
	Type    string `json:"type"` // "integer", "string", "boolean"
	Default any    `json:"default,omitempty"`
}

// ArgumentsJsonSchema is the top level structure for capability arguments
type ArgumentsJsonSchema struct {
	Properties map[string]ArgumentProperty `json:"properties"`
}
