package models

import (
	"github.com/pkg/errors"

	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
)

func ConvertAttributesToAPIDevice(singleDeviceAttributes []MongoDeviceAttribute) (devicestoretemplates.Device, error) {
	// This assumes all attributes and capabilities belong to the same device
	if len(singleDeviceAttributes) == 0 {
		return devicestoretemplates.Device{}, errors.New("singleDeviceAttributes is empty")
	}
	deviceId := singleDeviceAttributes[0].DeviceId
	device := devicestoretemplates.Device{
		Identifier: deviceId,
		Attributes: map[devicestoretemplates.AttributeKey]devicestoretemplates.AttributeState{},
	}
	for _, attribute := range singleDeviceAttributes {
		if attribute.DeviceId != deviceId {
			return device, errors.New("deviceIds do not match")
		}
		device.Attributes[devicestoretemplates.AttributeKey(attribute.AttributeName)] = attribute.AttributeState.ConvertToAPIAttributeState()
	}
	return device, nil
}

func CreateAPIDevicesFromAttributes(mutipleDeviceAttributes []MongoDeviceAttribute) (map[string]devicestoretemplates.Device, error) {
	perDeviceAttributes := map[string][]MongoDeviceAttribute{}
	var err error
	for _, deviceAttribute := range mutipleDeviceAttributes {
		perDeviceAttributes[deviceAttribute.DeviceId] = append(perDeviceAttributes[deviceAttribute.DeviceId], deviceAttribute)
	}
	apiDevices := map[string]devicestoretemplates.Device{}
	for deviceId, attributes := range perDeviceAttributes {
		apiDevices[deviceId], err = ConvertAttributesToAPIDevice(attributes)
		if err != nil {
			return apiDevices, err
		}
	}
	return apiDevices, nil
}
