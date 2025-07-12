package database

import (
	"fmt"
	"log"
	"time"

	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	database := &Database{DB: db}

	// Auto migrate tables
	if err := database.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed initial data
	if err := database.Seed(); err != nil {
		return nil, fmt.Errorf("failed to seed database: %w", err)
	}

	return database, nil
}

func (d *Database) Migrate() error {
	log.Println("Running database migrations...")

	return d.DB.AutoMigrate(
		&models.User{},
		&models.Achievement{},
		&models.UserAchievement{},
		&models.Garden{},
		&models.PlantType{},
		&models.Plant{},
		&models.Weather{},
		&models.WeatherForecast{},
	)
}

func (d *Database) Seed() error {
	log.Println("Seeding database with initial data...")

	// Seed achievements
	achievements := []models.Achievement{
		{
			Name:        "First Garden",
			Description: "Create your first garden",
			Icon:        "🌱",
			Points:      10,
			Category:    "gardening",
		},
		{
			Name:        "Plant Master",
			Description: "Plant 10 different types of plants",
			Icon:        "🌿",
			Points:      25,
			Category:    "gardening",
		},
		{
			Name:        "Harvest King",
			Description: "Harvest 50 plants",
			Icon:        "👑",
			Points:      50,
			Category:    "harvesting",
		},
		{
			Name:        "Weather Watcher",
			Description: "Check weather 10 times",
			Icon:        "🌤️",
			Points:      15,
			Category:    "weather",
		},
		{
			Name:        "Level 10",
			Description: "Reach level 10",
			Icon:        "⭐",
			Points:      100,
			Category:    "progression",
		},
	}

	for _, achievement := range achievements {
		var existing models.Achievement
		if err := d.DB.Where("name = ?", achievement.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := d.DB.Create(&achievement).Error; err != nil {
					return fmt.Errorf("failed to seed achievement %s: %w", achievement.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check achievement %s: %w", achievement.Name, err)
			}
		}
	}

	// Seed plant types
	plantTypes := []models.PlantType{
		{
			Name:            "Tomato",
			Description:     "A juicy red tomato that grows well in warm weather",
			Icon:            "🍅",
			GrowthTime:      120, // 2 hours
			WaterNeeds:      60,
			FertilizerNeeds: 20,
			Yield:           3,
			HarvestValue:    15,
			ExperienceValue: 10,
			MinLevel:        1,
			Season:          "summer",
			Weather:         "sunny",
			Rarity:          "common",
		},
		{
			Name:            "Carrot",
			Description:     "An orange root vegetable that grows underground",
			Icon:            "🥕",
			GrowthTime:      90, // 1.5 hours
			WaterNeeds:      50,
			FertilizerNeeds: 10,
			Yield:           2,
			HarvestValue:    12,
			ExperienceValue: 8,
			MinLevel:        1,
			Season:          "spring",
			Weather:         "all",
			Rarity:          "common",
		},
		{
			Name:            "Lettuce",
			Description:     "A leafy green vegetable that grows quickly",
			Icon:            "🥬",
			GrowthTime:      60, // 1 hour
			WaterNeeds:      70,
			FertilizerNeeds: 5,
			Yield:           1,
			HarvestValue:    8,
			ExperienceValue: 5,
			MinLevel:        1,
			Season:          "spring",
			Weather:         "cloudy",
			Rarity:          "common",
		},
		{
			Name:            "Strawberry",
			Description:     "A sweet red berry that requires careful tending",
			Icon:            "🍓",
			GrowthTime:      180, // 3 hours
			WaterNeeds:      80,
			FertilizerNeeds: 30,
			Yield:           2,
			HarvestValue:    25,
			ExperienceValue: 15,
			MinLevel:        3,
			Season:          "spring",
			Weather:         "sunny",
			Rarity:          "uncommon",
		},
		{
			Name:            "Golden Apple",
			Description:     "A rare golden apple with magical properties",
			Icon:            "🍎",
			GrowthTime:      360, // 6 hours
			WaterNeeds:      90,
			FertilizerNeeds: 50,
			Yield:           1,
			HarvestValue:    100,
			ExperienceValue: 50,
			MinLevel:        10,
			Season:          "autumn",
			Weather:         "sunny",
			Rarity:          "legendary",
		},
	}

	for _, plantType := range plantTypes {
		var existing models.PlantType
		if err := d.DB.Where("name = ?", plantType.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := d.DB.Create(&plantType).Error; err != nil {
					return fmt.Errorf("failed to seed plant type %s: %w", plantType.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check plant type %s: %w", plantType.Name, err)
			}
		}
	}

	log.Println("Database seeding completed successfully")
	return nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
