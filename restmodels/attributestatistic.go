package restmodels

// AttributeStatistic describes, for a given attribute name, how many
// devices store that attribute's value under each of the three possible
// types.
type AttributeStatistic struct {
	Name     string `json:"name"`
	NBoolean int    `json:"n-boolean"`
	NText    int    `json:"n-text"`
	NNumeric int    `json:"n-numeric"`
}
