package intermediaries

import (
	"github.com/Kaese72/device-store/rest/models"
)

type DeviceIntermediary struct {
	// The device store unique identifier
	ID int `db:"id,omitempty"`
	// Identifier from bridge perspective
	BridgeIdentifier string `db:"bridgeIdentifier"`
	// "Key" of the owning bridge
	BridgeKey string `db:"bridgeKey"`
	// The attributes of the device
	Attributes AttributeIntermediaryList `db:"attributes,omitempty"`
	// The capabilities of the device
	Capabilities DeviceCapabilityIntermediaryList `db:"capabilities,omitempty"`
}

var DeviceFilters = map[string]map[string]func(string) (string, []string){
	"bridge-identifier": {
		"eq": func(value string) (string, []string) {
			return "bridgeIdentifier = ?", []string{value}
		},
	},
	"id": {
		"eq": func(value string) (string, []string) {
			return "id = ?", []string{value}
		},
	},
}

func (d *DeviceIntermediary) ToRestModel() models.Device {
	return models.Device{
		ID:               d.ID,
		BridgeIdentifier: d.BridgeIdentifier,
		BridgeKey:        d.BridgeKey,
		Attributes:       d.Attributes.ToRestModel(),
		Capabilities:     d.Capabilities.ToRestModel(),
	}
}

func DeviceIntermediaryFromRest(device models.Device) DeviceIntermediary {
	return DeviceIntermediary{
		BridgeIdentifier: device.BridgeIdentifier,
		BridgeKey:        device.BridgeKey,
		ID:               device.ID,
		Attributes:       AttributeIntermediaryListFromRest(device.Attributes),
		Capabilities:     DeviceCapabilityIntermediaryListFromRest(device.Capabilities),
	}
}
