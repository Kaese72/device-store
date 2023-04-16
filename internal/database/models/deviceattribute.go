package models

import (
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoDeviceAttribute struct {
	DeviceId       string         `bson:"deviceId"`
	AttributeName  string         `bson:"attributeName"`
	AttributeState AttributeState `bson:"attributeState"`
}

func (attribute MongoDeviceAttribute) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": attribute,
	}
}

func ExtractAttributeModelsFromAPIDeviceModel(device devicestoretemplates.Device) []MongoDeviceAttribute {
	attributes := []MongoDeviceAttribute{}
	for attributeKey, attribute := range device.Attributes {
		attributes = append(attributes, MongoDeviceAttribute{
			DeviceId:      device.Identifier,
			AttributeName: string(attributeKey),
			AttributeState: AttributeState{
				Boolean: attribute.Boolean,
				Numeric: attribute.Numeric,
				Text:    attribute.Text,
			},
		})
	}
	return attributes
}
