package restmodels

import "time"

type GroupCapabilityTriggerAudit struct {
	ID           int       `json:"id"`
	GroupID      int       `json:"group-id"`
	Name         string    `json:"name"`
	Success      bool      `json:"success"`
	ErrorMessage *string   `json:"error-message,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Arguments    string    `json:"arguments"`
}
