package intermediaries

type GroupCapabilityIntermediaryActivation struct {
	// BridgeIdentifier is an encoded string that contains the information
	// needed to identify a group on the Adapter (/Bridge). It generally
	// contains information about what type and unique ID the group has on the Adapter
	// For example, for Phillips Hue, this might be something like "groups/34"
	BridgeIdentifier string
	// Name is the Adapter (And store) name of the capability.
	Name string
	// AdapterId is the ID of the adapter this capability is tied to.
	AdapterId int
	// "Bridge" is the old name for "Adapter". The Key used to identiy which adapter/bridge to use
	// BridgeKey        string
}
