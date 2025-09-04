package ingestmodels

type IngestGroup struct {
	Name             string                  `json:"name"`
	BridgeKey        string                  `json:"bridge-key"`
	BridgeIdentifier string                  `json:"bridge-identifier"`
	Capabilities     []IngestGroupCapability `json:"capabilities"`
	DeviceIds        []int                   `json:"device-ids"`
}
