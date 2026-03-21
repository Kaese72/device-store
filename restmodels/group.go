package restmodels

import "time"

type Group struct {
	ID               int               `json:"id"`
	Name             string            `json:"name"`
	AdapterId        int               `json:"adapter-id"`
	BridgeIdentifier string            `json:"bridge-identifier"`
	Updated          time.Time         `json:"updated"`
	Capabilities     []GroupCapability `json:"capabilities"`
	DeviceIds        []int             `json:"device-ids"`
}
