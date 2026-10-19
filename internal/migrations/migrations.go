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
	// Plants from before grower_id existed belong to their garden's owner.
	// AutoMigrate cannot add a NOT NULL column to a table with rows.
	if db.Migrator().HasTable("plants") && !db.Migrator().HasColumn("plants", "grower_id") {
		for _, statement := range []string{
			`ALTER TABLE plants ADD COLUMN grower_id uuid`,
			`UPDATE plants SET grower_id = gardens.user_id FROM gardens WHERE gardens.id = plants.garden_id`,
			`ALTER TABLE plants ALTER COLUMN grower_id SET NOT NULL`,
		} {
			if err := db.Exec(statement).Error; err != nil {
				return err
			}
		}
	}

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

	// Seed plant types, updating existing ones to the current catalog.
	for _, plantType := range plantTypes {
		var existing garden.PlantType
		if err := db.Where(garden.PlantType{Name: plantType.Name}).Assign(plantType).FirstOrCreate(&existing).Error; err != nil {
			return fmt.Errorf("failed to seed plant type %s: %w", plantType.Name, err)
		}
	}

	log.Println("Database seeding completed successfully")
	return nil
}

// plantTypes is the seed catalog. Each harvest returns the seed price plus a
// profit of growth minutes × a rate that rises with the unlock level: 0.20
// coins per minute at level 1, then 0.05 more per level up to 0.45 at level 6.
// HarvestValue is that revenue divided by Yield, rounded.
var plantTypes = []garden.PlantType{
	{
		Name:         "Lettuce",
		Description:  "A leafy green vegetable that grows quickly",
		Icon:         "🥬",
		GrowthTime:   30,
		Yield:        1,
		HarvestValue: 16,
		SeedPrice:    10,
		UnlockLevel:  1,
	},
	{
		Name:         "Sunflower",
		Description:  "A tall plant with a large flowering head that follows the sun",
		Icon:         "🌻",
		GrowthTime:   75,
		Yield:        1,
		HarvestValue: 27,
		SeedPrice:    12,
		UnlockLevel:  1,
	},
	{
		Name:         "Spinach",
		Description:  "A leafy green rich in iron and vitamins",
		Icon:         "🥬",
		GrowthTime:   40,
		Yield:        1,
		HarvestValue: 22,
		SeedPrice:    14,
		UnlockLevel:  1,
	},
	{
		Name:         "Carrot",
		Description:  "An orange root vegetable that grows underground",
		Icon:         "🥕",
		GrowthTime:   45,
		Yield:        2,
		HarvestValue: 12,
		SeedPrice:    15,
		UnlockLevel:  1,
	},
	{
		Name:         "Onion",
		Description:  "A pungent bulb vegetable used as a base in many dishes",
		Icon:         "🧅",
		GrowthTime:   65,
		Yield:        5,
		HarvestValue: 6,
		SeedPrice:    16,
		UnlockLevel:  2,
	},
	{
		Name:         "Cucumber",
		Description:  "A crunchy vine plant that bears refreshing green fruits",
		Icon:         "🥒",
		GrowthTime:   50,
		Yield:        3,
		HarvestValue: 10,
		SeedPrice:    18,
		UnlockLevel:  2,
	},
	{
		Name:         "Broccoli",
		Description:  "A green vegetable with a tree-like head packed with nutrients",
		Icon:         "🥦",
		GrowthTime:   55,
		Yield:        2,
		HarvestValue: 16,
		SeedPrice:    18,
		UnlockLevel:  2,
	},
	{
		Name:         "Tomato",
		Description:  "A juicy red tomato that grows well in warm weather",
		Icon:         "🍅",
		GrowthTime:   60,
		Yield:        3,
		HarvestValue: 12,
		SeedPrice:    20,
		UnlockLevel:  2,
	},
	{
		Name:         "Peas",
		Description:  "Small green legumes that grow in pods along vines",
		Icon:         "🫘",
		GrowthTime:   50,
		Yield:        4,
		HarvestValue: 9,
		SeedPrice:    20,
		UnlockLevel:  3,
	},
	{
		Name:         "Potato",
		Description:  "A starchy tuber that grows underground",
		Icon:         "🥔",
		GrowthTime:   80,
		Yield:        5,
		HarvestValue: 9,
		SeedPrice:    22,
		UnlockLevel:  3,
	},
	{
		Name:         "Garlic",
		Description:  "A flavorful bulb used as seasoning with a strong aroma",
		Icon:         "🧄",
		GrowthTime:   90,
		Yield:        6,
		HarvestValue: 9,
		SeedPrice:    24,
		UnlockLevel:  3,
	},
	{
		Name:         "Corn",
		Description:  "Tall stalks that produce kernels used as a staple food",
		Icon:         "🌽",
		GrowthTime:   70,
		Yield:        4,
		HarvestValue: 12,
		SeedPrice:    25,
		UnlockLevel:  4,
	},
	{
		Name:         "Pepper",
		Description:  "A spicy fruit available in various heat levels",
		Icon:         "🌶️",
		GrowthTime:   65,
		Yield:        2,
		HarvestValue: 25,
		SeedPrice:    28,
		UnlockLevel:  4,
	},
	{
		Name:         "Eggplant",
		Description:  "A purple vegetable with a smooth, glossy skin",
		Icon:         "🍆",
		GrowthTime:   85,
		Yield:        3,
		HarvestValue: 20,
		SeedPrice:    30,
		UnlockLevel:  4,
	},
	{
		Name:         "Watermelon",
		Description:  "A large fruit with a sweet, juicy interior",
		Icon:         "🍉",
		GrowthTime:   80,
		Yield:        1,
		HarvestValue: 64,
		SeedPrice:    32,
		UnlockLevel:  5,
	},
	{
		Name:         "Blueberry",
		Description:  "A small, sweet berry with a deep blue color",
		Icon:         "🫐",
		GrowthTime:   85,
		Yield:        2,
		HarvestValue: 34,
		SeedPrice:    34,
		UnlockLevel:  5,
	},
	{
		Name:         "Strawberry",
		Description:  "A sweet red berry that requires careful tending",
		Icon:         "🍓",
		GrowthTime:   90,
		Yield:        2,
		HarvestValue: 36,
		SeedPrice:    35,
		UnlockLevel:  5,
	},
	{
		Name:         "Pumpkin",
		Description:  "A large squash used for pies and decorations",
		Icon:         "🎃",
		GrowthTime:   100,
		Yield:        1,
		HarvestValue: 85,
		SeedPrice:    40,
		UnlockLevel:  6,
	},
}
