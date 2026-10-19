package garden

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GardenHandler struct {
	db           *database.Database
	auditService *audit.AuditService
}

func NewGardenHandler(db *database.Database, auditService *audit.AuditService) *GardenHandler {
	return &GardenHandler{db: db, auditService: auditService}
}

type CreateGardenRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100" example:"My First Garden"`
	Description string `json:"description" example:"A beautiful garden for growing vegetables"`
}

type UpdateGardenRequest struct {
	Name        string `json:"name" example:"Updated Garden Name"`
	Description string `json:"description" example:"Updated garden description"`
}

type PlantRequest struct {
	PlantTypeID uuid.UUID `json:"plant_type_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
	Position    *int      `json:"position" binding:"required,min=0,max=8" example:"0"`
}

// GetGardens godoc
// @Summary Get user gardens
// @Description Get all gardens for the current user
// @Tags gardens
// @Accept json
// @Produce json
// @Security bearer
// @Success 200 {object} map[string]interface{} "List of user gardens"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens [get]
func (h *GardenHandler) GetGardens(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get owned gardens
	var ownedGardens []Garden
	if err := h.db.DB.Where("user_id = ?", userID).Preload("Plants.PlantType").Find(&ownedGardens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch owned gardens"})
		return
	}

	// Get shared gardens
	var sharedGardens []GardenShare
	if err := h.db.DB.Scopes(unexpiredShares).Where("user_id = ?", userID).
		Preload("Garden.Plants.PlantType").
		Preload("Garden.User").
		Preload("SharedByUser").
		Find(&sharedGardens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch shared gardens"})
		return
	}

	// Transform shared gardens to include permission info
	var sharedGardenResponses []gin.H
	for _, share := range sharedGardens {
		sharedGardenResponses = append(sharedGardenResponses, gin.H{
			"garden":     share.Garden,
			"permission": share.Permissions,
			"shared_by":  share.SharedByUser,
			"shared_at":  share.SharedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"owned_gardens":  ownedGardens,
		"shared_gardens": sharedGardenResponses,
	})
}

// CreateGarden godoc
// @Summary Create a new garden
// @Description Create a new garden for the current user. Players own one garden at levels 1 and 2 and one more every two levels
// @Tags gardens
// @Accept json
// @Produce json
// @Security bearer
// @Param request body CreateGardenRequest true "Garden creation data"
// @Success 201 {object} map[string]interface{} "Created garden"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Garden limit for the player's level reached"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens [post]
func (h *GardenHandler) CreateGarden(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateGardenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	garden := Garden{
		UserID:      userID.(uuid.UUID),
		Name:        req.Name,
		Description: req.Description,
	}

	err := h.db.DB.Transaction(func(tx *gorm.DB) error {
		// Locking the owner's row serializes their garden creations, so
		// concurrent requests cannot exceed the limit.
		var level int
		if err := tx.Raw("SELECT level FROM users WHERE id = ? FOR UPDATE", garden.UserID).Scan(&level).Error; err != nil {
			return err
		}
		var owned int64
		if err := tx.Model(&Garden{}).Where("user_id = ?", garden.UserID).Count(&owned).Error; err != nil {
			return err
		}
		if int(owned) >= MaxGardens(level) {
			return gardenLimitError{owned: int(owned), nextLevel: 2*int(owned) + 1}
		}
		return tx.Create(&garden).Error
	})
	var limit gardenLimitError
	switch {
	case errors.As(err, &limit):
		c.JSON(http.StatusForbidden, gin.H{"error": limit.Error()})
		return
	case err != nil:
		h.auditService.LogFailure(c, audit.AuditActionGardenCreate, audit.AuditResourceGarden, &garden.ID, gin.H{"error": "Failed to create garden"})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create garden"})
		return
	}

	h.auditService.LogSuccess(c, audit.AuditActionGardenCreate, audit.AuditResourceGarden, &garden.ID, gin.H{
		"garden_name": garden.Name,
		"garden_id":   garden.ID,
	})

	c.JSON(http.StatusCreated, gin.H{"garden": garden})
}

// GetGarden godoc
// @Summary Get garden details
// @Description Get a specific garden with all its plants
// @Tags gardens
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} map[string]interface{} "Garden with plants, whether the caller owns it, and their permissions"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid garden ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens/{id} [get]
func (h *GardenHandler) GetGarden(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	access, ok := authorizeGarden(c, h.db.DB, gardenID, GardenPermissionView)
	if !ok {
		return
	}

	var garden Garden
	if err := h.db.DB.Where("id = ?", gardenID).Preload("Plants.PlantType").First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"garden":      garden,
		"is_owner":    access.Owner,
		"permissions": access.Capabilities(),
	})
}

// UpdateGarden godoc
// @Summary Update garden
// @Description Update garden information
// @Tags gardens
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param request body UpdateGardenRequest true "Garden update data"
// @Success 200 {object} map[string]interface{} "Updated garden"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Failure 403 {object} map[string]interface{} "Forbidden - Missing garden permission"
// @Router /gardens/{id} [put]
func (h *GardenHandler) UpdateGarden(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	var req UpdateGardenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, GardenPermissionManage); !ok {
		return
	}

	var garden Garden
	if err := h.db.DB.Where("id = ?", gardenID).First(&garden).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Update fields
	if req.Name != "" {
		garden.Name = req.Name
	}
	if req.Description != "" {
		garden.Description = req.Description
	}

	if err := h.db.DB.Save(&garden).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update garden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"garden": garden})
}

// DeleteGarden godoc
// @Summary Delete garden
// @Description Delete a garden and all its plants
// @Tags gardens
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Success 200 {object} map[string]interface{} "Garden deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid garden ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Failure 403 {object} map[string]interface{} "Forbidden - Missing garden permission"
// @Router /gardens/{id} [delete]
func (h *GardenHandler) DeleteGarden(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, ownerOnly); !ok {
		return
	}

	// The schema has no cascading deletes, so remove dependents first.
	err = h.db.DB.Transaction(func(tx *gorm.DB) error {
		if err := deleteListingsOf(tx, "SELECT id FROM plants WHERE garden_id = ?", gardenID); err != nil {
			return err
		}
		for _, dependent := range []any{&Plant{}, &GardenShare{}, &GardenAccessLink{}} {
			if err := tx.Where("garden_id = ?", gardenID).Delete(dependent).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&Garden{ID: gardenID}).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete garden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Garden deleted successfully"})
}

// PlantSeed godoc
// @Summary Plant a seed
// @Description Plant one of your seeds in a garden plot and earn experience
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param request body PlantRequest true "Plant data"
// @Success 201 {object} map[string]interface{} "Planted seed and the planting reward"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Garden or plant type not found"
// @Failure 409 {object} map[string]interface{} "Position already occupied"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Failure 403 {object} map[string]interface{} "Forbidden - Missing garden permission"
// @Router /gardens/{id}/plants [post]
func (h *GardenHandler) PlantSeed(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	var req PlantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, GardenPermissionPlant); !ok {
		return
	}

	var plantType PlantType
	if err := h.db.DB.First(&plantType, req.PlantTypeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant type"})
		return
	}

	now := time.Now()
	plant := Plant{
		GardenID:    gardenID,
		PlantTypeID: req.PlantTypeID,
		GrowerID:    userID.(uuid.UUID),
		Position:    *req.Position,
		PlantedAt:   now,
		GrownAt:     now,
	}
	reward := Reward{UserID: plant.GrowerID, Experience: auth.PlantingExperience}
	err = h.db.DB.Transaction(func(tx *gorm.DB) error {
		// Spend a seed only while the caller still has one, so concurrent
		// requests cannot plant more seeds than they own.
		spend := tx.Model(&SeedInventory{}).
			Where("user_id = ? AND plant_type_id = ? AND quantity > 0", userID, req.PlantTypeID).
			Update("quantity", gorm.Expr("quantity - 1"))
		if spend.Error != nil {
			return spend.Error
		}
		if spend.RowsAffected == 0 {
			return errNoSeeds
		}

		// The unique (garden, position) index settles races for one cell;
		// the losing request rolls its spent seed back.
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&plant)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			return errPositionTaken
		}

		var err error
		reward.Level, reward.LevelUp, err = auth.GrantExperience(tx, reward.UserID, reward.Experience)
		return err
	})
	switch {
	case errors.Is(err, errNoSeeds):
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "You don't have any seeds for this plant type. Please purchase seeds from the store first.",
			"plant_type_name": plantType.Name,
		})
		return
	case errors.Is(err, errPositionTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Position already occupied"})
		return
	case err != nil:
		h.auditService.LogFailure(c, audit.AuditActionPlantSeed, audit.AuditResourcePlant, &plant.ID, gin.H{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to plant seed"})
		return
	}

	// Load plant type for response
	h.db.DB.Preload("PlantType").First(&plant, plant.ID)

	h.auditService.LogSuccess(c, audit.AuditActionPlantSeed, audit.AuditResourcePlant, &plant.ID, gin.H{
		"garden_id":     gardenID,
		"plant_type_id": req.PlantTypeID,
		"position":      *req.Position,
		"plant_id":      plant.ID,
	})

	c.JSON(http.StatusCreated, gin.H{"plant": plant, "reward": reward})
}

// HarvestPlant godoc
// @Summary Harvest a plant
// @Description Harvest a ripe plant, freeing its plot. Its grower, who planted or bought it, earns its coins and experience
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param plantId path string true "Plant ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Success 200 {object} map[string]interface{} "Reward paid to the plant's grower"
// @Failure 400 {object} map[string]interface{} "Bad Request - Plant not ready for harvest"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Plant not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Failure 403 {object} map[string]interface{} "Forbidden - Missing garden permission"
// @Router /gardens/{id}/plants/{plantId}/harvest [post]
func (h *GardenHandler) HarvestPlant(c *gin.Context) {
	gardenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	plantID, err := uuid.Parse(c.Param("plantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, GardenPermissionHarvest); !ok {
		return
	}

	var plant Plant
	if err := h.db.DB.Preload("PlantType").
		Where("id = ? AND garden_id = ?", plantID, gardenID).
		First(&plant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant"})
		return
	}

	// Check if plant is harvestable
	if plant.Stage != PlantStageHarvestable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plant is not ready for harvest"})
		return
	}

	// The plant's grower earns the harvest, whoever harvests it.
	reward := Reward{
		Coins:      plant.PlantType.HarvestValue * plant.PlantType.Yield,
		Experience: plant.PlantType.GrowthTime,
	}
	err = h.db.DB.Transaction(func(tx *gorm.DB) error {
		if err := deleteListingsOf(tx, "SELECT id FROM plants WHERE id = ?", plantID); err != nil {
			return err
		}
		// Deleting the plant claims it, so concurrent harvests pay once.
		var harvested Plant
		if err := tx.Raw(`DELETE FROM plants WHERE id = ? AND garden_id = ? AND stage = ? RETURNING *`,
			plantID, gardenID, PlantStageHarvestable).Scan(&harvested).Error; err != nil {
			return err
		}
		if harvested.ID == uuid.Nil {
			return errNotHarvestable
		}

		reward.UserID = harvested.GrowerID
		if err := tx.Model(&auth.User{}).Where("id = ?", reward.UserID).
			Update("coins", gorm.Expr("coins + ?", reward.Coins)).Error; err != nil {
			return err
		}
		var err error
		reward.Level, reward.LevelUp, err = auth.GrantExperience(tx, reward.UserID, reward.Experience)
		return err
	})
	switch {
	case errors.Is(err, errNotHarvestable):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plant is not ready for harvest"})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to harvest plant"})
		return
	}

	h.auditService.LogSuccess(c, audit.AuditActionPlantHarvest, audit.AuditResourcePlant, &plant.ID, gin.H{
		"garden_id":    gardenID,
		"plant_id":     plantID,
		"rewarded":     reward.UserID,
		"coins_earned": reward.Coins,
		"experience":   reward.Experience,
		"level_up":     reward.LevelUp,
		"new_level":    reward.Level,
	})

	c.JSON(http.StatusOK, gin.H{"reward": reward})
}

// RemovePlant godoc
// @Summary Remove a plant
// @Description Remove a plant from the garden
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param plantId path string true "Plant ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Success 200 {object} map[string]interface{} "Plant removed successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Plant not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Failure 403 {object} map[string]interface{} "Forbidden - Missing garden permission"
// @Router /gardens/{id}/plants/{plantId} [delete]
func (h *GardenHandler) RemovePlant(c *gin.Context) {
	if _, exists := c.Get("user_id"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	gardenID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	plantID, err := uuid.Parse(c.Param("plantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant ID"})
		return
	}

	if _, ok := authorizeGarden(c, h.db.DB, gardenID, GardenPermissionManage); !ok {
		return
	}

	var plant Plant
	if err := h.db.DB.Where("id = ? AND garden_id = ?", plantID, gardenID).
		First(&plant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant"})
		return
	}

	err = h.db.DB.Transaction(func(tx *gorm.DB) error {
		if err := deleteListingsOf(tx, "SELECT id FROM plants WHERE id = ?", plant.ID); err != nil {
			return err
		}
		return tx.Delete(&plant).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove plant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plant removed successfully"})
}

// ListPlantTypes godoc
// @Summary List all plant types
// @Description Get a list of all available plant types
// @Tags plants
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of plant types"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /plants [get]
func (h *GardenHandler) ListPlantTypes(c *gin.Context) {
	var plantTypes []PlantType
	if err := h.db.DB.Order("unlock_level, seed_price, name").Find(&plantTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant types"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plant_types": plantTypes})
}

// deleteListingsOf removes marketplace listings, active or sold, for the plants
// the subquery selects, so the plants themselves can be deleted.
func deleteListingsOf(tx *gorm.DB, plantsQuery string, args ...any) error {
	return tx.Exec("DELETE FROM plant_listings WHERE plant_id IN ("+plantsQuery+")", args...).Error
}

var (
	errNoSeeds        = errors.New("no seeds of this plant type")
	errPositionTaken  = errors.New("garden position already occupied")
	errNotHarvestable = errors.New("plant is not ready for harvest")
)
