package ingestmodels

type Device struct {
	BridgeIdentifier string             `json:"bridge-identifier"`
	BridgeKey        string             `json:"bridge-key"`
	Attributes       []Attribute        `json:"attributes"`
	Capabilities     []DeviceCapability `json:"capabilities"`
	GroupIds         []int              `json:"group-ids"`
}
