package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type CapabilityIntermediary struct {
	DeviceId string `json:"DeviceId" db:"DeviceId"`
	Name     string `json:"name" db:"name"`
}

func (c *CapabilityIntermediary) ToRestModel() models.Capability {
	return models.Capability{
		Name: c.Name,
	}
}

func CapabilityIntermediaryFromRest(cap models.Capability) CapabilityIntermediary {
	return CapabilityIntermediary{
		Name: cap.Name,
	}
}

type CapabilityIntermediaryList []CapabilityIntermediary

func (caps *CapabilityIntermediaryList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &caps)
}

func (caps CapabilityIntermediaryList) ToRestModel() []models.Capability {
	restCaps := []models.Capability{}
	for _, cap := range caps {
		restCaps = append(restCaps, cap.ToRestModel())
	}
	return restCaps
}

func CapabilityIntermediaryListFromRest(caps []models.Capability) CapabilityIntermediaryList {
	intermediaries := CapabilityIntermediaryList{}
	for _, cap := range caps {
		intermediaries = append(intermediaries, CapabilityIntermediaryFromRest(cap))
	}
	return intermediaries
}
