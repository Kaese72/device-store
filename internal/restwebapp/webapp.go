package restwebapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/adapters"
	"github.com/Kaese72/device-store/internal/events"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence"
	"github.com/Kaese72/device-store/restmodels"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

type webApp struct {
	persistence persistence.RestPersistenceDB
	attendant   adapterattendant.AdapterTriggerClient
	events      *events.DeviceSubscriptions
}

func NewWebApp(persistence persistence.RestPersistenceDB, attendant adapterattendant.AdapterTriggerClient, events *events.DeviceSubscriptions) webApp {
	return webApp{
		persistence: persistence,
		attendant:   attendant,
		events:      events,
	}
}

// GetDevices returns a page of devices in the database
func (app webApp) GetDevices(ctx context.Context, input *struct {
	Filters string `query:"filters" doc:"a string JSON array of objects containting key, op, and value for filtering"`
	Offset  int    `query:"offset" default:"0" minimum:"0" doc:"number of matching devices to skip"`
	Limit   int    `query:"limit" default:"50" minimum:"1" maximum:"200" doc:"maximum number of devices to return"`
}) (*struct {
	TotalCount int `header:"X-Total-Count" doc:"total number of devices matching the filters, ignoring pagination"`
	Body       []restmodels.Device
}, error) {
	filters, err := restmodels.ParseQueryIntoFilters(input.Filters)
	if err != nil {
		return nil, err
	}
	restDevices, total, err := app.persistence.GetDevices(ctx, filters, restmodels.Pagination{Offset: input.Offset, Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	return &struct {
		TotalCount int `header:"X-Total-Count" doc:"total number of devices matching the filters, ignoring pagination"`
		Body       []restmodels.Device
	}{
		TotalCount: total,
		Body:       restDevices,
	}, nil
}

func (app webApp) GetDevice(ctx context.Context, input *struct {
	StoreDeviceIdentifier string `path:"storeDeviceIdentifier" doc:"the ID of the device to retrieve"`
}) (*struct {
	Body restmodels.Device
}, error) {
	// Create a filter for the deviceId and use the GetDevices method
	filter := []restmodels.Filter{
		{
			Key:      "id",
			Value:    input.StoreDeviceIdentifier,
			Operator: "eq",
		},
	}
	restDevices, _, err := app.persistence.GetDevices(ctx, filter, restmodels.Pagination{})
	if err != nil {
		return nil, err
	}
	if len(restDevices) == 0 {
		return nil, huma.Error404NotFound("device not found")
	}
	return &struct {
		Body restmodels.Device
	}{
		Body: restDevices[0],
	}, err
}

func (app webApp) DeleteDevice(ctx context.Context, input *struct {
	StoreDeviceIdentifier int `path:"storeDeviceIdentifier" doc:"the ID of the device to forget"`
}) (*struct{}, error) {
	err := app.persistence.DeleteDevice(ctx, input.StoreDeviceIdentifier)
	if err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

func (app webApp) GetAttributeAudits(ctx context.Context, input *struct {
	Filters string `query:"filters" doc:"a string JSON array of objects containing key, op, and value for filtering"`
	Offset  int    `query:"offset" default:"0" minimum:"0" doc:"number of matching audits to skip"`
	Limit   int    `query:"limit" default:"50" minimum:"1" maximum:"200" doc:"maximum number of audits to return"`
}) (*struct {
	TotalCount int `header:"X-Total-Count" doc:"total number of audits matching the filters, ignoring pagination"`
	Body       []restmodels.AttributeAudit
}, error) {
	filters, err := restmodels.ParseQueryIntoFilters(input.Filters)
	if err != nil {
		return nil, err
	}
	restAudits, total, err := app.persistence.GetAttributeAudits(ctx, filters, restmodels.Pagination{Offset: input.Offset, Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	return &struct {
		TotalCount int `header:"X-Total-Count" doc:"total number of audits matching the filters, ignoring pagination"`
		Body       []restmodels.AttributeAudit
	}{TotalCount: total, Body: restAudits}, nil
}

// GetAttributeStatistics returns, per attribute name, how many devices store
// that attribute's value as each of the three possible types.
func (app webApp) GetAttributeStatistics(ctx context.Context, input *struct {
	Names string `query:"names" doc:"a comma separated list of attribute names to restrict the results to; if omitted, all attribute names are returned"`
}) (*struct {
	Body []restmodels.AttributeStatistic
}, error) {
	var names []string
	for _, name := range strings.Split(input.Names, ",") {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		names = append(names, trimmed)
	}
	stats, err := app.persistence.GetAttributeStatistics(ctx, names)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body []restmodels.AttributeStatistic
	}{Body: stats}, nil
}

// StreamDeviceUpdates is a SSE endpoint that sends updates from
func (app webApp) StreamDeviceUpdates(ctx context.Context, input *struct{}, send sse.Sender) {
	// writer.Header().Set("Access-Control-Allow-Origin", "*")
	deviceUpdates := app.events.Subscribe(ctx)
	for update := range deviceUpdates {
		if err := send.Data(update); err != nil {
			// Even though this is a critical error, we continue
			logging.ErrorErr(err, ctx)
			continue
		}
	}
}

func (app webApp) TriggerDeviceCapability(ctx context.Context, input *struct {
	StoreDeviceIdentifier int                              `path:"storeDeviceIdentifier" doc:"the ID of the device to trigger capability for"`
	CapabilityID          string                           `path:"capabilityID" doc:"the capability to trigger"`
	Body                  *restmodels.DeviceCapabilityArgs `body:""`
}) (*struct{}, error) {
	logging.Info(fmt.Sprintf("Triggering capability '%s' of device '%d'", input.CapabilityID, input.StoreDeviceIdentifier), ctx)
	capability, err := app.persistence.GetDeviceCapabilityForActivation(ctx, input.StoreDeviceIdentifier, input.CapabilityID)
	if err != nil {
		return nil, err
	}
	adapter, err := app.attendant.GetAdapterAddress(ctx, capability.AdapterId)
	if err != nil {
		return nil, err
	}
	capArg := restmodels.DeviceCapabilityArgs{}
	if input.Body != nil {
		capArg = *input.Body
	}
	argsJSON, _ := json.Marshal(capArg)
	sysErr := adapters.TriggerDeviceCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, capArg)
	if sysErr != nil {
		errMsg := sysErr.Error()
		_ = app.persistence.WriteCapabilityTriggerAudit(ctx, input.StoreDeviceIdentifier, input.CapabilityID, false, &errMsg, string(argsJSON))
		return nil, sysErr
	}
	_ = app.persistence.WriteCapabilityTriggerAudit(ctx, input.StoreDeviceIdentifier, input.CapabilityID, true, nil, string(argsJSON))
	logging.Info("Capability seemingly successfully triggered", ctx)
	return &struct{}{}, nil
}

func (app webApp) GetDeviceCapabilityTriggerAudits(ctx context.Context, input *struct {
	StoreDeviceIdentifier int `path:"storeDeviceIdentifier" doc:"the ID of the device"`
	Offset                int `query:"offset" default:"0" minimum:"0" doc:"number of matching audits to skip"`
	Limit                 int `query:"limit" default:"50" minimum:"1" maximum:"200" doc:"maximum number of audits to return"`
}) (*struct {
	TotalCount int `header:"X-Total-Count" doc:"total number of audits for the device, ignoring pagination"`
	Body       []restmodels.CapabilityTriggerAudit
}, error) {
	audits, total, err := app.persistence.GetCapabilityTriggerAudits(ctx, input.StoreDeviceIdentifier, restmodels.Pagination{Offset: input.Offset, Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	if audits == nil {
		audits = []restmodels.CapabilityTriggerAudit{}
	}
	return &struct {
		TotalCount int `header:"X-Total-Count" doc:"total number of audits for the device, ignoring pagination"`
		Body       []restmodels.CapabilityTriggerAudit
	}{TotalCount: total, Body: audits}, nil
}

func (app webApp) DeleteGroup(ctx context.Context, input *struct {
	StoreGroupIdentifier int `path:"storeGroupIdentifier" doc:"the ID of the group to forget"`
}) (*struct{}, error) {
	err := app.persistence.DeleteGroup(ctx, input.StoreGroupIdentifier)
	if err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

func (app webApp) TriggerGroupCapability(ctx context.Context, input *struct {
	StoreGroupIdentifier int                              `path:"storeGroupIdentifier" doc:"the ID of the group to trigger capability for"`
	CapabilityID         string                           `path:"capabilityID" doc:"the capability to trigger"`
	Body                 *restmodels.DeviceCapabilityArgs `body:""`
}) (*struct{}, error) {
	logging.Info(fmt.Sprintf("Triggering capability '%s' of group '%d'", input.CapabilityID, input.StoreGroupIdentifier), ctx)
	capability, err := app.persistence.GetGroupCapabilityForActivation(ctx, input.StoreGroupIdentifier, input.CapabilityID)
	if err != nil {
		return nil, err
	}
	adapter, err := app.attendant.GetAdapterAddress(ctx, capability.AdapterId)
	if err != nil {
		return nil, err
	}
	capArgs := restmodels.DeviceCapabilityArgs{}
	if input.Body != nil {
		capArgs = *input.Body
	}
	argsJSON, _ := json.Marshal(capArgs)
	sysErr := adapters.TriggerGroupCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, capArgs)
	if sysErr != nil {
		errMsg := sysErr.Error()
		_ = app.persistence.WriteGroupCapabilityTriggerAudit(ctx, input.StoreGroupIdentifier, input.CapabilityID, false, &errMsg, string(argsJSON))
		return nil, sysErr
	}
	_ = app.persistence.WriteGroupCapabilityTriggerAudit(ctx, input.StoreGroupIdentifier, input.CapabilityID, true, nil, string(argsJSON))
	logging.Info("Capability seemingly successfully triggered", ctx)
	return &struct{}{}, nil
}

func (app webApp) GetGroupCapabilityTriggerAudits(ctx context.Context, input *struct {
	StoreGroupIdentifier int `path:"storeGroupIdentifier" doc:"the ID of the group"`
	Offset               int `query:"offset" default:"0" minimum:"0" doc:"number of matching audits to skip"`
	Limit                int `query:"limit" default:"50" minimum:"1" maximum:"200" doc:"maximum number of audits to return"`
}) (*struct {
	TotalCount int `header:"X-Total-Count" doc:"total number of audits for the group, ignoring pagination"`
	Body       []restmodels.GroupCapabilityTriggerAudit
}, error) {
	audits, total, err := app.persistence.GetGroupCapabilityTriggerAudits(ctx, input.StoreGroupIdentifier, restmodels.Pagination{Offset: input.Offset, Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	if audits == nil {
		audits = []restmodels.GroupCapabilityTriggerAudit{}
	}
	return &struct {
		TotalCount int `header:"X-Total-Count" doc:"total number of audits for the group, ignoring pagination"`
		Body       []restmodels.GroupCapabilityTriggerAudit
	}{TotalCount: total, Body: audits}, nil
}

func (app webApp) GetGroups(ctx context.Context, input *struct {
	Filters string `query:"filters" doc:"a string JSON array of objects containing key, op, and value for filtering"`
	Offset  int    `query:"offset" default:"0" minimum:"0" doc:"number of matching groups to skip"`
	Limit   int    `query:"limit" default:"50" minimum:"1" maximum:"200" doc:"maximum number of groups to return"`
}) (*struct {
	TotalCount int `header:"X-Total-Count" doc:"total number of groups matching the filters, ignoring pagination"`
	Body       []restmodels.Group
}, error) {
	filters, err := restmodels.ParseQueryIntoFilters(input.Filters)
	if err != nil {
		return nil, err
	}
	restGroups, total, err := app.persistence.GetGroups(ctx, filters, restmodels.Pagination{Offset: input.Offset, Limit: input.Limit})
	if err != nil {
		return nil, err
	}
	return &struct {
		TotalCount int `header:"X-Total-Count" doc:"total number of groups matching the filters, ignoring pagination"`
		Body       []restmodels.Group
	}{TotalCount: total, Body: restGroups}, nil
}

func (app webApp) GetGroup(ctx context.Context, input *struct {
	StoreGroupIdentifier string `path:"storeGroupIdentifier" doc:"the ID of the group to retrieve"`
}) (*struct {
	Body restmodels.Group
}, error) {
	// Create a filter for the groupId and use the GetGroups method
	filter := []restmodels.Filter{
		{
			Key:      "id",
			Operator: "eq",
			Value:    input.StoreGroupIdentifier,
		},
	}
	restGroups, _, err := app.persistence.GetGroups(ctx, filter, restmodels.Pagination{})
	if err != nil {
		return nil, err
	}
	if len(restGroups) == 0 {
		return nil, huma.Error404NotFound("group not found")
	}
	return &struct{ Body restmodels.Group }{Body: restGroups[0]}, nil
}
