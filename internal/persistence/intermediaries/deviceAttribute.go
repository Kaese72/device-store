package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type AttributeIntermediary struct {
	Name     string   `db:"name"`
	DeviceId string   `db:"deviceId"`
	BooleanX *int     `db:"boolean"`
	Numeric  *float32 `db:"numeric"`
	Text     *string  `db:"text"`
}

func (a *AttributeIntermediary) Boolean() *bool {
	if a.BooleanX == nil {
		return nil
	}
	b := *a.BooleanX != 0
	return &b
}

func (a *AttributeIntermediary) ToRestModel() models.Attribute {
	return models.Attribute{
		Name:    a.Name,
		Boolean: a.Boolean(),
		Numeric: a.Numeric,
		Text:    a.Text,
	}
}

func AttributeIntermediaryFromRest(attr models.Attribute) AttributeIntermediary {
	var boolean *int = nil
	if attr.Boolean != nil {
		if *attr.Boolean {
			boolean = &[]int{1}[0]
		} else {
			boolean = &[]int{0}[0]
		}
	}
	return AttributeIntermediary{
		Name:     attr.Name,
		BooleanX: boolean,
		Numeric:  attr.Numeric,
		Text:     attr.Text,
	}
}

type AttributeIntermediaryList []AttributeIntermediary

func (attrs *AttributeIntermediaryList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &attrs)
}

func (attrs AttributeIntermediaryList) ToRestModel() []models.Attribute {
	restAttrs := []models.Attribute{}
	for _, attr := range attrs {
		restAttrs = append(restAttrs, attr.ToRestModel())
	}
	return restAttrs
}

func AttributeIntermediaryListFromRest(attrs []models.Attribute) AttributeIntermediaryList {
	intermediaries := AttributeIntermediaryList{}
	for _, attr := range attrs {
		intermediaries = append(intermediaries, AttributeIntermediaryFromRest(attr))
	}
	return intermediaries
}
