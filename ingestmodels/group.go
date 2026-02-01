package ingestmodels

type IngestGroup struct {
	Name             string                  `json:"name" required:"true"`
	AdapterId        int                     `json:"adapter-id" readonly:"true"`
	BridgeIdentifier string                  `json:"bridge-identifier" required:"true"`
	Capabilities     []IngestGroupCapability `json:"capabilities" required:"false"`
	DeviceIds        []int                   `json:"device-ids" required:"false"`
}
