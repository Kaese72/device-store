package persistence

import (
	"context"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/internal/persistence/mariadb"
)

type DevicePersistenceDB interface {
	// Device Control
	GetDevices(context.Context, []intermediaries.Filter) ([]intermediaries.DeviceIntermediary, error)
	PostDevice(context.Context, intermediaries.DeviceIntermediary) error
	GetDeviceCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.DeviceCapabilityIntermediaryActivation, error)
	//// Groups
	GetGroups(context.Context, []intermediaries.Filter) ([]intermediaries.GroupIntermediary, error)
	PostGroup(ctx context.Context, group intermediaries.GroupIntermediary) error
	GetGroupCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.GroupCapabilityIntermediaryActivation, error)
}

func NewDevicePersistenceDB(conf config.DatabaseConfig) (DevicePersistenceDB, error) {
	return mariadb.NewMariadbPersistence(conf)
}
