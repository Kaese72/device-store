package restmodels

type Group struct {
	ID               int               `json:"id"`
	Name             string            `json:"name"`
	BridgeKey        string            `json:"bridge-key"`
	BridgeIdentifier string            `json:"bridge-identifier"`
	Capabilities     []GroupCapability `json:"capabilities"`
	DeviceIds        []int             `json:"device-ids"`
}
