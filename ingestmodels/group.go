package ingestmodels

type IngestGroup struct {
	Name             string                  `json:"name"`
	BridgeKey        string                  `json:"bridge-key" required:"false"`
	BridgeIdentifier string                  `json:"bridge-identifier"`
	Capabilities     []IngestGroupCapability `json:"capabilities"`
	DeviceIds        []int                   `json:"device-ids"`
}
