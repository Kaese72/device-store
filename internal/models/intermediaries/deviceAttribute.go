package intermediaries

type AttributeStateIntermediary struct {
	Boolean *bool
	Numeric *float32
	Text    *string
}

type AttributeIntermediary struct {
	DeviceStoreIdentifier string
	Name                  string
	State                 AttributeStateIntermediary
}
