package eventmodels

import "time"

type UpdatedAttribute struct {
	Name    string   `json:"name"`
	Boolean *bool    `json:"boolean-state,omitempty"`
	Numeric *float32 `json:"numeric-state,omitempty"`
	Text    *string  `json:"string-state,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

type DeviceAttributeUpdate struct {
	DeviceID   int                `json:"device-id"`
	Attributes []UpdatedAttribute `json:"attributes"`
}
