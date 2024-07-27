package intermediaries

type GroupCapabilityIntermediaryActivation struct {
	BridgeIdentifier string `db:"bridgeIdentifier"`
	Name             string `db:"name"`
	BridgeKey        string `db:"bridgeKey"`
}
