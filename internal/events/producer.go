package events

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/Kaese72/device-store/eventmodels"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventsProducer struct {
	connection         *amqp.Connection
	deviceUpdatesTopic string
}

func NewEventsProducer(conf config.EventConfig) (*EventsProducer, error) {
	conn, err := amqp.Dial(conf.ConnectionString)
	if err != nil {
		return nil, err
	}
	return &EventsProducer{connection: conn, deviceUpdatesTopic: conf.DeviceUpdatesTopic}, nil
}

func (h *EventsProducer) Close() error {
	if h.connection != nil {
		return h.connection.Close()
	}
	return nil
}

// DeviceUpdates returns a channel on which we get device updates on
// from the message queue. The channel is closed when the connection is closed.
func (h *EventsProducer) ProduceDeviceUpdates() (chan eventmodels.DeviceAttributeUpdate, error) {
	ch, err := h.connection.Channel()
	if err != nil {
		return nil, err
	}
	// FIXME this piece of code is in two places... How to avoid duplication?
	err = ch.ExchangeDeclare(
		"deviceAttributeUpdates", // name
		"fanout",                 // Send to all attached queues
		true,                     // durable
		false,                    // auto-deleted
		false,                    // internal
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return nil, err
	}
	retChan := make(chan eventmodels.DeviceAttributeUpdate, 10)
	go func() {
		for update := range retChan {
			body, err := json.Marshal(update)
			if err != nil {
				logging.ErrorErr(err, context.Background())
				continue // We do not want to stop even if something goes wrong
			}
			err = ch.PublishWithContext(context.Background(),
				"deviceAttributeUpdates",      // default exchange
				strconv.Itoa(update.DeviceID), // routing key
				false,                         // mandatory
				false,                         // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        body,
				},
			)
			if err != nil {
				logging.ErrorErr(err, context.Background())
				continue // We do not want to stop even if something goes wrong
			}
		}
	}()
	return retChan, nil
}
