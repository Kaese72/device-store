package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MongoDeviceAttribute struct {
	DeviceStoreIdentifier string         `bson:"storeDeviceIdentifier,omitempty"` // The device this attribute belongs to
	Name                  string         `bson:"name,omitempty"`
	State                 AttributeState `bson:"state"`
}

func (attribute MongoDeviceAttribute) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": attribute,
	}
}

func (attribute MongoDeviceAttribute) UniqueQuery() bson.D {
	return bson.D{bson.E{Key: "storeDeviceIdentifier", Value: attribute.DeviceStoreIdentifier}, primitive.E{Key: "name", Value: attribute.Name}}

}

// func ExtractAttributeModelsFromAPIDeviceModel(device devicestoretemplates.Device) []MongoDeviceAttribute {
// 	attributes := []MongoDeviceAttribute{}
// 	for attributeKey, attribute := range device.Attributes {
// 		attributes = append(attributes, MongoDeviceAttribute{
// 			DeviceId:  device.Identifier,
// 			BridgeKey: device.BridgeKey,
// 			// StoreIdentifier: XXX, Intentionally left out since read-only
// 			AttributeName: string(attributeKey),
// 			AttributeState: AttributeState{
// 				Boolean: attribute.Boolean,
// 				Numeric: attribute.Numeric,
// 				Text:    attribute.Text,
// 			},
// 		})
// 	}
// 	return attributes
// }
