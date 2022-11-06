package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kaese72/device-store/config"
	"github.com/Kaese72/device-store/database"
	"github.com/Kaese72/device-store/rest"
	"github.com/Kaese72/sdup-lib/logging"
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

	var conf config.Config
	err := myVip.Unmarshal(&conf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err := myVip.WriteConfigAs("./config.used.yaml"); err != nil {
		logging.Error(err.Error())
		return
	}

	if err := conf.Validate(); err != nil {
		logging.Error(err.Error())
		conf.PopulateExample()
		res, err := json.MarshalIndent(conf, "", "   ")
		if err != nil {
			logging.Error(err.Error())
			return
		}
		fmt.Print(string(res))
		return
	}

	persistence, err := database.NewDevicePersistenceDB(conf.Database)
	if err != nil {
		logging.Error(err.Error())
		return
	}
	logging.Info("Successfully contacted database")
	rest.PersistenceAPIListenAndServe(conf.HTTPConfig, persistence)
}
