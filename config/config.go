package config

import (
	"errors"
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

type HTTPConfig struct {
	Address string `json:"address" mapstructure:"address"`
	Port    int    `json:"port" mapstructure:"port"`
}

func (conf HTTPConfig) Validate() error {
	if len(conf.Address) == 0 {
		return errors.New("need to supply a http listen address")
	}

	if conf.Port == 0 {
		return errors.New("need to supply a http listen port")
	}
	return nil
}

type Config struct {
	Database   DatabaseConfig `json:"database" mapstructure:"database"`
	HTTPConfig HTTPConfig     `json:"http-server" mapstructure:"http-server"`
}

func (conf *Config) PopulateExample() {
	conf.Database = DatabaseConfig{
		MongoDB: &MongoDBConfig{
			ConnectionString: "localhost:27017",
		},
	}
	conf.HTTPConfig = HTTPConfig{
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
