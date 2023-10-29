package intermediaries

import (
	"time"
)

type CapabilityIntermediary struct {
	DeviceId  string
	Name      string
	BridgeKey string
	LastSeen  time.Time
}
