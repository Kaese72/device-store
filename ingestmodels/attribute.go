package ingestmodels

import "time"

type IngestAttribute struct {
	Name    string    `json:"name"`
	Boolean *bool     `json:"boolean-state,omitempty"`
	Numeric *float32  `json:"numeric-state,omitempty"`
	Text    *string   `json:"string-state,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}
