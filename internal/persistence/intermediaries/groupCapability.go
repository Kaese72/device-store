package intermediaries

import (
	"encoding/json"
	"errors"

	"github.com/Kaese72/device-store/rest/models"
)

type GroupCapabilityIntermediary struct {
	GroupId string `db:"groupId"`
	Name    string `db:"name"`
}

func (c *GroupCapabilityIntermediary) ToRestModel() models.GroupCapability {
	return models.GroupCapability{
		Name: c.Name,
	}
}

func GroupCapabilityIntermediaryFromRest(cap models.GroupCapability) GroupCapabilityIntermediary {
	return GroupCapabilityIntermediary{
		Name: cap.Name,
	}
}

type GroupCapabilityIntermediaryList []GroupCapabilityIntermediary

func (caps *GroupCapabilityIntermediaryList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &caps)
}

func (caps GroupCapabilityIntermediaryList) ToRestModel() []models.GroupCapability {
	restCaps := []models.GroupCapability{}
	for _, cap := range caps {
		restCaps = append(restCaps, cap.ToRestModel())
	}
	return restCaps
}

func GroupCapabilityIntermediaryListFromRest(caps []models.GroupCapability) GroupCapabilityIntermediaryList {
	intermediaries := GroupCapabilityIntermediaryList{}
	for _, cap := range caps {
		intermediaries = append(intermediaries, GroupCapabilityIntermediaryFromRest(cap))
	}
	return intermediaries
}
