package models

type Group struct {
	Identifier   string                       `json:"identifier"`
	Name         string                       `json:"name"`
	Capabilities map[CapabilityKey]Capability `json:"capabilities"`
}
