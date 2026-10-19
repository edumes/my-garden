package db

import (
	"gorm.io/gorm"
)

type Database interface {
	GetDB() *gorm.DB
}

// AutoMigrate runs migrations for the given models
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
