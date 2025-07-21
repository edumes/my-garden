package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	Action     string     `json:"action" gorm:"not null"`
	Resource   string     `json:"resource" gorm:"not null"`
	ResourceID *uuid.UUID `json:"resource_id" gorm:"type:uuid"`
	IPAddress  string     `json:"ip_address"`
	Status     string     `json:"status" gorm:"not null"`
	CreatedAt  time.Time  `json:"created_at"`

	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type AuditAction string

const (
	// Authentication actions
	AuditActionUserRegister  AuditAction = "user_register"
	AuditActionUserLogin     AuditAction = "user_login"
	AuditActionUserLogout    AuditAction = "user_logout"
	AuditActionTokenRefresh  AuditAction = "token_refresh"
	AuditActionProfileUpdate AuditAction = "profile_update"

	// Garden actions
	AuditActionGardenCreate AuditAction = "garden_create"
	AuditActionGardenUpdate AuditAction = "garden_update"
	AuditActionGardenDelete AuditAction = "garden_delete"
	AuditActionGardenView   AuditAction = "garden_view"

	// Plant actions
	AuditActionPlantSeed      AuditAction = "plant_seed"
	AuditActionPlantHarvest   AuditAction = "plant_harvest"
	AuditActionPlantRemove    AuditAction = "plant_remove"
	AuditActionPlantWater     AuditAction = "plant_water"
	AuditActionPlantFertilize AuditAction = "plant_fertilize"

	// Garden sharing actions
	AuditActionGardenShare          AuditAction = "garden_share"
	AuditActionGardenJoin           AuditAction = "garden_join"
	AuditActionSharePermission      AuditAction = "share_permission_update"
	AuditActionShareRemove          AuditAction = "share_remove"
	AuditActionAccessLinkCreate     AuditAction = "access_link_create"
	AuditActionAccessLinkDeactivate AuditAction = "access_link_deactivate"

	// Store actions
	AuditActionSeedPurchase  AuditAction = "seed_purchase"
	AuditActionInventoryView AuditAction = "inventory_view"

	// System actions
	AuditActionError    AuditAction = "error"
	AuditActionSecurity AuditAction = "security"
)

type AuditResource string

const (
	AuditResourceUser   AuditResource = "user"
	AuditResourceGarden AuditResource = "garden"
	AuditResourcePlant  AuditResource = "plant"
	AuditResourceShare  AuditResource = "garden_share"
	AuditResourceStore  AuditResource = "store"
	AuditResourceSystem AuditResource = "system"
)

type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailure AuditStatus = "failure"
	AuditStatusError   AuditStatus = "error"
)
