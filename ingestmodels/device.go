package ingestmodels

type IngestDevice struct {
	BridgeIdentifier string                   `json:"bridge-identifier" required:"true"`
	Attributes       []IngestAttribute        `json:"attributes" required:"false"`
	Capabilities     []IngestDeviceCapability `json:"capabilities" required:"false"`
	GroupIds         []int                    `json:"group-ids" required:"false"`
	// The AdapterId is set from the JWT token and is not expected from the client
	AdapterId int `json:"-"`
}
