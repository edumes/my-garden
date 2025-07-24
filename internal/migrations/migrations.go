package migrations

import (
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// First migrate auth models
	if err := db.AutoMigrate(
		&auth.User{},
		&auth.UserAchievement{},
		&auth.Achievement{},
	); err != nil {
		return err
	}

	// Then migrate garden models
	if err := db.AutoMigrate(
		&garden.Garden{},
		&garden.Plant{},
		&garden.PlantType{},
		&garden.GardenShare{},
		&garden.GardenAccessLink{},
	); err != nil {
		return err
	}

	// Finally migrate audit models
	if err := audit.RegisterModels(db); err != nil {
		return err
	}

	return nil
}
