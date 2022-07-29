package config

import (
	"errors"
)

type MongoDBConfig struct {
	ConnectionString string `json:"connection-string"`
}

func (conf MongoDBConfig) Validate() error {
	if len(conf.ConnectionString) == 0 {
		return errors.New("need to supply a mongodb connection string")
	}
	return nil
}

type DatabaseConfig struct {
	MongoDB *MongoDBConfig `json:"mongodb"`
}

func (conf DatabaseConfig) Validate() error {
	if conf.MongoDB == nil {
		return errors.New("need to supply at least one database backend")
	}

	return nil
}

type HTTPConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
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
