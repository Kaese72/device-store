package persistence

import (
	"context"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/internal/persistence/mariadb"
)

type DevicePersistenceDB interface {
	// Device Control
	GetDevices(context.Context) ([]intermediaries.DeviceIntermediary, error)
	//GetStoreDevice(int, bool, context.Context) (intermediaries.DeviceIntermediary, error)
	//// Attributes
	//GetDeviceAttributes(string, context.Context) ([]intermediaries.AttributeIntermediary, error)
	//// Capabilities
	PostDevice(context.Context, intermediaries.DeviceIntermediary) error
	GetCapabilityForActivation(ctx context.Context, deviceId int, capName string) (intermediaries.CapabilityIntermediaryActivation, error)
	//GetDeviceCapabilities(deviceId string, ctx context.Context) ([]intermediaries.CapabilityIntermediary, error)
	//// Groups
	//FilterGroups(context.Context) ([]devicestoretemplates.Group, error)
	//GetGroupByIdentifier(groupId string, expandCapabilities bool, ctx context.Context) (devicestoretemplates.Group, error)
	//GetGroupCapability(groupId string, capName string, ctx context.Context) (intermediaries.GroupCapabilityIntermediary, error)
	//UpdateGroup(group devicestoretemplates.Group, sourceBridge string, ctx context.Context) (devicestoretemplates.Group, error)
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	return mariadb.NewMariadbPersistence(conf)
}
