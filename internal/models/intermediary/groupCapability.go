package models

import (
	"time"
)

type GroupCapabilityIntermediary struct {
	GroupId             string
	CapabilityName      string
	CapabilityBridgeKey string
	LastSeen            time.Time
}
