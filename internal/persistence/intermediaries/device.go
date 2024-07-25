package intermediaries

import (
	"github.com/Kaese72/device-store/rest/models"
)

type DeviceIntermediary struct {
	// Identifier from bridge perspective
	BridgeIdentifier string `db:"bridgeIdentifier"`
	// "Key" of the owning bridge
	BridgeKey string `db:"bridgeKey"`
	// The device store unique identifier
	DeviceStoreIdentifier int `db:"id,omitempty"`
	// FIXME Database scan likely fails here
	Attributes   AttributeIntermediaryList  `db:"attributes,omitempty"`
	Capabilities CapabilityIntermediaryList `db:"capabilities,omitempty"`
}

var DeviceFilters = map[string]map[string]func(string) (string, []string){
	"bridge-identifier": {
		"eq": func(value string) (string, []string) {
			return "bridgeIdentifier = ?", []string{value}
		},
	},
	"store-identifier": {
		"eq": func(value string) (string, []string) {
			return "id = ?", []string{value}
		},
	},
}

func (d *DeviceIntermediary) ToRestModel() models.Device {
	return models.Device{
		BridgeIdentifier: d.BridgeIdentifier,
		BridgeKey:        d.BridgeKey,
		StoreIdentifier:  d.DeviceStoreIdentifier,
		Attributes:       d.Attributes.ToRestModel(),
		Capabilities:     d.Capabilities.ToRestModel(),
	}
}

func DeviceIntermediaryFromRest(device models.Device) DeviceIntermediary {
	return DeviceIntermediary{
		BridgeIdentifier:      device.BridgeIdentifier,
		BridgeKey:             device.BridgeKey,
		DeviceStoreIdentifier: device.StoreIdentifier,
		Attributes:            AttributeIntermediaryListFromRest(device.Attributes),
		Capabilities:          CapabilityIntermediaryListFromRest(device.Capabilities),
	}
}
