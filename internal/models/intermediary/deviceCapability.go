package models

import (
	"time"
)

type CapabilityIntermediary struct {
	DeviceId            string
	CapabilityName      string
	CapabilityBridgeKey string
	LastSeen            time.Time
}
