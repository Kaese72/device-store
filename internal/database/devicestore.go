package database

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Kaese72/device-store/internal/config"
	intermediary "github.com/Kaese72/device-store/internal/models/intermediary"
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
	GetCapability(deviceId string, capName string, ctx context.Context) (intermediary.CapabilityIntermediary, error)
	//// Groups
	FilterGroups(context.Context) ([]devicestoretemplates.Group, error)
	GetGroupByIdentifier(groupId string, expandCapabilities bool, ctx context.Context) (devicestoretemplates.Group, error)
	GetGroupCapability(groupId string, capName string, ctx context.Context) (intermediary.GroupCapabilityIntermediary, error)
	UpdateGroup(group devicestoretemplates.Group, sourceBridge string, ctx context.Context) (devicestoretemplates.Group, error)
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	if conf.MongoDB.Validate() == nil {
		return NewMongoDBDevicePersistence(conf.MongoDB)

	} else {
		return nil, errors.New("no applicable database backend provided")
	}
}
