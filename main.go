package main

import (
	"context"
	"net/http"
	"os"

	"github.com/Kaese72/device-store/gql"
	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/database"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/server"
	"github.com/gorilla/mux"
	"go.elastic.co/apm/module/apmgorilla"
)

func main() {
	// # Viper configuration

	persistence, err := database.NewDevicePersistenceDB(config.Loaded.Database)
	if err != nil {
		logging.Error(err.Error(), context.Background())
		os.Exit(1)
	}

	adapterAttendant := adapterattendant.NewAdapterAttendant(config.Loaded.AdapterAttendant)
	logging.Info("Successfully contacted database", context.Background())

	router := mux.NewRouter()
	apmgorilla.Instrument(router)

	restRouter := router.PathPrefix("/device-store/").Subrouter()
	err = server.PersistenceAPIListenAndServe(restRouter, persistence, adapterAttendant)
	if err != nil {
		logging.Error(err.Error(), context.TODO())
		return
	}

	gqlRouter := router.PathPrefix("/device-store-gql/").Subrouter()
	err = gql.GraphQLListenAndServe(gqlRouter, persistence)
	if err != nil {
		logging.Error(err.Error(), context.TODO())
		return
	}

	server := &http.Server{
		Handler: router,
		Addr:    "0.0.0.0:8080",
	}

	if err := server.ListenAndServe(); err != nil {
		logging.Error(err.Error(), context.TODO())
		return
	}
}
