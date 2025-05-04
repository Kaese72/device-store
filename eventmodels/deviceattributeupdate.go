package eventmodels

type Attribute struct {
	Name    string   `json:"name"`
	Boolean *bool    `json:"boolean-state,omitempty"`
	Numeric *float32 `json:"numeric-state,omitempty"`
	Text    *string  `json:"string-state,omitempty"`
}

type DeviceAttributeUpdate struct {
	DeviceID   int         `json:"device-id"`
	Attributes []Attribute `json:"attributes"`
}
