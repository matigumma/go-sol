package app

import (
	"sync"
	"tg_reader_bot/internal/config"
	"tg_reader_bot/internal/telegram"

	"gorm.io/gorm"
)

type Container struct {
	Config   *config.ConfigStructure
	Database *gorm.DB
	Client   *telegram.TGClient
}

var (
	container *Container
	once      sync.Once
)

func (c *Container) Init(config *config.ConfigStructure, database *gorm.DB) {
	container.Config = config
	container.Database = database
}

func GetContainer() *Container {
	once.Do(func() {
		container = &Container{}
	})
	return container
}

func GetDatabase() *gorm.DB {
	return GetContainer().Database
}

func GetConfig() *config.ConfigStructure {
	return GetContainer().Config
}

func GetClient() *telegram.TGClient {
	return GetContainer().Client
}
