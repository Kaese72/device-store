package restmodels

type BooleanArgumentSpec struct {
	Default *bool `json:"default,omitempty"`
}

type NumericArgumentSpec struct {
	Min     float32  `json:"min"`
	Max     float32  `json:"max"`
	Default *float32 `json:"default,omitempty"`
}

type TextArgumentSpec struct {
	Default *string `json:"default,omitempty"`
	Min     *int    `json:"min,omitempty"`
	Max     *int    `json:"max,omitempty"`
}

type ArgumentSpec struct {
	Name    string               `json:"name"`
	Boolean *BooleanArgumentSpec `json:"boolean,omitempty"`
	Numeric *NumericArgumentSpec `json:"numeric,omitempty"`
	Text    *TextArgumentSpec    `json:"text,omitempty"`
}

// func (i ArgumentSpec) Resolve(ctx huma.Context, prefix *huma.PathBuffer) []error {
// 	// Only one of Boolean, Numeric, or Text may be set
// 	count := 0
// 	if i.Boolean != nil {
// 		count++
// 	}
// 	if i.Numeric != nil {
// 		count++
// 	}
// 	if i.Text != nil {
// 		count++
// 	}
// 	if count != 1 {
// 		return []error{huma.Error400BadRequest("Argument spec must have one and only one of boolean, numeric, or text defined")}
// 	}
// 	return nil
// }

// // Make sure that we fulfil the correct interface for validation to work
// var _ huma.ResolverWithPath = (*ArgumentSpec)(nil)
