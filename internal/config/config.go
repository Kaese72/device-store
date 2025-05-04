package config

import (
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/Kaese72/device-store/internal/logging"
	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Host     string `json:"host" mapstructure:"host" `
	Port     int    `json:"port" mapstructure:"port"`
	User     string `json:"user" mapstructure:"user"`
	Password string `json:"password" mapstructure:"password"`
	Database string `json:"database" mapstructure:"database"`
}

func (conf DatabaseConfig) Validate() error {
	if conf.Host == "" {
		errors.New("must supply database host")
	}
	return nil
}

type EventConfig struct {
	DeviceUpdatesTopic string `json:"deviceUpdates" mapstructure:"deviceUpdates"`
	ConnectionString   string `json:"connectionString" mapstructure:"connectionString"`
}

func (conf EventConfig) Validate() error {
	if conf.DeviceUpdatesTopic == "" {
		return errors.New("must supply event device updates topic")
	}
	if conf.ConnectionString == "" {
		return errors.New("must supply event connection string")
	}
	return nil
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
	PurgeDB          bool                   `json:"purge-db"`
	Event            EventConfig            `json:"event" mapstructure:"event"`
}

func (conf Config) Validate() error {
	if err := conf.Database.Validate(); err != nil {
		return err
	}
	if err := conf.AdapterAttendant.Validate(); err != nil {
		return err
	}
	if err := conf.Event.Validate(); err != nil {
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
	// # Database configuration, if left out.
	viper.BindEnv("database.host")
	viper.BindEnv("database.port")
	viper.BindEnv("database.user")
	viper.BindEnv("database.password")
	viper.BindEnv("database.database")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.database", "devicestore")
	viper.BindEnv("purge-db")
	viper.SetDefault("purge-db", false)

	// # Device attendant
	viper.BindEnv("adapter-attendant.url")
	viper.SetDefault("adapter-attendant.url", "http://adapter-attendant:8080")

	// # Logging
	viper.BindEnv("logging.stdout")
	viper.SetDefault("logging.stdout", true)
	viper.BindEnv("logging.http.url")

	// Event
	viper.BindEnv("event.deviceUpdates")
	viper.SetDefault("event.deviceUpdates", "deviceUpdates")
	viper.BindEnv("event.connectionString")

	err := viper.Unmarshal(&Loaded)
	if err != nil {
		logging.Error(err.Error(), context.TODO())
		os.Exit(1)
	}
	if err := Loaded.Validate(); err != nil {
		logging.Error(err.Error(), context.TODO())
		os.Exit(1)
	}
}
