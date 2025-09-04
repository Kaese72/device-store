package ingestwebapp

import (
	"context"

	"github.com/Kaese72/device-store/eventmodels"
	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/internal/persistence"
	"github.com/danielgtaylor/huma/v2"
)

type webApp struct {
	persistence       persistence.IngestPersistenceDB
	deviceUpdatesChan chan eventmodels.DeviceAttributeUpdate
}

func NewWebApp(persistence persistence.IngestPersistenceDB, deviceUpdatesChan chan eventmodels.DeviceAttributeUpdate) webApp {
	return webApp{
		persistence:       persistence,
		deviceUpdatesChan: deviceUpdatesChan,
	}
}

func (app webApp) PostDevice(ctx context.Context, input *struct {
	BridgeKey string                    `header:"Bridge-Key" doc:"Bridge key for authentication"`
	Body      ingestmodels.IngestDevice `body:""`
}) (*struct{}, error) {
	if input.BridgeKey == "" {
		return nil, huma.Error400BadRequest("May not update devices when not identifying as a bridge")
	}
	device := input.Body
	device.BridgeKey = input.BridgeKey
	deviceId, updates, err := app.persistence.PostDevice(ctx, device)
	if err != nil {
		return nil, err
	}
	if len(updates) != 0 {
		deviceUpdateEvent := eventmodels.DeviceAttributeUpdate{
			DeviceID:   deviceId,
			Attributes: []eventmodels.UpdatedAttribute{},
		}
		for _, update := range updates {
			deviceUpdateEvent.Attributes = append(deviceUpdateEvent.Attributes, eventmodels.UpdatedAttribute{
				Name:    update.Name,
				Boolean: update.Boolean,
				Text:    update.Text,
				Numeric: update.Numeric,
			})
		}
		app.deviceUpdatesChan <- deviceUpdateEvent
	}
	return &struct{}{}, nil
}

func (app webApp) PostGroup(ctx context.Context, input *struct {
	BridgeKey string                   `header:"Bridge-Key" doc:"Bridge key for authentication"`
	Body      ingestmodels.IngestGroup `body:""`
}) (*struct{}, error) {
	if input.BridgeKey == "" {
		return nil, huma.Error400BadRequest("May not update groups when not identifying as a bridge")
	}
	group := input.Body
	group.BridgeKey = input.BridgeKey
	err := app.persistence.PostGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}
