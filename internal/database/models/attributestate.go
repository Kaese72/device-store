package models

import devicestoretemplates "github.com/Kaese72/device-store/rest/models"

type AttributeState struct {
	Boolean *bool    `bson:"boolean-state"`
	Numeric *float32 `bson:"numeric-state"`
	Text    *string  `bson:"string-state"`
}

func (attr AttributeState) ConvertToAPIAttributeState() devicestoretemplates.AttributeState {
	return devicestoretemplates.AttributeState{
		Boolean: attr.Boolean,
		Numeric: attr.Numeric,
		Text:    attr.Text,
	}
}

func NewAttributeStateFromAPIAttributeState(attr devicestoretemplates.AttributeState) AttributeState {
	return AttributeState{
		Boolean: attr.Boolean,
		Numeric: attr.Numeric,
		Text:    attr.Text,
	}
}
