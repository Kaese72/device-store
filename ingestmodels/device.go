package ingestmodels

type IngestDevice struct {
	BridgeIdentifier string                   `json:"bridge-identifier"`
	BridgeKey        string                   `json:"bridge-key"`
	Attributes       []IngestAttribute        `json:"attributes"`
	Capabilities     []IngestDeviceCapability `json:"capabilities"`
	Triggers         []IngestTrigger          `json:"triggers"`
	GroupIds         []int                    `json:"group-ids"`
}
