package ingestmodels

type IngestArgumentProperty struct {
	Type    string `json:"type"` // "integer", "string", "boolean"
	Default any    `json:"default,omitempty"`
}

// ArgumentsJsonSchema is the top level structure for capability arguments
type IngestArgumentsJsonSchema struct {
	Properties map[string]IngestArgumentProperty `json:"properties"`
}
