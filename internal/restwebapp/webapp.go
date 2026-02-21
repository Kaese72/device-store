package restwebapp

import (
	"context"
	"fmt"

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

// GetDevices returns all devices in the database
func (app webApp) GetDevices(ctx context.Context, input *struct {
	Filters string `query:"filters" doc:"a string JSON array of objects containting key, op, and value for filtering"`
}) (*struct {
	Body []restmodels.Device
}, error) {
	filters, err := restmodels.ParseQueryIntoFilters(input.Filters)
	if err != nil {
		return nil, err
	}
	restDevices, err := app.persistence.GetDevices(ctx, filters)
	return &struct {
		Body []restmodels.Device
	}{
		Body: restDevices,
	}, err
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
	restDevices, err := app.persistence.GetDevices(ctx, filter)
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
}) (*struct {
	Body []restmodels.AttributeAudit
}, error) {
	filters, err := restmodels.ParseQueryIntoFilters(input.Filters)
	if err != nil {
		return nil, err
	}
	restAudits, err := app.persistence.GetAttributeAudits(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &struct{ Body []restmodels.AttributeAudit }{Body: restAudits}, nil
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
	sysErr := adapters.TriggerDeviceCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, capArg)
	if sysErr != nil {
		return nil, sysErr
	}
	logging.Info("Capability seemingly successfully triggered", ctx)
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
	sysErr := adapters.TriggerGroupCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, capArgs)
	if sysErr != nil {
		return nil, sysErr
	}
	logging.Info("Capability seemingly successfully triggered", ctx)
	return &struct{}{}, nil
}

func (app webApp) GetGroups(ctx context.Context, input *struct {
	Filters string `query:"filters" doc:"a string JSON array of objects containing key, op, and value for filtering"`
}) (*struct {
	Body []restmodels.Group
}, error) {
	filters, err := restmodels.ParseQueryIntoFilters(input.Filters)
	if err != nil {
		return nil, err
	}
	restGroups, err := app.persistence.GetGroups(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &struct{ Body []restmodels.Group }{Body: restGroups}, nil
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
	restGroups, err := app.persistence.GetGroups(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(restGroups) == 0 {
		return nil, huma.Error404NotFound("group not found")
	}
	return &struct{ Body restmodels.Group }{Body: restGroups[0]}, nil
}
