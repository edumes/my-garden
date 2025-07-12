package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Garden struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`

	// Garden properties
	Size            int `json:"size" gorm:"default:9"`             // 3x3 grid
	SoilQuality     int `json:"soil_quality" gorm:"default:50"`    // 0-100
	WaterLevel      int `json:"water_level" gorm:"default:50"`     // 0-100
	FertilizerLevel int `json:"fertilizer_level" gorm:"default:0"` // 0-100

	// Garden upgrades
	HasSprinkler  bool `json:"has_sprinkler" gorm:"default:false"`
	HasGreenhouse bool `json:"has_greenhouse" gorm:"default:false"`
	HasComposter  bool `json:"has_composter" gorm:"default:false"`

	// Timestamps
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastWateredAt    *time.Time `json:"last_watered_at"`
	LastFertilizedAt *time.Time `json:"last_fertilized_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Plants []Plant `json:"plants,omitempty" gorm:"foreignKey:GardenID"`
}

func (g *Garden) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

// Plant represents a plant in a garden
type Plant struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GardenID    uuid.UUID `json:"garden_id" gorm:"type:uuid;not null"`
	PlantTypeID uuid.UUID `json:"plant_type_id" gorm:"type:uuid;not null"`

	// Position in garden grid (0-8 for 3x3 grid)
	Position int `json:"position" gorm:"not null"`

	// Plant state
	Stage          PlantStage `json:"stage" gorm:"default:'seed'"`
	Health         int        `json:"health" gorm:"default:100"`        // 0-100
	WaterLevel     int        `json:"water_level" gorm:"default:50"`    // 0-100
	GrowthProgress float64    `json:"growth_progress" gorm:"default:0"` // 0-100

	// Timestamps
	PlantedAt        time.Time  `json:"planted_at"`
	LastWateredAt    *time.Time `json:"last_watered_at"`
	LastFertilizedAt *time.Time `json:"last_fertilized_at"`
	HarvestedAt      *time.Time `json:"harvested_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Relationships
	Garden    Garden    `json:"garden" gorm:"foreignKey:GardenID"`
	PlantType PlantType `json:"plant_type" gorm:"foreignKey:PlantTypeID"`
}

func (p *Plant) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PlantStage represents the growth stage of a plant
type PlantStage string

const (
	PlantStageSeed        PlantStage = "seed"
	PlantStageSprout      PlantStage = "sprout"
	PlantStageGrowing     PlantStage = "growing"
	PlantStageMature      PlantStage = "mature"
	PlantStageHarvestable PlantStage = "harvestable"
	PlantStageWithered    PlantStage = "withered"
)

// PlantType represents available plant types
type PlantType struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`

	// Growth properties
	GrowthTime      int `json:"growth_time" gorm:"not null"`       // in minutes
	WaterNeeds      int `json:"water_needs" gorm:"default:50"`     // 0-100
	FertilizerNeeds int `json:"fertilizer_needs" gorm:"default:0"` // 0-100

	// Harvest properties
	Yield           int `json:"yield" gorm:"default:1"`            // items per harvest
	HarvestValue    int `json:"harvest_value" gorm:"default:10"`   // coins per item
	ExperienceValue int `json:"experience_value" gorm:"default:5"` // XP per harvest

	// Requirements
	MinLevel int    `json:"min_level" gorm:"default:1"`
	Season   string `json:"season"`  // spring, summer, autumn, winter, all
	Weather  string `json:"weather"` // sunny, cloudy, rainy, all

	// Rarity
	Rarity string `json:"rarity" gorm:"default:'common'"` // common, uncommon, rare, epic, legendary

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (pt *PlantType) BeforeCreate(tx *gorm.DB) error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	return nil
}
