package store

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/garden"
)

type StoreHandler struct {
	service      *Service
	auditService *audit.AuditService
}

func NewStoreHandler(service *Service, auditService *audit.AuditService) *StoreHandler {
	return &StoreHandler{
		service:      service,
		auditService: auditService,
	}
}

type BuySeedRequest struct {
	PlantTypeID uuid.UUID `json:"plant_type_id" binding:"required"`
	Quantity    int       `json:"quantity" binding:"required,min=1,max=10"`
}

type BuySeedResponse struct {
	PlantType garden.PlantType `json:"plant_type"`
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
	plantTypes, err := h.service.GetStoreInventory()
	if err != nil {
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

	response, err := h.service.BuySeed(userID.(uuid.UUID), req.PlantTypeID, req.Quantity)
	if err != nil {
		h.auditService.LogFailure(c, audit.AuditActionSeedPurchase, audit.AuditResourceStore, &req.PlantTypeID, gin.H{"error": err.Error()})

		switch err.Error() {
		case "quantity must be between 1 and 10":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case "record not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant type not found"})
		default:
			if contains(err.Error(), "insufficient coins") {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process purchase"})
			}
		}
		return
	}

	h.auditService.LogSuccess(c, audit.AuditActionSeedPurchase, audit.AuditResourceStore, &req.PlantTypeID, gin.H{
		"plant_type_id": req.PlantTypeID,
		"quantity":      req.Quantity,
		"total_cost":    response.TotalCost,
		"user_coins":    response.UserCoins,
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

	seedInventory, err := h.service.GetUserSeedInventory(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user seed inventory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"seed_inventory": seedInventory,
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
