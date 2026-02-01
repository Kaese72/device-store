package ingestwebapp

import (
	"context"

	"github.com/Kaese72/device-store/eventmodels"
	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/internal/persistence"
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
	Body ingestmodels.IngestDevice `body:""`
}) (*struct{}, error) {
	device := input.Body
	device.AdapterId = ctx.Value(adapterIDContextKey{}).(int)
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
	Body ingestmodels.IngestGroup `body:""`
}) (*struct{}, error) {
	group := input.Body
	group.AdapterId = ctx.Value(adapterIDContextKey{}).(int)
	err := app.persistence.PostGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}
