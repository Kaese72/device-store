package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoDevice struct {
	BridgeIdentifier      string `bson:"bridgeIdentifier"` // The identifier of this device from the bridge perspective
	BridgeKey             string `bson:"bridgeKey"`        // The "key" of the bridge owning this device
	DeviceStoreIdentifier string `bson:"_id,omitempty"`    // Should never be directly set. Generated on first insert and then never touched again
}

func (device MongoDevice) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": device,
	}
}

func (device MongoDevice) UniqueBridgeQuery() bson.D {
	return bson.D{primitive.E{Key: "bridgeIdentifier", Value: device.BridgeIdentifier}, primitive.E{Key: "bridgeKey", Value: device.BridgeKey}}
}

func UniqueDeviceStoreQuery(deviceStoreIdentifier string) bson.D {
	objID, _ := primitive.ObjectIDFromHex(deviceStoreIdentifier)
	return bson.D{primitive.E{Key: "_id", Value: objID}}
}
