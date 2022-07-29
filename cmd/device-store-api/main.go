package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Kaese72/device-store/config"
	"github.com/Kaese72/device-store/database"
	"github.com/Kaese72/device-store/rest"
	"github.com/Kaese72/sdup-lib/logging"
)

type Config struct {
	Database   config.DatabaseConfig `json:"database"`
	HTTPConfig config.HTTPConfig     `json:"http-server"`
}

func (conf *Config) PopulateExample() {
	conf.Database = config.DatabaseConfig{
		MongoDB: &config.MongoDBConfig{
			ConnectionString: "localhost:27017",
		},
	}
	conf.HTTPConfig = config.HTTPConfig{
		Address: "localhost",
		Port:    8080,
	}
}

func (conf Config) Validate() error {
	if err := conf.Database.Validate(); err != nil {
		return err
	}
	if err := conf.HTTPConfig.Validate(); err != nil {
		return err
	}
	return nil
}

func ReadConfig() (Config, error) {
	conf := Config{}
	if _, err := os.Stat("./settings.json"); err == nil {
		file, err := os.Open("./settings.json")
		if err != nil {
			logging.Error(fmt.Sprintf("Unable to open local settings file, %s", err.Error()))
			return conf, err
		}
		if err := json.NewDecoder(file).Decode(&conf); err != nil {
			logging.Error(err.Error())
			return conf, err
		}

	} else {
		if err := json.NewDecoder(os.Stdin).Decode(&conf); err != nil {
			logging.Error(err.Error())
			return conf, err
		}
	}

	return conf, nil
}

func main() {
	conf, err := ReadConfig()
	if err != nil {
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
