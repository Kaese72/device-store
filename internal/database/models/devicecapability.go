package models

import (
	"time"

	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoDeviceCapability struct {
	StoreDeviceIdentifier  string    `bson:"storeDeviceIdentifier,omitempty"` // The device this attribute belongs to in the device store
	BridgeDeviceIdentifier string    `bson:"bridgeDeviceIdentifier"`          // The device identifier in the bridge
	BridgeKey              string    `bson:"bridgeKey"`                       // The "key" of the bridge
	Name                   string    `bson:"name"`
	LastSeen               time.Time `bson:"lastSeen"`
}

func (capability MongoDeviceCapability) ConvertToAPICapability() devicestoretemplates.Capability {
	return devicestoretemplates.Capability{
		LastSeen: capability.LastSeen,
	}
}

func (capability MongoDeviceCapability) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": map[string]string{
			"storeDeviceIdentifier":  capability.StoreDeviceIdentifier,
			"bridgeKey":              capability.BridgeKey,
			"bridgeDeviceIdentifier": capability.BridgeDeviceIdentifier,
			"name":                   capability.Name,
		},
		"$currentDate": bson.M{
			"lastSeen": bson.M{"$type": "timestamp"},
		},
	}
}

func (capability MongoDeviceCapability) UniqueQuery() bson.D {
	return MongoDeviceCapabilityUniqueQuery(capability.StoreDeviceIdentifier, capability.Name)
}

func MongoDeviceCapabilityUniqueQuery(identifier string, name string) bson.D {
	x := bson.D{primitive.E{Key: "storeDeviceIdentifier", Value: identifier}}
	if name != "" {
		x = append(x, primitive.E{Key: "name", Value: name})
	}
	return x
}

func ExtractCapabilityModelsFromAPIDeviceModel(device devicestoretemplates.Device, deviceStoreIdentifier string, bridgeKey string) []MongoDeviceCapability {
	capabilities := []MongoDeviceCapability{}
	for capabilityKey := range device.Capabilities {
		capabilities = append(capabilities, MongoDeviceCapability{
			StoreDeviceIdentifier:  deviceStoreIdentifier,
			BridgeDeviceIdentifier: device.Identifier,
			BridgeKey:              bridgeKey,
			Name:                   string(capabilityKey),
		})
	}
	return capabilities
}
