package ingestmodels

type Device struct {
	BridgeIdentifier string             `json:"bridge-identifier"`
	BridgeKey        string             `json:"bridge-key"`
	Attributes       []Attribute        `json:"attributes"`
	Capabilities     []DeviceCapability `json:"capabilities"`
	Triggers         []Trigger          `json:"triggers"`
	GroupIds         []int              `json:"group-ids"`
}
