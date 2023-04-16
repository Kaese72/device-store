package database

import (
	"errors"

	"github.com/Kaese72/device-store/internal/config"
	models "github.com/Kaese72/device-store/internal/models/intermediary"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
)

type DevicePersistenceDB interface {
	// Device Control
	FilterDevices() ([]devicestoretemplates.Device, error)
	GetDeviceByIdentifier(string, bool) (devicestoretemplates.Device, error)
	//// Attributes
	UpdateDeviceAttributes(devicestoretemplates.Device, bool) (devicestoretemplates.Device, error)
	//// Capabilities
	UpdateDeviceAttributesAndCapabilities(devicestoretemplates.Device, string) (devicestoretemplates.Device, error)
	GetCapability(deviceId string, capName string) (models.CapabilityIntermediary, error)
	//TriggerCapability(deviceId string, capName string, capArgs devicestoretemplates.CapabilityArgs) error
}

type NotFound error
type UnknownError error
type UserError error

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	if conf.MongoDB.Validate() == nil {
		return NewMongoDBDevicePersistence(conf.MongoDB)

	} else {
		return nil, errors.New("no applicable database backend provided")
	}
}
