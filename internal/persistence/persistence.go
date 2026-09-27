package persistence

import (
	"context"

	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/restmodels"
	"github.com/Kaese72/huemie-lib/query"
)

type RestPersistenceDB interface {
	// Device Control
	// GetDevices returns the page of devices matching filters and ordered by
	// sorts (or the implementation's default order if sorts is empty), along
	// with the total number of devices matching filters (ignoring pagination).
	GetDevices(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination query.Pagination) ([]restmodels.Device, int, error)
	DeleteDevice(ctx context.Context, storeIdentifier int) error
	// UpdateDeviceName sets the human-readable name of a device and returns the updated device.
	UpdateDeviceName(ctx context.Context, storeIdentifier int, name string) (restmodels.Device, error)
	GetDeviceCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.DeviceCapabilityIntermediaryActivation, error)
	// GetAttributeStatistics returns, for each known attribute name (optionally
	// restricted to names), how many devices store that attribute's value
	// under each of the three possible types.
	GetAttributeStatistics(ctx context.Context, names []string) ([]restmodels.AttributeStatistic, error)
	// Audits
	GetAttributeAudits(ctx context.Context, filters []query.Filter, pagination query.Pagination) ([]restmodels.AttributeAudit, int, error)
	WriteCapabilityTriggerAudit(ctx context.Context, deviceId int, capabilityName string, success bool, errorMessage *string, arguments string) error
	GetCapabilityTriggerAudits(ctx context.Context, deviceId int, pagination query.Pagination) ([]restmodels.CapabilityTriggerAudit, int, error)
	//// Groups
	// GetGroups returns the page of groups matching filters and ordered by
	// sorts (or the implementation's default order if sorts is empty), along
	// with the total number of groups matching filters (ignoring pagination).
	GetGroups(ctx context.Context, filters []query.Filter, sorts []query.Sort, pagination query.Pagination) ([]restmodels.Group, int, error)
	// SearchGroupsByName performs a fuzzy search for groups by name, returning
	// up to limit results ordered from best to worst match.
	SearchGroupsByName(ctx context.Context, query string, limit int) ([]restmodels.GroupSearchResult, error)
	DeleteGroup(ctx context.Context, storeIdentifier int) error
	GetGroupCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.GroupCapabilityIntermediaryActivation, error)
	WriteGroupCapabilityTriggerAudit(ctx context.Context, groupId int, capabilityName string, success bool, errorMessage *string, arguments string) error
	GetGroupCapabilityTriggerAudits(ctx context.Context, groupId int, pagination query.Pagination) ([]restmodels.GroupCapabilityTriggerAudit, int, error)
}

type IngestPersistenceDB interface {
	// Device Control
	// PostDevice updates a device and returns the stuff that has been changed
	PostDevice(context.Context, ingestmodels.IngestDevice) (int, []ingestmodels.IngestAttribute, error)
	//// Groups
	PostGroup(context.Context, ingestmodels.IngestGroup) error
}
