package database

import (
	"errors"

	"github.com/Kaese72/device-store/internal/config"
	models "github.com/Kaese72/device-store/internal/models/intermediary"
	"github.com/Kaese72/device-store/internal/systemerrors"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
)

type DevicePersistenceDB interface {
	// Device Control
	FilterDevices() ([]devicestoretemplates.Device, systemerrors.SystemError)
	GetDeviceByIdentifier(string, bool) (devicestoretemplates.Device, systemerrors.SystemError)
	//// Attributes
	UpdateDeviceAttributes(devicestoretemplates.Device, bool) (devicestoretemplates.Device, systemerrors.SystemError)
	//// Capabilities
	UpdateDeviceAttributesAndCapabilities(devicestoretemplates.Device, string) (devicestoretemplates.Device, systemerrors.SystemError)
	GetCapability(deviceId string, capName string) (models.CapabilityIntermediary, systemerrors.SystemError)
	//TriggerCapability(deviceId string, capName string, capArgs devicestoretemplates.CapabilityArgs) error
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	if conf.MongoDB.Validate() == nil {
		return NewMongoDBDevicePersistence(conf.MongoDB)

	} else {
		return nil, errors.New("no applicable database backend provided")
	}
}
