package audit

import (
	"time"

	"github.com/google/uuid"
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
}

// AuditAction represents the type of action being audited
type AuditAction string

const (
	// User actions
	AuditActionUserRegister         AuditAction = "user_register"
	AuditActionUserLogin            AuditAction = "user_login"
	AuditActionUserLogout           AuditAction = "user_logout"
	AuditActionTokenRefresh         AuditAction = "token_refresh"
	AuditActionProfileUpdate        AuditAction = "profile_update"
	AuditActionPlantWater           AuditAction = "plant_water"
	AuditActionPlantFertilize       AuditAction = "plant_fertilize"
	AuditActionGardenShare          AuditAction = "garden_share"
	AuditActionInventoryView        AuditAction = "inventory_view"
	AuditActionError                AuditAction = "error"
	AuditActionSecurity             AuditAction = "security"
	AuditActionAccessLinkDeactivate AuditAction = "access_link_deactivate"

	// Garden actions
	AuditActionCreateGarden     AuditAction = "create_garden"
	AuditActionUpdateGarden     AuditAction = "update_garden"
	AuditActionDeleteGarden     AuditAction = "delete_garden"
	AuditActionPlantSeed        AuditAction = "plant_seed"
	AuditActionPlantHarvest     AuditAction = "plant_harvest"
	AuditActionPlantRemove      AuditAction = "plant_remove"
	AuditActionShareGarden      AuditAction = "share_garden"
	AuditActionUnshareGarden    AuditAction = "unshare_garden"
	AuditActionSeedPurchase     AuditAction = "seed_purchase"
	AuditActionGardenCreate     AuditAction = "garden_create"
	AuditActionGardenUpdate     AuditAction = "garden_update"
	AuditActionGardenView       AuditAction = "garden_view"
	AuditActionGardenDelete     AuditAction = "garden_delete"
	AuditActionGardenJoin       AuditAction = "garden_join"
	AuditActionAccessLinkCreate AuditAction = "access_link_create"
	AuditActionSharePermission  AuditAction = "share_permission"
	AuditActionShareRemove      AuditAction = "share_remove"
)

// AuditResource represents the type of resource being audited
type AuditResource string

const (
	AuditResourceUser   AuditResource = "user"
	AuditResourceGarden AuditResource = "garden"
	AuditResourcePlant  AuditResource = "plant"
	AuditResourceShare  AuditResource = "share"
	AuditResourceStore  AuditResource = "store"
	AuditResourceSystem AuditResource = "system"
)

type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailure AuditStatus = "failure"
	AuditStatusError   AuditStatus = "error"
)

type AuditFilters struct {
	UserID     *uuid.UUID
	Action     string
	Resource   string
	ResourceID *uuid.UUID
	Status     string
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	Offset     int
}
