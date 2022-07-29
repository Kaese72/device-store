package models

import (
	"github.com/Kaese72/sdup-lib/devicestoretemplates"
)

type Bridge struct {
	Identifier devicestoretemplates.BridgeKey `bson:"identifier"`
	URI        string                         `bson:"uri"`
}

func (bridge Bridge) ConvertToAPIBridge() devicestoretemplates.Bridge {
	return devicestoretemplates.Bridge{
		Identifier: bridge.Identifier,
		URI:        bridge.URI,
	}
}

func NewBridgeFromAPIBridge(bridge devicestoretemplates.Bridge) Bridge {
	return Bridge{
		Identifier: bridge.Identifier,
		URI:        bridge.URI,
	}
}
