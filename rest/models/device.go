package models

type Device struct {
	BridgeIdentifier string       `json:"bridge-identifier"`
	BridgeKey        string       `json:"bridge-key"`
	StoreIdentifier  int          `json:"store-identifier"`
	Attributes       []Attribute  `json:"attributes"`
	Capabilities     []Capability `json:"capabilities"`
}
