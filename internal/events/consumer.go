package events

import (
	"context"
	"encoding/json"

	"github.com/Kaese72/device-store/eventmodels"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/logging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventsConsumer struct {
	connection         *amqp.Connection
	deviceUpdatesTopic string
}

func NewEventsConsumer(conf config.EventConfig) (*EventsConsumer, error) {
	conn, err := amqp.Dial(conf.ConnectionString)
	if err != nil {
		return nil, err
	}
	return &EventsConsumer{connection: conn, deviceUpdatesTopic: conf.DeviceUpdatesTopic}, nil
}

func (h *EventsConsumer) Close() error {
	if h.connection != nil {
		return h.connection.Close()
	}
	return nil
}

// DeviceUpdates returns a channel on which we get device updates on
// from the message queue. The channel is closed when the connection is closed.
func (h *EventsConsumer) DeviceUpdates(ctx context.Context) (<-chan eventmodels.DeviceAttributeUpdate, error) {
	ch, err := h.connection.Channel()
	if err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(
		h.deviceUpdatesTopic, // name
		false,                // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer, auto-assign identifier
		true,   // auto-ack
		false,  // Not exclusive. Every instance will be consuming this message
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}
	out := make(chan eventmodels.DeviceAttributeUpdate)
	go func() {
		defer ch.Close()
		clientDone := ctx.Done()
	EVENT_LOOP:
		for {
			select {
			case <-clientDone:
				// Client is done, close the channel and return
				break EVENT_LOOP
			case msg, ok := <-msgs:
				if !ok {
					// Channel is closed, close the output channel and return
					break EVENT_LOOP
				}
				// Received a message, continue processing
				var update eventmodels.DeviceAttributeUpdate
				err := json.Unmarshal(msg.Body, &update)
				if err != nil {
					logging.ErrorErr(err, ctx, nil)
					// Continue processing the next message even though this one failed
					break
				}
				// FIXME log here
				out <- update
			}
		}
		// Cleanup and terminate
		close(out)
	}()
	return out, nil
}
