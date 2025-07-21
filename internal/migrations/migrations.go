package migrations

import (
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// First migrate base models
	if err := db.AutoMigrate(
		&garden.Garden{},
		&garden.Plant{},
		&garden.GardenShare{},
		&garden.GardenAccessLink{},
	); err != nil {
		return err
	}

	// Then migrate audit models
	if err := audit.RegisterModels(db); err != nil {
		return err
	}

	return nil
}
