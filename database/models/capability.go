package models

import (
	"time"

	"github.com/Kaese72/sdup-lib/devicestoretemplates"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoDeviceCapability struct {
	DeviceId            string                         `bson:"deviceId"`
	CapabilityName      string                         `bson:"capabilityName"`
	CapabilityBridgeKey devicestoretemplates.BridgeKey `bson:"capabilityBridgeKey"`
	CapabilityBridgeURI string                         `bson:"bridgeURI"`
	LastSeen            time.Time                      `bson:"lastSeen"`
}

func (capability MongoDeviceCapability) ConvertToAPICapability() devicestoretemplates.Capability {
	return devicestoretemplates.Capability{
		LastSeen: capability.LastSeen,
	}
}

func (capability MongoDeviceCapability) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": map[string]string{
			"deviceId":            capability.DeviceId,
			"capabilityName":      capability.CapabilityName,
			"capabilityBridgeKey": string(capability.CapabilityBridgeKey),
			"bridgeURI":           capability.CapabilityBridgeURI,
		},
		"$currentDate": bson.M{
			"lastSeen": bson.M{"$type": "timestamp"},
		},
	}
}

func ExtractCapabilityModelsFromAPIDeviceModel(device devicestoretemplates.Device, bridge Bridge) []MongoDeviceCapability {
	capabilities := []MongoDeviceCapability{}
	for capabilityKey := range device.Capabilities {
		capabilities = append(capabilities, MongoDeviceCapability{
			DeviceId:            device.Identifier,
			CapabilityName:      string(capabilityKey),
			CapabilityBridgeKey: devicestoretemplates.BridgeKey(bridge.Identifier),
			CapabilityBridgeURI: bridge.URI,
		})
	}
	return capabilities
}

func ReduceToMostRelevantCapabilities(capabilities []MongoDeviceCapability) []MongoDeviceCapability {
	// Assumes all passed in capabilities belong to the same device
	capMap := map[string]MongoDeviceCapability{}
	for _, capability := range capabilities {
		if capability.LastSeen.After(capMap[capability.CapabilityName].LastSeen) {
			capMap[capability.CapabilityName] = capability
		}
	}
	ret := []MongoDeviceCapability{}
	for _, capability := range capMap {
		ret = append(ret, capability)
	}
	return ret
}
