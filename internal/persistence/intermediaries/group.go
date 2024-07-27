package intermediaries

import "github.com/Kaese72/device-store/rest/models"

type GroupIntermediary struct {
	// The device store unique identifier
	ID int `db:"id,omitempty"`
	// Identifier from bridge perspective
	BridgeIdentifier string `db:"bridgeIdentifier"`
	// "Key" of the owning bridge
	BridgeKey string `db:"bridgeKey"`
	// FIXME Database scan likely fails here
	Capabilities GroupCapabilityIntermediaryList `db:"capabilities,omitempty"`
}

var GroupFilters = map[string]map[string]func(string) (string, []string){
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

func (d *GroupIntermediary) ToRestModel() models.Group {
	return models.Group{
		ID:               d.ID,
		BridgeIdentifier: d.BridgeIdentifier,
		BridgeKey:        d.BridgeKey,
		Capabilities:     d.Capabilities.ToRestModel(),
	}
}

func GroupIntermediaryFromRest(device models.Group) GroupIntermediary {
	return GroupIntermediary{
		BridgeIdentifier: device.BridgeIdentifier,
		BridgeKey:        device.BridgeKey,
		ID:               device.ID,
		Capabilities:     GroupCapabilityIntermediaryListFromRest(device.Capabilities),
	}
}
