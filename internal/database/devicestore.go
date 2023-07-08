package database

import (
	"context"
	"errors"

	"github.com/Kaese72/device-store/internal/config"
	models "github.com/Kaese72/device-store/internal/models/intermediary"
	"github.com/Kaese72/device-store/internal/systemerrors"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
)

type DevicePersistenceDB interface {
	// Device Control
	FilterDevices(context.Context) ([]devicestoretemplates.Device, systemerrors.SystemError)
	GetDeviceByIdentifier(string, bool, context.Context) (devicestoretemplates.Device, systemerrors.SystemError)
	//// Attributes
	UpdateDeviceAttributes(devicestoretemplates.Device, bool, context.Context) (devicestoretemplates.Device, systemerrors.SystemError)
	//// Capabilities
	UpdateDeviceAttributesAndCapabilities(devicestoretemplates.Device, string, context.Context) (devicestoretemplates.Device, systemerrors.SystemError)
	GetCapability(string, string, context.Context) (models.CapabilityIntermediary, systemerrors.SystemError)
	//TriggerCapability(deviceId string, capName string, capArgs devicestoretemplates.CapabilityArgs) error
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	if conf.MongoDB.Validate() == nil {
		return NewMongoDBDevicePersistence(conf.MongoDB)

	} else {
		return nil, errors.New("no applicable database backend provided")
	}
}
