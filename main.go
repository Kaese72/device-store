package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Kaese72/device-store/internal/adapterattendant"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/database"
	"github.com/Kaese72/device-store/rest"
	"github.com/Kaese72/huemie-lib/logging"
	"github.com/spf13/viper"
)

func main() {
	// # Viper configuration
	myVip := viper.New()
	// We have elected to no use AutomaticEnv() because of https://github.com/spf13/viper/issues/584
	// myVip.AutomaticEnv()
	// Set replaces to allow keys like "database.mongodb.connection-string"
	myVip.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// # Configuration file configuration
	myVip.SetConfigName("config")
	myVip.AddConfigPath(".")
	myVip.AddConfigPath("99_local")
	myVip.AddConfigPath("/etc/device-store-api/")
	if err := myVip.ReadInConfig(); err != nil {
		logging.Error(err.Error())
	}

	// # Default values where appropriate
	// API configuration
	myVip.BindEnv("http-server.address")
	myVip.SetDefault("http-server.address", "0.0.0.0")

	myVip.BindEnv("http-server.port")
	myVip.SetDefault("http-server.port", 8080)
	// # Database configuration, if left out, assume no mongo configuration
	myVip.BindEnv("database.mongodb.connection-string")
	myVip.BindEnv("database.mongodb.db-name")
	myVip.SetDefault("database.mongodb.db-name", "huemie")

	// # Device attendant
	myVip.BindEnv("adapter-attendant.url")

	// # Logging
	myVip.BindEnv("logging.stdout")
	myVip.SetDefault("logging.stdout", true)
	myVip.BindEnv("logging.http.url")

	var conf config.Config
	err := myVip.Unmarshal(&conf)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := myVip.WriteConfigAs("./config.used.yaml"); err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}

	if err := conf.Validate(); err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}

	persistence, err := database.NewDevicePersistenceDB(conf.Database)
	if err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}
	adapterAttendant := adapterattendant.NewAdapterAttendant(conf.AdapterAttendant)
	logging.Info("Successfully contacted database")
	rest.PersistenceAPIListenAndServe(conf.HTTPConfig, persistence, adapterAttendant)
}
