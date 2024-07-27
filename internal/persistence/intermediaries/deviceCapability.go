package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type DeviceCapabilityIntermediary struct {
	DeviceId string `json:"DeviceId" db:"DeviceId"`
	Name     string `json:"name" db:"name"`
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
