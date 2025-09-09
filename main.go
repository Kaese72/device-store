package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Kaese72/device-store/eventmodels"
	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/events"
	"github.com/Kaese72/device-store/internal/ingestwebapp"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence/mariadb"
	"github.com/Kaese72/device-store/internal/restwebapp"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humamux"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/gorilla/mux"

	_ "go.elastic.co/apm/module/apmsql/mysql"
)

func main() {
	if err := config.Loaded.Validate(); err != nil {
		logging.Error(err.Error(), context.TODO())
		os.Exit(1)
	}
	// # Viper configuration
	dbPersistence, err := mariadb.NewMariadbPersistence(config.Loaded.Database)
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}
	eventsHandler, err := events.NewEventsConsumer(config.Loaded.Event)
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}
	defer eventsHandler.Close()
	deviceUpdates, err := eventsHandler.DeviceUpdates(context.Background())
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}

	eventsProducer, err := events.NewEventsProducer(config.Loaded.Event)
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}
	defer eventsProducer.Close()
	deviceUpdateChan, err := eventsProducer.ProduceDeviceUpdates()
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}

	adapterAttendant := adapterattendant.NewAdapterAttendant(config.Loaded.AdapterAttendant)
	restWebapp := restwebapp.NewWebApp(dbPersistence, adapterAttendant, deviceUpdates)
	ingestWebapp := ingestwebapp.NewWebApp(dbPersistence, deviceUpdateChan)

	// Create Huma API
	router := mux.NewRouter()
	humaConfig := huma.DefaultConfig("device-store", "1.0.0")
	humaConfig.OpenAPIPath = "/device-store/openapi"
	humaConfig.DocsPath = "/device-store/docs"
	api := humamux.New(router, humaConfig)

	// Device Store endpoints
	huma.Get(api, "/device-store/v0/devices", restWebapp.GetDevices)
	huma.Get(api, "/device-store/v0/devices/{storeDeviceIdentifier:[0-9]+}", restWebapp.GetDevice)
	huma.Post(api, "/device-store/v0/devices/{storeDeviceIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerDeviceCapability)

	sse.Register(api, huma.Operation{
		OperationID: "device_updates",
		Method:      http.MethodGet,
		Path:        "/device-store/v0/devices/events",
		Summary:     "Server sent events for devices",
	}, map[string]any{
		// Mapping of event type name to Go struct for that event.
		"update": eventmodels.DeviceAttributeUpdate{},
	}, restWebapp.StreamDeviceUpdates)

	huma.Get(api, "/device-store/v0/audits/attributes", restWebapp.GetAttributeAudits)

	huma.Get(api, "/device-store/v0/groups", restWebapp.GetGroups)
	huma.Get(api, "/device-store/v0/groups/{storeGroupIdentifier:[0-9]+}", restWebapp.GetGroup)
	huma.Post(api, "/device-store/v0/groups/{storeGroupIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerGroupCapability)

	// // Device Ingest endpoints
	// Uncomment below if ingestWebapp is enabled and Huma-compatible
	huma.Post(api, "/device-ingest/v0/devices", ingestWebapp.PostDevice)
	huma.Post(api, "/device-ingest/v0/groups", ingestWebapp.PostGroup)

	// Start the server
	if err := http.ListenAndServe(":8080", router); err != nil {
		logging.Error(err.Error(), context.TODO())
	}
}
