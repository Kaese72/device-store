package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Kaese72/huemie-lib/logging"
	"github.com/spf13/viper"
)

type MongoDBConfig struct {
	ConnectionString string `json:"connection-string" mapstructure:"connection-string"`
	DbName           string `json:"db-name" mapstructure:"db-name"`
}

func (conf MongoDBConfig) Validate() error {
	if len(conf.ConnectionString) == 0 {
		return errors.New("need to supply a mongodb connection string")
	}
	return nil
}

type DatabaseConfig struct {
	MongoDB MongoDBConfig `json:"mongodb" mapstructure:"mongodb"`
}

func (conf DatabaseConfig) Validate() error {
	if conf.MongoDB.Validate() == nil {
		// MongoDB validation passed, indicating that there is at least one valid config
		return nil
	}
	return errors.New("need to supply at least one database backend")
}

type AdapterAttendantConfig struct {
	URL string `json:"url" mapstructure:"url"`
}

func (conf AdapterAttendantConfig) Validate() error {
	if conf.URL == "" {
		return errors.New("must supply adapter base URL")
	}
	return nil
}

type Config struct {
	Database         DatabaseConfig         `json:"database" mapstructure:"database"`
	AdapterAttendant AdapterAttendantConfig `json:"adapter-attendant" mapstructure:"adapter-attendant"`
}

func (conf *Config) PopulateExample() {
	conf.Database = DatabaseConfig{
		MongoDB: MongoDBConfig{
			ConnectionString: "localhost:27017",
		},
	}
	conf.AdapterAttendant = AdapterAttendantConfig{
		URL: "http://somehost:8080/rest/v0",
	}
}

func (conf Config) Validate() error {
	if err := conf.Database.Validate(); err != nil {
		return err
	}
	if err := conf.AdapterAttendant.Validate(); err != nil {
		return err
	}
	return nil
}

var Loaded Config

func init() {
	// We have elected to no use AutomaticEnv() because of https://github.com/spf13/viper/issues/584
	// myVip.AutomaticEnv()
	// Set replaces to allow keys like "database.mongodb.connection-string"
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	// # Database configuration, if left out, assume no mongo configuration
	viper.BindEnv("database.mongodb.connection-string")
	viper.BindEnv("database.mongodb.db-name")
	viper.SetDefault("database.mongodb.db-name", "huemie")

	// # Device attendant
	viper.BindEnv("adapter-attendant.url")

	// # Logging
	viper.BindEnv("logging.stdout")
	viper.SetDefault("logging.stdout", true)
	viper.BindEnv("logging.http.url")
	err := viper.Unmarshal(&Loaded)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err := Loaded.Validate(); err != nil {
		logging.Error(err.Error())
		os.Exit(1)
	}
}
