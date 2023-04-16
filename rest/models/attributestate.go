package models

type AttributeState struct {
	Boolean *bool    `json:"boolean-state"`
	Numeric *float32 `json:"numeric-state"`
	Text    *string  `json:"string-state"`
}

func exclusiveNil(pointer1, pointer2 interface{}) bool {
	return (pointer1 == nil) != (pointer2 == nil)
}

// Equal checks whether two states are equal.
func (state AttributeState) Equal(other AttributeState) bool {
	if exclusiveNil(state.Boolean, other.Boolean) ||
		exclusiveNil(state.Numeric, other.Numeric) ||
		exclusiveNil(state.Text, other.Text) {
		// At least one of the states is set in one but not the other
		return false
	}
	//FIXME If none are set, this will evaluate to false while it should probably evaluate to true
	if (state.Boolean != nil && (*state.Boolean == *other.Boolean)) ||
		(state.Numeric != nil && (*state.Numeric == *other.Numeric)) ||
		(state.Text != nil && (*state.Text == *other.Text)) {

		return true
	}
	return false
}
