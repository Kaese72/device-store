package ingestmodels

type IngestDevice struct {
	BridgeIdentifier string                   `json:"bridge-identifier"`
	BridgeKey        string                   `json:"bridge-key" required:"false"`
	Attributes       []IngestAttribute        `json:"attributes", required:"false"`
	Capabilities     []IngestDeviceCapability `json:"capabilities" required:"false"`
	Triggers         []IngestTrigger          `json:"triggers" required:"false"`
	GroupIds         []int                    `json:"group-ids" required:"false"`
}
