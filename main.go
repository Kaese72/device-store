package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/events"
	"github.com/Kaese72/device-store/internal/ingestwebapp"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence/mariadb"
	"github.com/Kaese72/device-store/internal/restwebapp"
	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla"
	_ "go.elastic.co/apm/module/apmsql/mysql"
)

func main() {
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

	router := mux.NewRouter()
	apmgorilla.Instrument(router)

	// REST WebApp
	adapterAttendant := adapterattendant.NewAdapterAttendant(config.Loaded.AdapterAttendant)
	restWebapp := restwebapp.NewWebApp(dbPersistence, adapterAttendant, eventsHandler)
	restRouter := router.PathPrefix("/device-store/v0/").Subrouter()
	restRouter.HandleFunc("/devices", restWebapp.GetDevices).Methods("GET")
	restRouter.HandleFunc("/devices/events", restWebapp.StreamDeviceUpdates).Methods("GET")
	restRouter.HandleFunc("/groups", restWebapp.GetGroups).Methods("GET")
	restRouter.HandleFunc("/devices/{storeDeviceIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerDeviceCapability).Methods("POST")
	restRouter.HandleFunc("/groups/{storeGroupIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerGroupCapability).Methods("POST")

	// Ingest WebApp
	ingestWebapp := ingestwebapp.NewWebApp(dbPersistence, deviceUpdateChan)
	ingestRouter := router.PathPrefix("/device-ingest/v0/").Subrouter()
	ingestRouter.HandleFunc("/devices", ingestWebapp.PostDevice).Methods("POST")
	ingestRouter.HandleFunc("/groups", ingestWebapp.PostGroup).Methods("POST")

	logging.Info("Successfully contacted database", context.Background())

	server := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	if err := server.ListenAndServe(); err != nil {
		logging.Error(err.Error(), context.TODO())
		return
	}
}
