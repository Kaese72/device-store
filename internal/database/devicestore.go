package database

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/models/intermediaries"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
)

type DevicePersistenceDB interface {
	// Device Control
	GetDevices(context.Context) ([]intermediaries.DeviceIntermediary, error)
	GetStoreDevice(string, bool, context.Context) (intermediaries.DeviceIntermediary, error)
	//// Attributes
	GetDeviceAttributes(string, context.Context) ([]intermediaries.AttributeIntermediary, error)
	//// Capabilities
	UpdateDevice(devicestoretemplates.Device, string, context.Context) error
	GetCapability(deviceId string, capName string, ctx context.Context) (intermediaries.CapabilityIntermediary, error)
	GetDeviceCapabilities(deviceId string, ctx context.Context) ([]intermediaries.CapabilityIntermediary, error)
	//// Groups
	FilterGroups(context.Context) ([]devicestoretemplates.Group, error)
	GetGroupByIdentifier(groupId string, expandCapabilities bool, ctx context.Context) (devicestoretemplates.Group, error)
	GetGroupCapability(groupId string, capName string, ctx context.Context) (intermediaries.GroupCapabilityIntermediary, error)
	UpdateGroup(group devicestoretemplates.Group, sourceBridge string, ctx context.Context) (devicestoretemplates.Group, error)
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	if conf.MongoDB.Validate() == nil {
		db, err := NewMongoDBDevicePersistence(conf.MongoDB)
		if err != nil {
			return db, err
		}
		if conf.Purge {
			err := db.purge()
			if err != nil {
				return db, err
			}
		}
		return db, err

	} else {
		return nil, errors.New("no applicable database backend provided")
	}
}
