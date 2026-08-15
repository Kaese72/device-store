package restmodels

import "time"

type Device struct {
	ID               int                `json:"id"`
	Name             string             `json:"name"`
	BridgeIdentifier string             `json:"bridge-identifier"`
	AdapterId        int                `json:"adapter-id"`
	Updated          time.Time          `json:"updated"`
	Attributes       []Attribute        `json:"attributes"`
	Capabilities     []DeviceCapability `json:"capabilities"`
	GroupIds         []int              `json:"group-ids"`
}
