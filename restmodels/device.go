package restmodels

type Device struct {
	ID               int                `json:"id"`
	BridgeIdentifier string             `json:"bridge-identifier"`
	AdapterId        int                `json:"adapter-id"`
	Updated          string             `json:"updated"`
	Attributes       []Attribute        `json:"attributes"`
	Capabilities     []DeviceCapability `json:"capabilities"`
	GroupIds         []int              `json:"group-ids"`
}
