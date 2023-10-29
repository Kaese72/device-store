package intermediaries

type DeviceIntermediary struct {
	BridgeIdentifier      string // Identifier from bridge perspective
	BridgeKey             string // "Key" of the owning bridge
	DeviceStoreIdentifier string // The device store unique identifier
}
