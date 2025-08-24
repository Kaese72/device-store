package restmodels

import "time"

type AttributeAudit struct {
	ID              int       `json:"id"`
	DeviceID        int       `json:"deviceId"`
	Name            string    `json:"name"`
	Timestamp       time.Time `json:"timestamp"`
	OldBooleanValue *bool     `json:"oldBooleanValue"`
	OldNumericValue *float64  `json:"oldNumericValue"`
	OldTextValue    *string   `json:"oldTextValue"`
	NewBooleanValue *bool     `json:"newBooleanValue"`
	NewNumericValue *float64  `json:"newNumericValue"`
	NewTextValue    *string   `json:"newTextValue"`
}
