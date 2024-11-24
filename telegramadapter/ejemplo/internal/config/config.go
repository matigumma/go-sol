package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/go-faster/errors"
)

type ConfigStructure struct {
	AppID       int    `json:"app_id"`
	AppHash     string `json:"app_hash"`
	APIToken    string `json:"api_token"`
	PhoneNumber string `json:"phone_number"`
	DbSettings  struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"db_settings"`
}

func Init() (*ConfigStructure, error) {
	jsonFile, err := os.ReadFile("configs/config.json")
	if err != nil {

		return nil, errors.Wrap(err, "error on reading config")
	}
	var config ConfigStructure
	err = json.Unmarshal(jsonFile, &config)
	if err != nil {
		return nil, errors.Wrap(err, "error on parsing config file")
	}
	return &config, nil
}

func (config *ConfigStructure) GetDatabaseQuery() string {
	query := url.URL{
		User:     url.UserPassword(config.DbSettings.Username, config.DbSettings.Password),
		Host:     fmt.Sprintf("tcp(%s)", config.DbSettings.Host),
		Path:     config.DbSettings.Database,
		RawQuery: "charset=utf8mb4&parseTime=True",
	}
	return query.String()[2:]
}
