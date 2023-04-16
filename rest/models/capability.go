package models

import (
	"time"
)

type Capability struct {
	// LastSpecUpdated reflects when the capability was last updated.
	// Read only from the device-store perspective
	LastSeen time.Time `json:"last-spec-updated,omitempty"`
}

type CapabilityArgs map[string]interface{}
