package main

import (
	"context"
	"fmt"
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
	"github.com/Kaese72/huemie-lib/middleware"
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

	adapterTrigger := adapterattendant.NewAdapterTrigger(config.Loaded.AdapterAttendant)
	restWebapp := restwebapp.NewWebApp(dbPersistence, adapterTrigger, deviceUpdates)
	ingestWebapp := ingestwebapp.NewWebApp(dbPersistence, deviceUpdateChan)

	pubKey, err := middleware.LoadPublicKeyFromFile(config.Loaded.Auth.RSAPublicKeyPath)
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}

	// Public router (device-store + device-ingest)
	publicRouter := mux.NewRouter()
	publicRouter.Use(middleware.UseTokenMiddleware(pubKey, "/device-ingest/", "/device-store/openapi", "/device-store/docs"))
	publicRouter.Use(ingestwebapp.DeviceIngestJWTMiddleware(config.Loaded.DeviceIngest.JWTSecret))
	publicHumaConfig := huma.DefaultConfig("device-store", "1.0.0")
	publicHumaConfig.OpenAPIPath = "/device-store/openapi"
	publicHumaConfig.DocsPath = "/device-store/docs"
	publicAPI := humamux.New(publicRouter, publicHumaConfig)

	huma.Get(publicAPI, "/device-store/v0/devices", restWebapp.GetDevices)
	huma.Get(publicAPI, "/device-store/v0/devices/{storeDeviceIdentifier:[0-9]+}", restWebapp.GetDevice)
	huma.Delete(publicAPI, "/device-store/v0/devices/{storeDeviceIdentifier:[0-9]+}", restWebapp.DeleteDevice)
	huma.Post(publicAPI, "/device-store/v0/devices/{storeDeviceIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerDeviceCapability)

	sse.Register(publicAPI, huma.Operation{
		OperationID: "device_updates",
		Method:      http.MethodGet,
		Path:        "/device-store/v0/devices/events",
		Summary:     "Server sent events for devices",
	}, map[string]any{
		"update": eventmodels.DeviceAttributeUpdate{},
	}, restWebapp.StreamDeviceUpdates)

	huma.Get(publicAPI, "/device-store/v0/audits/attributes", restWebapp.GetAttributeAudits)
	huma.Get(publicAPI, "/device-store/v0/devices/{storeDeviceIdentifier:[0-9]+}/capability-trigger-audits", restWebapp.GetDeviceCapabilityTriggerAudits)

	huma.Get(publicAPI, "/device-store/v0/groups", restWebapp.GetGroups)
	huma.Get(publicAPI, "/device-store/v0/groups/{storeGroupIdentifier:[0-9]+}", restWebapp.GetGroup)
	huma.Delete(publicAPI, "/device-store/v0/groups/{storeGroupIdentifier:[0-9]+}", restWebapp.DeleteGroup)
	huma.Post(publicAPI, "/device-store/v0/groups/{storeGroupIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerGroupCapability)
	huma.Get(publicAPI, "/device-store/v0/groups/{storeGroupIdentifier:[0-9]+}/capability-trigger-audits", restWebapp.GetGroupCapabilityTriggerAudits)

	huma.Post(publicAPI, "/device-ingest/v0/devices", ingestWebapp.PostDevice)
	huma.Post(publicAPI, "/device-ingest/v0/groups", ingestWebapp.PostGroup)

	// Internal router (device-store-internal) — no auth, restrict via NetworkPolicy
	internalRouter := mux.NewRouter()
	internalAPI := humamux.New(internalRouter, huma.DefaultConfig("device-store-internal", "1.0.0"))

	huma.Get(internalAPI, "/device-store-internal/v0/devices/{storeDeviceIdentifier:[0-9]+}", restWebapp.GetDevice)
	huma.Post(internalAPI, "/device-store-internal/v0/devices/{storeDeviceIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerDeviceCapability)
	huma.Post(internalAPI, "/device-store-internal/v0/groups/{storeGroupIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerGroupCapability)

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Loaded.InternalPort), internalRouter); err != nil {
			logging.Error(err.Error(), context.TODO())
			os.Exit(1)
		}
	}()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Loaded.PublicPort), publicRouter); err != nil {
		logging.Error(err.Error(), context.TODO())
	}
}
