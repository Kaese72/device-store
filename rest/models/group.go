package models

type Group struct {
	Identifier   string       `json:"identifier"`
	Name         string       `json:"name"`
	Capabilities []Capability `json:"capabilities"`
}
