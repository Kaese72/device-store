package models

import (
	"time"

	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoGroupCapability struct {
	GroupId        string    `bson:"groupId"`
	CapabilityName string    `bson:"capabilityName"`
	GroupBridgeKey string    `bson:"groupBridgeKey"`
	LastSeen       time.Time `bson:"lastSeen"`
}

func (capability MongoGroupCapability) ConvertToAPICapability() devicestoretemplates.Capability {
	return devicestoretemplates.Capability{
		LastSeen: capability.LastSeen,
	}
}

func (capability MongoGroupCapability) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": map[string]string{
			"groupId":        capability.GroupId,
			"capabilityName": capability.CapabilityName,
			"groupBridgeKey": string(capability.GroupBridgeKey),
		},
		"$currentDate": bson.M{
			"lastSeen": bson.M{"$type": "timestamp"},
		},
	}
}

func ExtractGroupCapabilityFromAPI(group devicestoretemplates.Group, bridgeKey string) []MongoGroupCapability {
	capabilities := []MongoGroupCapability{}
	for capabilityKey := range group.Capabilities {
		capabilities = append(capabilities, MongoGroupCapability{
			GroupId:        group.Identifier,
			CapabilityName: string(capabilityKey),
			GroupBridgeKey: bridgeKey,
		})
	}
	return capabilities
}
