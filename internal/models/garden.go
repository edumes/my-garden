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

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

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
	GrowthProgress float64    `json:"growth_progress" gorm:"default:0"` // 0-100

	// Timestamps
	PlantedAt   time.Time  `json:"planted_at"`
	HarvestedAt *time.Time `json:"harvested_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

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
)

// PlantType represents available plant types
type PlantType struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`

	// Growth properties
	GrowthTime   int `json:"growth_time" gorm:"not null"` // in minutes
	Yield        int `json:"yield" gorm:"default:1"`      // items per harvest
	HarvestValue int
}
