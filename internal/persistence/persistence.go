package persistence

import (
	"context"

	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/restmodels"
)

type RestPersistenceDB interface {
	// Device Control
	GetDevices(context.Context, []restmodels.Filter) ([]restmodels.Device, error)
	GetDeviceCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.DeviceCapabilityIntermediaryActivation, error)
	//// Groups
	GetGroups(context.Context, []restmodels.Filter) ([]restmodels.Group, error)
	GetGroupCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.GroupCapabilityIntermediaryActivation, error)
}

type IngestPersistenceDB interface {
	// Device Control
	// PostDevice updates a device and returns the stuff that has been changed
	PostDevice(context.Context, ingestmodels.Device) ([]ingestmodels.Attribute, error)
	//// Groups
	PostGroup(context.Context, ingestmodels.Group) error
}
