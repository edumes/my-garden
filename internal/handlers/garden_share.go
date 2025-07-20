package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
	"gorm.io/gorm"
)

type GardenShareHandler struct {
	db *database.Database
}

func NewGardenShareHandler(db *database.Database) *GardenShareHandler {
	return &GardenShareHandler{db: db}
}

type CreateAccessLinkRequest struct {
	GardenID    uuid.UUID                `json:"garden_id" binding:"required"`
	Permissions models.GardenPermissions `json:"permissions" binding:"required"`
	MaxUses     *int                     `json:"max_uses"`
	ExpiresAt   *time.Time               `json:"expires_at"`
}

type JoinGardenRequest struct {
	Token string `json:"token" binding:"required"`
}

type UpdateSharePermissionsRequest struct {
	UserID      uuid.UUID                `json:"user_id" binding:"required"`
	Permissions models.GardenPermissions `json:"permissions" binding:"required"`
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

	// Verify garden exists and user owns it
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", req.GardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Validate permissions
	for _, perm := range req.Permissions {
		if perm != models.GardenPermissionView &&
			perm != models.GardenPermissionPlant &&
			perm != models.GardenPermissionHarvest &&
			perm != models.GardenPermissionManage {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission: " + string(perm)})
			return
		}
	}

	accessLink := models.GardenAccessLink{
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
// @Success 200 {object} map[string]interface{} "Successfully joined garden"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Access link not found or expired"
// @Failure 410 {object} map[string]interface{} "Access link expired or max uses reached"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /garden-shares/join [post]
func (h *GardenShareHandler) JoinGarden(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req JoinGardenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find access link
	var accessLink models.GardenAccessLink
	if err := h.db.DB.Where("token = ? AND is_active = ?", req.Token, true).First(&accessLink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Access link not found or inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access link"})
		return
	}

	// Check if link is expired
	if accessLink.ExpiresAt != nil && time.Now().After(*accessLink.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "Access link has expired"})
		return
	}

	// Check if max uses reached
	if accessLink.MaxUses != nil && accessLink.UsedCount >= *accessLink.MaxUses {
		c.JSON(http.StatusGone, gin.H{"error": "Access link has reached maximum uses"})
		return
	}

	// Check if user already has access to this garden
	var existingShare models.GardenShare
	if err := h.db.DB.Where("garden_id = ? AND user_id = ?", accessLink.GardenID, userID).First(&existingShare).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You already have access to this garden"})
		return
	}

	// Create garden share
	gardenShare := models.GardenShare{
		GardenID:    accessLink.GardenID,
		UserID:      userID.(uuid.UUID),
		Permissions: accessLink.Permissions,
		SharedBy:    accessLink.CreatedBy,
	}

	if err := h.db.DB.Create(&gardenShare).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join garden"})
		return
	}

	// Increment used count
	h.db.DB.Model(&accessLink).Update("used_count", accessLink.UsedCount+1)

	// Load garden details
	var garden models.Garden
	h.db.DB.Where("id = ?", accessLink.GardenID).First(&garden)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined garden",
		"garden":  garden,
		"share":   gardenShare,
	})
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

	var shares []models.GardenShare
	if err := h.db.DB.Where("user_id = ?", userID).
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	// Verify garden exists and user owns it
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	var shares []models.GardenShare
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
	userID, exists := c.Get("user_id")
	if !exists {
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

	// Verify garden exists and user owns it
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Validate permissions
	for _, perm := range req.Permissions {
		if perm != models.GardenPermissionView &&
			perm != models.GardenPermissionPlant &&
			perm != models.GardenPermissionHarvest &&
			perm != models.GardenPermissionManage {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission: " + string(perm)})
			return
		}
	}

	// Find and update the share
	var share models.GardenShare
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
	userID, exists := c.Get("user_id")
	if !exists {
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

	// Verify garden exists and user owns it
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Find and delete the share
	var share models.GardenShare
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	// Verify garden exists and user owns it
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	var accessLinks []models.GardenAccessLink
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
	userID, exists := c.Get("user_id")
	if !exists {
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

	// Verify garden exists and user owns it
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found or access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Find and deactivate the access link
	var accessLink models.GardenAccessLink
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
