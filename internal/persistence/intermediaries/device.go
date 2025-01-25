package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type GroupIdsList []int

func (g *GroupIdsList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &g)
}

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
	// The groups this device is a member of
	GroupIds GroupIdsList `db:"groupIds,omitempty"`
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
		GroupIds:         d.GroupIds,
	}
}

func DeviceIntermediaryFromRest(device models.Device) DeviceIntermediary {
	return DeviceIntermediary{
		BridgeIdentifier: device.BridgeIdentifier,
		BridgeKey:        device.BridgeKey,
		ID:               device.ID,
		Attributes:       AttributeIntermediaryListFromRest(device.Attributes),
		Capabilities:     DeviceCapabilityIntermediaryListFromRest(device.Capabilities),
		GroupIds:         device.GroupIds,
	}
}
