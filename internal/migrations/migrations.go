package migrations

import (
	"fmt"
	"log"

	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/marketplace"
	"github.com/my-garden/api/internal/weather"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func AutoMigrate(db *gorm.DB) error {
	// migrate base models
	if err := db.AutoMigrate(
		&audit.AuditLog{},
		&auth.User{},
		&auth.Achievement{},
		&auth.UserAchievement{},
		&garden.Garden{},
		&garden.GardenShare{},
		&garden.GardenAccessLink{},
		&garden.PlantType{},
		&garden.Plant{},
		&garden.SeedInventory{},
		&weather.Weather{},
		&weather.WeatherForecast{},
		&marketplace.DeliveryRequest{},
		&marketplace.PlantListing{},
		&marketplace.Transaction{},
		&marketplace.Block{},
		&marketplace.Wallet{},
		&marketplace.ConsensusNode{},
	); err != nil {
		return err
	}

	return nil
}

func Seed(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	// Seed achievements
	achievements := []auth.Achievement{
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
		var existing auth.Achievement
		if err := db.Where("name = ?", achievement.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&achievement).Error; err != nil {
					return fmt.Errorf("failed to seed achievement %s: %w", achievement.Name, err)
				}
			} else {
				return fmt.Errorf("failed to check achievement %s: %w", achievement.Name, err)
			}
		}
	}

	// Seed plant types
	plantTypes := []garden.PlantType{
		{
			Name:         "Tomato",
			Description:  "A juicy red tomato that grows well in warm weather",
			Icon:         "🍅",
			GrowthTime:   60,
			Yield:        3,
			HarvestValue: 15,
			SeedPrice:    20,
		},
		{
			Name:         "Carrot",
			Description:  "An orange root vegetable that grows underground",
			Icon:         "🥕",
			GrowthTime:   45,
			Yield:        2,
			HarvestValue: 12,
			SeedPrice:    15,
		},
		{
			Name:         "Lettuce",
			Description:  "A leafy green vegetable that grows quickly",
			Icon:         "🥬",
			GrowthTime:   30,
			Yield:        1,
			HarvestValue: 8,
			SeedPrice:    10,
		},
		{
			Name:         "Strawberry",
			Description:  "A sweet red berry that requires careful tending",
			Icon:         "🍓",
			GrowthTime:   90,
			Yield:        2,
			HarvestValue: 25,
			SeedPrice:    35,
		},
		{
			Name:         "Corn",
			Description:  "Tall stalks that produce kernels used as a staple food",
			Icon:         "🌽",
			GrowthTime:   70,
			Yield:        4,
			HarvestValue: 20,
			SeedPrice:    25,
		},
		{
			Name:         "Potato",
			Description:  "A starchy tuber that grows underground",
			Icon:         "🥔",
			GrowthTime:   80,
			Yield:        5,
			HarvestValue: 18,
			SeedPrice:    22,
		},
		{
			Name:         "Cucumber",
			Description:  "A crunchy vine plant that bears refreshing green fruits",
			Icon:         "🥒",
			GrowthTime:   50,
			Yield:        3,
			HarvestValue: 14,
			SeedPrice:    18,
		},
		{
			Name:         "Pepper",
			Description:  "A spicy fruit available in various heat levels",
			Icon:         "🌶️",
			GrowthTime:   65,
			Yield:        2,
			HarvestValue: 17,
			SeedPrice:    28,
		},
		{
			Name:         "Pumpkin",
			Description:  "A large squash used for pies and decorations",
			Icon:         "🎃",
			GrowthTime:   100,
			Yield:        1,
			HarvestValue: 30,
			SeedPrice:    40,
		},
		{
			Name:         "Sunflower",
			Description:  "A tall plant with a large flowering head that follows the sun",
			Icon:         "🌻",
			GrowthTime:   75,
			Yield:        1,
			HarvestValue: 10,
			SeedPrice:    12,
		},
		{
			Name:         "Eggplant",
			Description:  "A purple vegetable with a smooth, glossy skin",
			Icon:         "🍆",
			GrowthTime:   85,
			Yield:        3,
			HarvestValue: 22,
			SeedPrice:    30,
		},
		{
			Name:         "Broccoli",
			Description:  "A green vegetable with a tree-like head packed with nutrients",
			Icon:         "🥦",
			GrowthTime:   55,
			Yield:        2,
			HarvestValue: 16,
			SeedPrice:    18,
		},
		{
			Name:         "Spinach",
			Description:  "A leafy green rich in iron and vitamins",
			Icon:         "🥬",
			GrowthTime:   40,
			Yield:        1,
			HarvestValue: 12,
			SeedPrice:    14,
		},
		{
			Name:         "Peas",
			Description:  "Small green legumes that grow in pods along vines",
			Icon:         "🫘",
			GrowthTime:   50,
			Yield:        4,
			HarvestValue: 15,
			SeedPrice:    20,
		},
		{
			Name:         "Watermelon",
			Description:  "A large fruit with a sweet, juicy interior",
			Icon:         "🍉",
			GrowthTime:   80,
			Yield:        1,
			HarvestValue: 28,
			SeedPrice:    32,
		},
		{
			Name:         "Onion",
			Description:  "A pungent bulb vegetable used as a base in many dishes",
			Icon:         "🧅",
			GrowthTime:   65,
			Yield:        5,
			HarvestValue: 14,
			SeedPrice:    16,
		},
		{
			Name:         "Garlic",
			Description:  "A flavorful bulb used as seasoning with a strong aroma",
			Icon:         "🧄",
			GrowthTime:   90,
			Yield:        6,
			HarvestValue: 20,
			SeedPrice:    24,
		},
		{
			Name:         "Blueberry",
			Description:  "A small, sweet berry with a deep blue color",
			Icon:         "🫐",
			GrowthTime:   85,
			Yield:        2,
			HarvestValue: 26,
			SeedPrice:    34,
		},
	}

	for _, plantType := range plantTypes {
		var existing garden.PlantType
		if err := db.Where("name = ?", plantType.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&plantType).Error; err != nil {
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
