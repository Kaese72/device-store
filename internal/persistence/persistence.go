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
	DeleteDevice(ctx context.Context, storeIdentifier int) error
	GetDeviceCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.DeviceCapabilityIntermediaryActivation, error)
	// Audits
	GetAttributeAudits(context.Context, []restmodels.Filter) ([]restmodels.AttributeAudit, error)
	//// Groups
	GetGroups(context.Context, []restmodels.Filter) ([]restmodels.Group, error)
	GetGroupCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.GroupCapabilityIntermediaryActivation, error)
}

type IngestPersistenceDB interface {
	// Device Control
	// PostDevice updates a device and returns the stuff that has been changed
	PostDevice(context.Context, ingestmodels.IngestDevice) (int, []ingestmodels.IngestAttribute, error)
	//// Groups
	PostGroup(context.Context, ingestmodels.IngestGroup) error
}
