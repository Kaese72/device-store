package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence"
	"github.com/Kaese72/device-store/internal/server"
	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla"
	_ "go.elastic.co/apm/module/apmsql/mysql"
)

func main() {
	// # Viper configuration

	persistence, err := persistence.NewDevicePersistenceDB(config.Loaded.Database)
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}

	adapterAttendant := adapterattendant.NewAdapterAttendant(config.Loaded.AdapterAttendant)
	logging.Info("Successfully contacted database", context.Background())

	router := mux.NewRouter()
	apmgorilla.Instrument(router)

	restRouter := router.PathPrefix("/device-store/v0/").Subrouter()
	webapp := server.NewWebApp(persistence, adapterAttendant)

	// Devices
	restRouter.HandleFunc("/devices", webapp.GetDevices).Methods("GET")
	restRouter.HandleFunc("/devices", webapp.PostDevice).Methods("POST")
	restRouter.HandleFunc("/devices/{storeDeviceIdentifier:[0-9]+}/capabilities/{capabilityID}", webapp.TriggerDeviceCapability).Methods("POST")
	// Groups
	restRouter.HandleFunc("/groups", webapp.GetGroups).Methods("GET")
	restRouter.HandleFunc("/groups", webapp.PostGroup).Methods("POST")
	restRouter.HandleFunc("/groups/{storeGroupIdentifier:[0-9]+}/capabilities/{capabilityID}", webapp.TriggerGroupCapability).Methods("POST")

	server := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	if err := server.ListenAndServe(); err != nil {
		logging.Error(err.Error(), context.TODO())
		return
	}
}
