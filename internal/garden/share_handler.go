package garden

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GardenShareHandler struct {
	db *database.Database
}

func NewGardenShareHandler(db *database.Database) *GardenShareHandler {
	return &GardenShareHandler{db: db}
}

type CreateAccessLinkRequest struct {
	GardenID    uuid.UUID         `json:"garden_id" binding:"required"`
	Permissions GardenPermissions `json:"permissions" binding:"required"`
	MaxUses     *int              `json:"max_uses"`
	ExpiresAt   *time.Time        `json:"expires_at"`
}

type JoinGardenRequest struct {
	Token string `json:"token" binding:"required"`
}

type UpdateSharePermissionsRequest struct {
	UserID      uuid.UUID         `json:"user_id" binding:"required"`
	Permissions GardenPermissions `json:"permissions" binding:"required"`
}

// CreateAccessLink godoc
// @Summary Create garden access link
// @Description Create a shareable link for garden access with specified permissions
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param request body CreateAccessLinkRequest true "Access link creation data"
// @Success 201 {object} map[string]interface{} "Created access link"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not garden owner"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/access-links [post]
func (h *GardenShareHandler) CreateAccessLink(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateAccessLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, req.GardenID, ownerOnly); !ok {
		return
	}

	// Validate permissions
	for _, perm := range req.Permissions {
		if perm != GardenPermissionView &&
			perm != GardenPermissionPlant &&
			perm != GardenPermissionHarvest &&
			perm != GardenPermissionManage {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission: " + string(perm)})
			return
		}
	}

	accessLink := GardenAccessLink{
		GardenID:    req.GardenID,
		Permissions: req.Permissions,
		MaxUses:     req.MaxUses,
		ExpiresAt:   req.ExpiresAt,
		CreatedBy:   userID.(uuid.UUID),
		IsActive:    true,
	}

	if err := h.db.DB.Create(&accessLink).Error; err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access link"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"access_link": accessLink,
		"share_url":   "/join-garden/" + accessLink.Token,
	})
}

// JoinGarden godoc
// @Summary Join garden via access link
// @Description Join a garden using an access link token
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param request body JoinGardenRequest true "Join garden data"
// @Success 200 {object} map[string]interface{} "Joined the garden, or already a member"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Access link not found or expired"
// @Failure 410 {object} map[string]interface{} "Access link expired or max uses reached"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Failure 409 {object} map[string]interface{} "The caller owns the garden"
// @Router /garden-shares/join [post]
func (h *GardenShareHandler) JoinGarden(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid := userID.(uuid.UUID)

	var req JoinGardenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var link GardenAccessLink
	if err := h.db.DB.Where("token = ?", req.Token).Take(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found or inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access link"})
		return
	}

	var garden Garden
	if err := h.db.DB.Where("id = ?", link.GardenID).Take(&garden).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}
	if garden.UserID == uid {
		c.JSON(http.StatusConflict, gin.H{"error": "You own this garden"})
		return
	}

	share, err := h.redeem(link, uid)
	if errors.Is(err, errLinkUnavailable) || errors.Is(err, errAlreadyHasAccess) {
		// Joining is idempotent: members, including those whose concurrent
		// request won, get their current access back without spending a use.
		var existing GardenShare
		if findErr := h.db.DB.Scopes(unexpiredShares).
			Where("garden_id = ? AND user_id = ?", link.GardenID, uid).
			Take(&existing).Error; findErr == nil {
			share, err = existing, nil
		}
	}

	switch {
	case errors.Is(err, errLinkUnavailable):
		respondLinkUnavailable(c, link)
	case errors.Is(err, errAlreadyHasAccess):
		c.JSON(http.StatusConflict, gin.H{"error": "This garden was already shared with you"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join garden"})
	default:
		c.JSON(http.StatusOK, gin.H{
			"message": "Successfully joined garden",
			"garden":  garden,
			"share":   share,
		})
	}
}

var (
	errLinkUnavailable  = errors.New("access link inactive, expired, or used up")
	errAlreadyHasAccess = errors.New("garden already shared with the user")
)

// redeem spends one use of the link and shares its garden with the user, as
// one transaction. The use is claimed with a single conditional update, so
// concurrent requests can never exceed max_uses or redeem an expired or
// deactivated link; a rejected share gives the use back.
func (h *GardenShareHandler) redeem(link GardenAccessLink, userID uuid.UUID) (GardenShare, error) {
	share := GardenShare{
		GardenID:    link.GardenID,
		UserID:      userID,
		Permissions: link.Permissions,
		SharedBy:    link.CreatedBy,
	}
	err := h.db.DB.Transaction(func(tx *gorm.DB) error {
		claim := tx.Exec(`UPDATE garden_access_links SET used_count = used_count + 1
			WHERE id = ? AND is_active
			AND (expires_at IS NULL OR expires_at > ?)
			AND (max_uses IS NULL OR used_count < max_uses)`, link.ID, time.Now())
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected == 0 {
			return errLinkUnavailable
		}

		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&share)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			return errAlreadyHasAccess
		}
		return nil
	})
	return share, err
}

// respondLinkUnavailable explains why a link could not be redeemed.
func respondLinkUnavailable(c *gin.Context, link GardenAccessLink) {
	switch {
	case !link.IsActive:
		c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found or inactive"})
	case link.ExpiresAt != nil && !time.Now().Before(*link.ExpiresAt):
		c.JSON(http.StatusGone, gin.H{"error": "Access link has expired"})
	default:
		c.JSON(http.StatusGone, gin.H{"error": "Access link has reached maximum uses"})
	}
}

// GetSharedGardens godoc
// @Summary Get shared gardens
// @Description Get all gardens shared with the current user
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Success 200 {object} map[string]interface{} "List of shared gardens"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/shared-with-me [get]
func (h *GardenShareHandler) GetSharedGardens(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var shares []GardenShare
	if err := h.db.DB.Scopes(unexpiredShares).Where("user_id = ?", userID).
		Preload("Garden.User").
		Preload("Garden.Plants.PlantType").
		Preload("SharedByUser").
		Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared gardens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shared_gardens": shares})
}

// GetGardenShares godoc
// @Summary Get garden shares
// @Description Get all users who have access to a specific garden
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param garden_id path string true "Garden ID"
// @Success 200 {object} map[string]interface{} "List of garden shares"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not garden owner"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/garden/{garden_id} [get]
func (h *GardenShareHandler) GetGardenShares(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, ownerOnly); !ok {
		return
	}

	var shares []GardenShare
	if err := h.db.DB.Where("garden_id = ?", gardenID).
		Preload("User").
		Preload("SharedByUser").
		Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden shares"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"garden_shares": shares})
}

// UpdateSharePermissions godoc
// @Summary Update share permissions
// @Description Update permissions for a user's access to a garden
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param garden_id path string true "Garden ID"
// @Param request body UpdateSharePermissionsRequest true "Permission update data"
// @Success 200 {object} map[string]interface{} "Updated garden share"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not garden owner"
// @Failure 404 {object} map[string]interface{} "Garden share not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/garden/{garden_id}/permissions [put]
func (h *GardenShareHandler) UpdateSharePermissions(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	var req UpdateSharePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, ownerOnly); !ok {
		return
	}

	// Validate permissions
	for _, perm := range req.Permissions {
		if perm != GardenPermissionView &&
			perm != GardenPermissionPlant &&
			perm != GardenPermissionHarvest &&
			perm != GardenPermissionManage {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission: " + string(perm)})
			return
		}
	}

	// Find and update the share
	var share GardenShare
	if err := h.db.DB.Where("garden_id = ? AND user_id = ?", gardenID, req.UserID).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden share not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden share"})
		return
	}

	share.Permissions = req.Permissions
	if err := h.db.DB.Save(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"garden_share": share})
}

// RemoveGardenShare godoc
// @Summary Remove garden share
// @Description Remove a user's access to a garden
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param garden_id path string true "Garden ID"
// @Param user_id path string true "User ID to remove"
// @Success 200 {object} map[string]interface{} "Garden share removed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not garden owner"
// @Failure 404 {object} map[string]interface{} "Garden share not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/garden/{garden_id}/user/{user_id} [delete]
func (h *GardenShareHandler) RemoveGardenShare(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	shareUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, ownerOnly); !ok {
		return
	}

	// Find and delete the share
	var share GardenShare
	if err := h.db.DB.Where("garden_id = ? AND user_id = ?", gardenID, shareUserID).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden share not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden share"})
		return
	}

	if err := h.db.DB.Delete(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove garden share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Garden share removed successfully"})
}

// GetAccessLinks godoc
// @Summary Get access links
// @Description Get all access links for a garden
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param garden_id path string true "Garden ID"
// @Success 200 {object} map[string]interface{} "List of access links"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not garden owner"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/garden/{garden_id}/access-links [get]
func (h *GardenShareHandler) GetAccessLinks(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, ownerOnly); !ok {
		return
	}

	var accessLinks []GardenAccessLink
	if err := h.db.DB.Where("garden_id = ?", gardenID).
		Preload("CreatedByUser").
		Find(&accessLinks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_links": accessLinks})
}

// DeactivateAccessLink godoc
// @Summary Deactivate access link
// @Description Deactivate an access link for a garden
// @Tags garden-sharing
// @Accept json
// @Produce json
// @Security bearer
// @Param garden_id path string true "Garden ID"
// @Param link_id path string true "Access Link ID"
// @Success 200 {object} map[string]interface{} "Access link deactivated successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not garden owner"
// @Failure 404 {object} map[string]interface{} "Access link not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/garden/{garden_id}/access-links/{link_id}/deactivate [post]
func (h *GardenShareHandler) DeactivateAccessLink(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	linkID, err := uuid.Parse(c.Param("link_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid link ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, ownerOnly); !ok {
		return
	}

	// Find and deactivate the access link
	var accessLink GardenAccessLink
	if err := h.db.DB.Where("id = ? AND garden_id = ?", linkID, gardenID).First(&accessLink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access link"})
		return
	}

	accessLink.IsActive = false
	if err := h.db.DB.Save(&accessLink).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate access link"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Access link deactivated successfully"})
}
