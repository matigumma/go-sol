package database

import (
	"tg_reader_bot/internal/models"

	// "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(query string) (*gorm.DB, error) {
	// db, err := gorm.Open(mysql.Open(query), &gorm.Config{})
	db, err := gorm.Open(sqlite.Open(query), &gorm.Config{}) // Update to use SQLite

	db.AutoMigrate(
		&models.User{},
		&models.Peer{},
		&models.KeyWords{},
	)

	return db, err
}
