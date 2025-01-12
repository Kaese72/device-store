package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type DeviceCapabilityIntermediary struct {
	// "db" is used when fetching directory from the database
	// "json" is used when fetching capabilities as part of a subquery on other models
	DeviceId string `db:"DeviceId" json:"DeviceId"`
	Name     string `db:"name" json:"name"`
}

func (c *DeviceCapabilityIntermediary) ToRestModel() models.DeviceCapability {
	return models.DeviceCapability{
		Name: c.Name,
	}
}

func DeviceCapabilityIntermediaryFromRest(cap models.DeviceCapability) DeviceCapabilityIntermediary {
	return DeviceCapabilityIntermediary{
		Name: cap.Name,
	}
}

type DeviceCapabilityIntermediaryList []DeviceCapabilityIntermediary

func (caps *DeviceCapabilityIntermediaryList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &caps)
}

func (caps DeviceCapabilityIntermediaryList) ToRestModel() []models.DeviceCapability {
	restCaps := []models.DeviceCapability{}
	for _, cap := range caps {
		restCaps = append(restCaps, cap.ToRestModel())
	}
	return restCaps
}

func DeviceCapabilityIntermediaryListFromRest(caps []models.DeviceCapability) DeviceCapabilityIntermediaryList {
	intermediaries := DeviceCapabilityIntermediaryList{}
	for _, cap := range caps {
		intermediaries = append(intermediaries, DeviceCapabilityIntermediaryFromRest(cap))
	}
	return intermediaries
}
