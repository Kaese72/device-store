package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type DeviceIdsList []int

func (d *DeviceIdsList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &d)
}

type GroupIntermediary struct {
	// The device store unique identifier
	ID int `db:"id,omitempty"`
	// Identifier from bridge perspective
	BridgeIdentifier string `db:"bridgeIdentifier"`
	// "Key" of the owning bridge
	BridgeKey string `db:"bridgeKey"`
	// The name of the group
	Name string `db:"name"`
	// The capabilities of the group
	Capabilities GroupCapabilityIntermediaryList `db:"capabilities,omitempty"`
	// The device IDs in the group
	DeviceIds DeviceIdsList `db:"deviceIds,omitempty"`
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
	"bridge-key": {
		"eq": func(value string) (string, []string) {
			return "bridgeKey = ?", []string{value}
		},
	},
}

func (g *GroupIntermediary) ToRestModel() models.Group {
	return models.Group{
		ID:               g.ID,
		BridgeIdentifier: g.BridgeIdentifier,
		BridgeKey:        g.BridgeKey,
		Name:             g.Name,
		Capabilities:     g.Capabilities.ToRestModel(),
		DeviceIds:        g.DeviceIds,
	}
}

func GroupIntermediaryFromRest(group models.Group) GroupIntermediary {
	return GroupIntermediary{
		ID:               group.ID,
		BridgeIdentifier: group.BridgeIdentifier,
		BridgeKey:        group.BridgeKey,
		Name:             group.Name,
		Capabilities:     GroupCapabilityIntermediaryListFromRest(group.Capabilities),
		DeviceIds:        group.DeviceIds,
	}
}
