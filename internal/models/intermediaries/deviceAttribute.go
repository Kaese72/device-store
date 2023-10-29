package intermediaries

type AttributeStateIntermediary struct {
	Boolean *bool
	Numeric *float32
	Text    *string
}

type AttributeIntermediary struct {
	DeviceId string
	Name     string
	State    AttributeStateIntermediary
}
