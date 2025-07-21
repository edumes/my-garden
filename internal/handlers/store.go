package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
	"github.com/my-garden/api/internal/services"
	"gorm.io/gorm"
)

type StoreHandler struct {
	db           *database.Database
	auditService *services.AuditService
}

func NewStoreHandler(db *database.Database, auditService *services.AuditService) *StoreHandler {
	return &StoreHandler{db: db, auditService: auditService}
}

type BuySeedRequest struct {
	PlantTypeID uuid.UUID `json:"plant_type_id" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required,min=1,max=10"`
}

type BuySeedResponse struct {
	PlantType models.PlantType `json:"plant_type"`
	Quantity  int              `json:"quantity"`
	TotalCost int              `json:"total_cost"`
	UserCoins int              `json:"user_coins"`
}

// GetStoreInventory godoc
// @Summary Get store inventory
// @Description Get all available plant types with their prices
// @Tags store
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Store inventory"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /store/inventory [get]
func (h *StoreHandler) GetStoreInventory(c *gin.Context) {
	var plantTypes []models.PlantType
	if err := h.db.DB.Find(&plantTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch store inventory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"inventory": plantTypes,
	})
}

// BuySeed godoc
// @Summary Buy seeds
// @Description Purchase seeds for planting
// @Tags store
// @Accept json
// @Produce json
// @Param request body BuySeedRequest true "Buy seed request"
// @Success 200 {object} BuySeedResponse "Purchase successful"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Plant type not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /store/buy [post]
func (h *StoreHandler) BuySeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req BuySeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate quantity
	if req.Quantity < 1 || req.Quantity > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be between 1 and 10"})
		return
	}

	// Get plant type
	var plantType models.PlantType
	if err := h.db.DB.First(&plantType, req.PlantTypeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant type"})
		return
	}

	// Calculate total cost
	totalCost := plantType.SeedPrice * req.Quantity

	// Get user and check if they have enough coins
	var user models.User
	if err := h.db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	if user.Coins < totalCost {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Insufficient coins",
			"required":  totalCost,
			"available": user.Coins,
		})
		return
	}

	// Start transaction
	tx := h.db.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Deduct coins from user
	if err := tx.Model(&user).Update("coins", user.Coins-totalCost).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user coins"})
		return
	}

	// Create seed inventory record (optional - for tracking purposes)
	seedInventory := models.SeedInventory{
		UserID:      userID.(uuid.UUID),
		PlantTypeID: req.PlantTypeID,
		Quantity:    req.Quantity,
		PurchasedAt: time.Now(),
	}

	if err := tx.Create(&seedInventory).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record purchase"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		h.auditService.LogFailure(c, models.AuditActionSeedPurchase, models.AuditResourceStore, nil, gin.H{"error": "Failed to commit transaction"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Get updated user coins
	var updatedUser models.User
	h.db.DB.First(&updatedUser, userID)

	response := BuySeedResponse{
		PlantType: plantType,
		Quantity:  req.Quantity,
		TotalCost: totalCost,
		UserCoins: updatedUser.Coins,
	}

	h.auditService.LogSuccess(c, models.AuditActionSeedPurchase, models.AuditResourceStore, &req.PlantTypeID, gin.H{
		"plant_type_id": req.PlantTypeID,
		"quantity":      req.Quantity,
		"total_cost":    totalCost,
		"user_coins":    updatedUser.Coins,
	})

	c.JSON(http.StatusOK, response)
}

// GetUserSeedInventory godoc
// @Summary Get user's seed inventory
// @Description Get the user's purchased seeds
// @Tags store
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "User seed inventory"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /store/inventory/user [get]
func (h *StoreHandler) GetUserSeedInventory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var seedInventory []models.SeedInventory
	if err := h.db.DB.Preload("PlantType").
		Where("user_id = ?", userID).
		Order("purchased_at DESC").
		Find(&seedInventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user seed inventory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seed_inventory": seedInventory,
	})
}
