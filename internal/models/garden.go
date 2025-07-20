package models

import (
	"encoding/json"
	"fmt"
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
	HarvestValue int `json:"harvest_value"`

	// Store properties
	SeedPrice int `json:"seed_price" gorm:"default:10"` // coins per seed
}

// SeedInventory represents user's purchased seeds
type SeedInventory struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	PlantTypeID uuid.UUID `json:"plant_type_id" gorm:"type:uuid;not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	PurchasedAt time.Time `json:"purchased_at"`

	// Relationships
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	PlantType PlantType `json:"plant_type" gorm:"foreignKey:PlantTypeID"`
}

func (si *SeedInventory) BeforeCreate(tx *gorm.DB) error {
	if si.ID == uuid.Nil {
		si.ID = uuid.New()
	}
	return nil
}

// GardenPermission represents the permissions a user has on a shared garden
type GardenPermission string

const (
	GardenPermissionView    GardenPermission = "view"
	GardenPermissionPlant   GardenPermission = "plant"
	GardenPermissionHarvest GardenPermission = "harvest"
	GardenPermissionManage  GardenPermission = "manage"
)

// Value implements the driver.Valuer interface for GardenPermission
func (gp GardenPermission) Value() (interface{}, error) {
	return string(gp), nil
}

// Scan implements the sql.Scanner interface for GardenPermission
func (gp *GardenPermission) Scan(value interface{}) error {
	if value == nil {
		*gp = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*gp = GardenPermission(v)
	case []byte:
		*gp = GardenPermission(string(v))
	default:
		return fmt.Errorf("cannot scan %T into GardenPermission", value)
	}
	return nil
}

// GardenPermissions represents an array of garden permissions
type GardenPermissions []GardenPermission

// Value implements the driver.Valuer interface for GardenPermissions
func (gps GardenPermissions) Value() (interface{}, error) {
	if len(gps) == 0 {
		return []byte("[]"), nil
	}

	values := make([]string, len(gps))
	for i, gp := range gps {
		values[i] = string(gp)
	}

	jsonBytes, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

// Scan implements the sql.Scanner interface for GardenPermissions
func (gps *GardenPermissions) Scan(value interface{}) error {
	if value == nil {
		*gps = nil
		return nil
	}

	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return fmt.Errorf("cannot scan %T into GardenPermissions", value)
	}

	if str == "" || str == "[]" {
		*gps = make(GardenPermissions, 0)
		return nil
	}

	var values []string
	if err := json.Unmarshal([]byte(str), &values); err != nil {
		return err
	}

	permissions := make(GardenPermissions, len(values))
	for i, value := range values {
		permissions[i] = GardenPermission(value)
	}

	*gps = permissions
	return nil
}

// MarshalJSON implements json.Marshaler
func (gps GardenPermissions) MarshalJSON() ([]byte, error) {
	if len(gps) == 0 {
		return []byte("[]"), nil
	}

	values := make([]string, len(gps))
	for i, gp := range gps {
		values[i] = string(gp)
	}

	return json.Marshal(values)
}

// UnmarshalJSON implements json.Unmarshaler
func (gps *GardenPermissions) UnmarshalJSON(data []byte) error {
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	permissions := make(GardenPermissions, len(values))
	for i, value := range values {
		permissions[i] = GardenPermission(value)
	}

	*gps = permissions
	return nil
}

// GormDataType returns the GORM data type
func (GardenPermissions) GormDataType() string {
	return "jsonb"
}

// GardenShare represents a user's access to a shared garden
type GardenShare struct {
	ID          uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GardenID    uuid.UUID         `json:"garden_id" gorm:"type:uuid;not null"`
	UserID      uuid.UUID         `json:"user_id" gorm:"type:uuid;not null"`
	Permissions GardenPermissions `json:"permissions" gorm:"type:jsonb;not null;serializer:json"`
	SharedBy    uuid.UUID         `json:"shared_by" gorm:"type:uuid;not null"`
	SharedAt    time.Time         `json:"shared_at"`
	ExpiresAt   *time.Time        `json:"expires_at"`

	// Relationships
	Garden       Garden `json:"garden" gorm:"foreignKey:GardenID"`
	User         User   `json:"user" gorm:"foreignKey:UserID"`
	SharedByUser User   `json:"shared_by_user" gorm:"foreignKey:SharedBy"`
}

func (gs *GardenShare) BeforeCreate(tx *gorm.DB) error {
	if gs.ID == uuid.Nil {
		gs.ID = uuid.New()
	}
	if gs.SharedAt.IsZero() {
		gs.SharedAt = time.Now()
	}

	// Manually serialize permissions to JSON for jsonb column
	if len(gs.Permissions) > 0 {
		values := make([]string, len(gs.Permissions))
		for i, gp := range gs.Permissions {
			values[i] = string(gp)
		}
		jsonBytes, _ := json.Marshal(values)
		tx.Statement.SetColumn("permissions", jsonBytes)
	} else {
		tx.Statement.SetColumn("permissions", []byte("[]"))
	}

	return nil
}

func (gs *GardenShare) BeforeUpdate(tx *gorm.DB) error {
	// Manually serialize permissions to JSON for jsonb column
	if len(gs.Permissions) > 0 {
		values := make([]string, len(gs.Permissions))
		for i, gp := range gs.Permissions {
			values[i] = string(gp)
		}
		jsonBytes, _ := json.Marshal(values)
		tx.Statement.SetColumn("permissions", jsonBytes)
	} else {
		tx.Statement.SetColumn("permissions", []byte("[]"))
	}

	return nil
}

// GardenAccessLink represents a shareable link for garden access
type GardenAccessLink struct {
	ID          uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	GardenID    uuid.UUID         `json:"garden_id" gorm:"type:uuid;not null"`
	Token       string            `json:"token" gorm:"uniqueIndex;not null"`
	Permissions GardenPermissions `json:"permissions" gorm:"type:jsonb;not null;serializer:json"`
	MaxUses     *int              `json:"max_uses"`
	UsedCount   int               `json:"used_count" gorm:"default:0"`
	CreatedBy   uuid.UUID         `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at"`
	IsActive    bool              `json:"is_active" gorm:"default:true"`

	// Relationships
	Garden        Garden `json:"garden" gorm:"foreignKey:GardenID"`
	CreatedByUser User   `json:"created_by_user" gorm:"foreignKey:CreatedBy"`
}

func (gal *GardenAccessLink) BeforeCreate(tx *gorm.DB) error {
	if gal.ID == uuid.Nil {
		gal.ID = uuid.New()
	}
	if gal.Token == "" {
		gal.Token = generateAccessToken()
	}
	if gal.CreatedAt.IsZero() {
		gal.CreatedAt = time.Now()
	}

	// Manually serialize permissions to JSON for jsonb column
	if len(gal.Permissions) > 0 {
		values := make([]string, len(gal.Permissions))
		for i, gp := range gal.Permissions {
			values[i] = string(gp)
		}
		jsonBytes, _ := json.Marshal(values)
		tx.Statement.SetColumn("permissions", jsonBytes)
	} else {
		tx.Statement.SetColumn("permissions", []byte("[]"))
	}

	return nil
}

func (gal *GardenAccessLink) BeforeUpdate(tx *gorm.DB) error {
	// Manually serialize permissions to JSON for jsonb column
	if len(gal.Permissions) > 0 {
		values := make([]string, len(gal.Permissions))
		for i, gp := range gal.Permissions {
			values[i] = string(gp)
		}
		jsonBytes, _ := json.Marshal(values)
		tx.Statement.SetColumn("permissions", jsonBytes)
	} else {
		tx.Statement.SetColumn("permissions", []byte("[]"))
	}

	return nil
}

// generateAccessToken creates a secure random token for access links
func generateAccessToken() string {
	return uuid.New().String()[:12]
}
