package garden

import (
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GardenAccess is what one user may do in one garden: everything as its
// owner, or the capabilities of an unexpired share.
type GardenAccess struct {
	Owner       bool
	Permissions GardenPermissions
}

// ownerOnly marks actions no share can grant, such as deleting the garden or
// managing who it is shared with. It is a var, not a const, so swag does not
// list it among the GardenPermission values in the API documentation.
var ownerOnly GardenPermission = "owner"

var allPermissions = GardenPermissions{
	GardenPermissionView,
	GardenPermissionPlant,
	GardenPermissionHarvest,
	GardenPermissionManage,
}

// Can reports whether the access allows want. Shares hold independent
// capabilities: manage grants all of them, and any share grants view.
func (a GardenAccess) Can(want GardenPermission) bool {
	if want == ownerOnly || a.Owner {
		return a.Owner
	}
	return want == GardenPermissionView ||
		slices.Contains(a.Permissions, want) ||
		slices.Contains(a.Permissions, GardenPermissionManage)
}

// Capabilities lists every permission the access allows.
func (a GardenAccess) Capabilities() GardenPermissions {
	return slices.DeleteFunc(slices.Clone(allPermissions), func(p GardenPermission) bool {
		return !a.Can(p)
	})
}

var errNoGardenAccess = errors.New("garden not found or not shared with the user")

// unexpiredShares limits a GardenShare query to shares that are still valid.
func unexpiredShares(db *gorm.DB) *gorm.DB {
	return db.Where("(garden_shares.expires_at IS NULL OR garden_shares.expires_at > ?)", time.Now())
}

// loadGardenAccess returns the user's access to the garden, or
// errNoGardenAccess when the garden does not exist or the user has no
// unexpired share of it.
func loadGardenAccess(db *gorm.DB, userID, gardenID uuid.UUID) (GardenAccess, error) {
	var garden Garden
	err := db.Select("user_id").Where("id = ?", gardenID).Take(&garden).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GardenAccess{}, errNoGardenAccess
	}
	if err != nil {
		return GardenAccess{}, err
	}
	if garden.UserID == userID {
		return GardenAccess{Owner: true}, nil
	}

	var share GardenShare
	err = db.Scopes(unexpiredShares).
		Where("garden_id = ? AND user_id = ?", gardenID, userID).
		Take(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GardenAccess{}, errNoGardenAccess
	}
	if err != nil {
		return GardenAccess{}, err
	}
	return GardenAccess{Permissions: share.Permissions}, nil
}

// authorizeGarden checks that the authenticated user may do want in the
// garden. When they may not, it writes the response and returns false: 404
// without any access, so gardens stay private, and 403 with access that
// lacks the capability.
func authorizeGarden(c *gin.Context, db *gorm.DB, gardenID uuid.UUID, want GardenPermission) (GardenAccess, bool) {
	access, err := loadGardenAccess(db, c.MustGet("user_id").(uuid.UUID), gardenID)
	switch {
	case errors.Is(err, errNoGardenAccess):
		c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
		return access, false
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check garden access"})
		return access, false
	case !access.Can(want):
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to do this in this garden"})
		return access, false
	}
	return access, true
}
