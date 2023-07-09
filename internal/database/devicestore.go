package database

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Kaese72/device-store/internal/config"
	models "github.com/Kaese72/device-store/internal/models/intermediary"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
)

type DevicePersistenceDB interface {
	// Device Control
	FilterDevices(context.Context) ([]devicestoretemplates.Device, error)
	GetDeviceByIdentifier(string, bool, context.Context) (devicestoretemplates.Device, error)
	//// Attributes
	UpdateDeviceAttributes(devicestoretemplates.Device, bool, context.Context) (devicestoretemplates.Device, error)
	//// Capabilities
	UpdateDeviceAttributesAndCapabilities(devicestoretemplates.Device, string, context.Context) (devicestoretemplates.Device, error)
	GetCapability(string, string, context.Context) (models.CapabilityIntermediary, error)
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	if conf.MongoDB.Validate() == nil {
		return NewMongoDBDevicePersistence(conf.MongoDB)

	} else {
		return nil, errors.New("no applicable database backend provided")
	}
}
