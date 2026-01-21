package ingestmodels

import "github.com/danielgtaylor/huma/v2"

type IngestBooleanArgumentSpec struct {
	Default *bool `json:"default,omitempty"`
}

type IngestNumericArgumentSpec struct {
	Min     float32  `json:"min"`
	Max     float32  `json:"max"`
	Default *float32 `json:"default,omitempty"`
}

type IngestTextArgumentSpec struct {
	Default *string `json:"default,omitempty"`
	Min     *int    `json:"min,omitempty"`
	Max     *int    `json:"max,omitempty"`
}

type IngestArgumentSpec struct {
	Name    string                     `json:"name"`
	Boolean *IngestBooleanArgumentSpec `json:"boolean,omitempty"`
	Numeric *IngestNumericArgumentSpec `json:"numeric,omitempty"`
	Text    *IngestTextArgumentSpec    `json:"text,omitempty"`
}

func (i IngestArgumentSpec) Resolve(ctx huma.Context, prefix *huma.PathBuffer) []error {
	// Only one of Boolean, Numeric, or Text may be set
	count := 0
	if i.Boolean != nil {
		count++
	}
	if i.Numeric != nil {
		count++
	}
	if i.Text != nil {
		count++
	}
	if count != 1 {
		return []error{huma.Error400BadRequest("Argument spec must have one and only one of boolean, numeric, or text defined")}
	}
	return nil
}

// Make sure that we fulfil the correct interface for validation to work
var _ huma.ResolverWithPath = (*IngestArgumentSpec)(nil)
