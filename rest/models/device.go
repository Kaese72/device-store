package models

//AttributeKey is the string identifier of an attribute
type AttributeKey string

type CapabilityKey string

type Device struct {
	Identifier      string                          `json:"identifier"`
	BridgeKey       string                          `json:"bridge-key"`
	StoreIdentifier string                          `json:"store-identifier"`
	Attributes      map[AttributeKey]AttributeState `json:"attributes"`
	Capabilities    map[CapabilityKey]Capability    `json:"capabilities"`
}
