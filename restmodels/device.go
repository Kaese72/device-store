package restmodels

type Device struct {
	ID               int                `json:"id"`
	BridgeIdentifier string             `json:"bridge-identifier"`
	BridgeKey        string             `json:"bridge-key"`
	Attributes       []Attribute        `json:"attributes"`
	Triggers         []Trigger          `json:"triggers"`
	Capabilities     []DeviceCapability `json:"capabilities"`
	GroupIds         []int              `json:"group-ids"`
}
