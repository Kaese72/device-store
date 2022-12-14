package models

import (
	"time"

	"github.com/Kaese72/sdup-lib/devicestoretemplates"
)

type CapabilityIntermediary struct {
	DeviceId            string
	CapabilityName      string
	CapabilityBridgeKey devicestoretemplates.BridgeKey
	LastSeen            time.Time
}
