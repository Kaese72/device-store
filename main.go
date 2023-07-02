package main

import (
	"os"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/database"
	"github.com/Kaese72/device-store/internal/server"
	"github.com/Kaese72/huemie-lib/logging"
)

func main() {
	// # Viper configuration

	persistence, err := database.NewDevicePersistenceDB(config.Loaded.Database)
	if err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}
	adapterAttendant := adapterattendant.NewAdapterAttendant(config.Loaded.AdapterAttendant)
	logging.Info("Successfully contacted database")
	server.PersistenceAPIListenAndServe(persistence, adapterAttendant)
}
