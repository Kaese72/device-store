package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
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
	// ingestPersistence, err := persistence.NewIngestPersistenceDB(config.Loaded.Database)
	// if err != nil {
	// 	logging.Error(err.Error(), context.Background())
	// 	os.Exit(1)
	// }

	router := mux.NewRouter()
	apmgorilla.Instrument(router)

	// REST WebApp
	restRouter := router.PathPrefix("/device-store/v0/").Subrouter()
	adapterAttendant := adapterattendant.NewAdapterAttendant(config.Loaded.AdapterAttendant)
	restWebapp := restwebapp.NewWebApp(dbPersistence, adapterAttendant)
	restRouter.HandleFunc("/devices", restWebapp.GetDevices).Methods("GET")
	restRouter.HandleFunc("/groups", restWebapp.GetGroups).Methods("GET")
	restRouter.HandleFunc("/devices/{storeDeviceIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerDeviceCapability).Methods("POST")
	restRouter.HandleFunc("/groups/{storeGroupIdentifier:[0-9]+}/capabilities/{capabilityID}", restWebapp.TriggerGroupCapability).Methods("POST")

	// Ingest WebApp
	injestRouter := router.PathPrefix("/device-store/v0/").Subrouter()
	ingestWebapp := ingestwebapp.NewWebApp(dbPersistence)
	injestRouter.HandleFunc("/devices", ingestWebapp.PostDevice).Methods("POST")
	injestRouter.HandleFunc("/groups", ingestWebapp.PostGroup).Methods("POST")

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
