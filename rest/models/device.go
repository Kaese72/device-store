package models

type Device struct {
	ID               int                `json:"id"`
	BridgeIdentifier string             `json:"bridge-identifier"`
	BridgeKey        string             `json:"bridge-key"`
	Attributes       []Attribute        `json:"attributes"`
	Capabilities     []DeviceCapability `json:"capabilities"`
}
