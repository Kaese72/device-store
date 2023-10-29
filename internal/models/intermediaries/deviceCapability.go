package intermediaries

import (
	"time"
)

type CapabilityIntermediary struct {
	StoreDeviceIdentifier  string
	BridgeDeviceIdentifier string
	BridgeKey              string
	Name                   string
	LastSeen               time.Time
}
