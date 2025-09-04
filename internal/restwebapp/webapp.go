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
	attendant   adapterattendant.Attendant
	events      *events.DeviceSubscriptions
}

func NewWebApp(persistence persistence.RestPersistenceDB, attendant adapterattendant.Attendant, events *events.DeviceSubscriptions) webApp {
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
	StoreDeviceIdentifier int                             `path:"storeDeviceIdentifier" doc:"the ID of the device to trigger capability for"`
	CapabilityID          string                          `path:"capabilityID" doc:"the capability to trigger"`
	Body                  restmodels.DeviceCapabilityArgs `body:""`
}) (*struct{}, error) {
	logging.Info(fmt.Sprintf("Triggering capability '%s' of device '%d'", input.CapabilityID, input.StoreDeviceIdentifier), ctx)
	capability, err := app.persistence.GetDeviceCapabilityForActivation(ctx, input.StoreDeviceIdentifier, input.CapabilityID)
	if err != nil {
		return nil, err
	}
	adapter, err := app.attendant.GetAdapter(string(capability.BridgeKey), ctx)
	if err != nil {
		return nil, err
	}
	sysErr := adapters.TriggerDeviceCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, input.Body)
	if sysErr != nil {
		return nil, sysErr
	}
	logging.Info("Capability seemingly successfully triggered", ctx)
	return &struct{}{}, nil
}

func (app webApp) TriggerGroupCapability(ctx context.Context, input *struct {
	StoreGroupIdentifier int                             `path:"storeGroupIdentifier" doc:"the ID of the group to trigger capability for"`
	CapabilityID         string                          `path:"capabilityID" doc:"the capability to trigger"`
	Body                 restmodels.DeviceCapabilityArgs `body:""`
}) (*struct{}, error) {
	logging.Info(fmt.Sprintf("Triggering capability '%s' of group '%d'", input.CapabilityID, input.StoreGroupIdentifier), ctx)
	capability, err := app.persistence.GetGroupCapabilityForActivation(ctx, input.StoreGroupIdentifier, input.CapabilityID)
	if err != nil {
		return nil, err
	}
	adapter, err := app.attendant.GetAdapter(string(capability.BridgeKey), ctx)
	if err != nil {
		return nil, err
	}
	sysErr := adapters.TriggerGroupCapability(ctx, adapter, capability.BridgeIdentifier, capability.Name, input.Body)
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
