package restmodels

import "time"

type CapabilityTriggerAudit struct {
	ID           int       `json:"id"`
	DeviceID     int       `json:"device-id"`
	Name         string    `json:"name"`
	Success      bool      `json:"success"`
	ErrorMessage *string   `json:"error-message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Arguments    string    `json:"arguments"`
}
