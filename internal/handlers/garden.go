package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
	"gorm.io/gorm"
)

type GardenHandler struct {
	db *database.Database
}

func NewGardenHandler(db *database.Database) *GardenHandler {
	return &GardenHandler{db: db}
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
	Position    int       `json:"position" binding:"required,min=0,max=8" example:"0"`
}

type WaterPlantRequest struct {
	Amount int `json:"amount" binding:"required,min=1,max=100" example:"30"`
}

type FertilizePlantRequest struct {
	Amount int `json:"amount" binding:"required,min=1,max=100" example:"20"`
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

	var gardens []models.Garden
	if err := h.db.DB.Where("user_id = ?", userID).Preload("Plants.PlantType").Find(&gardens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gardens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"gardens": gardens})
}

// CreateGarden godoc
// @Summary Create a new garden
// @Description Create a new garden for the current user
// @Tags gardens
// @Accept json
// @Produce json
// @Security bearer
// @Param request body CreateGardenRequest true "Garden creation data"
// @Success 201 {object} map[string]interface{} "Created garden"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
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

	garden := models.Garden{
		UserID:      userID.(uuid.UUID),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.db.DB.Create(&garden).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create garden"})
		return
	}

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
// @Success 200 {object} map[string]interface{} "Garden details with plants"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid garden ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Garden not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens/{id} [get]
func (h *GardenHandler) GetGarden(c *gin.Context) {
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

	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).Preload("Plants.PlantType").First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"garden": garden})
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
// @Router /gardens/{id} [put]
func (h *GardenHandler) UpdateGarden(c *gin.Context) {
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

	var req UpdateGardenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
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
// @Router /gardens/{id} [delete]
func (h *GardenHandler) DeleteGarden(c *gin.Context) {
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

	// Check if garden exists and belongs to user
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Delete garden (this will cascade delete plants)
	if err := h.db.DB.Delete(&garden).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete garden"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Garden deleted successfully"})
}

// PlantSeed godoc
// @Summary Plant a seed
// @Description Plant a seed in a garden
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param request body PlantRequest true "Plant data"
// @Success 201 {object} map[string]interface{} "Planted seed"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Garden or plant type not found"
// @Failure 409 {object} map[string]interface{} "Position already occupied"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
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

	// Check if garden exists and belongs to user
	var garden models.Garden
	if err := h.db.DB.Where("id = ? AND user_id = ?", gardenID, userID).First(&garden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Garden not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden"})
		return
	}

	// Check if position is already occupied
	var existingPlant models.Plant
	if err := h.db.DB.Where("garden_id = ? AND position = ?", gardenID, req.Position).First(&existingPlant).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Position already occupied"})
		return
	}

	// Check if plant type exists
	var plantType models.PlantType
	if err := h.db.DB.First(&plantType, req.PlantTypeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant type not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant type"})
		return
	}

	// Create plant
	plant := models.Plant{
		GardenID:    gardenID,
		PlantTypeID: req.PlantTypeID,
		Position:    req.Position,
		PlantedAt:   time.Now(),
	}

	if err := h.db.DB.Create(&plant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to plant seed"})
		return
	}

	// Load plant type for response
	h.db.DB.Preload("PlantType").First(&plant, plant.ID)

	c.JSON(http.StatusCreated, gin.H{"plant": plant})
}

// WaterPlant godoc
// @Summary Water a plant
// @Description Water a plant to increase its water level
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param plantId path string true "Plant ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Param request body WaterPlantRequest true "Water amount"
// @Success 200 {object} map[string]interface{} "Updated plant"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Plant not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens/{id}/plants/{plantId} [put]
func (h *GardenHandler) WaterPlant(c *gin.Context) {
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

	plantID, err := uuid.Parse(c.Param("plantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant ID"})
		return
	}

	var req WaterPlantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if plant exists and belongs to user's garden
	var plant models.Plant
	if err := h.db.DB.Joins("JOIN gardens ON plants.garden_id = gardens.id").
		Where("plants.id = ? AND gardens.id = ? AND gardens.user_id = ?", plantID, gardenID, userID).
		First(&plant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant"})
		return
	}

	// Update water level
	now := time.Now()
	plant.WaterLevel = min(100, plant.WaterLevel+req.Amount)
	plant.LastWateredAt = &now

	if err := h.db.DB.Save(&plant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to water plant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plant": plant})
}

// FertilizePlant godoc
// @Summary Fertilize a plant
// @Description Apply fertilizer to a plant
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param plantId path string true "Plant ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Param request body FertilizePlantRequest true "Fertilizer amount"
// @Success 200 {object} map[string]interface{} "Updated plant"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Plant not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens/{id}/plants/{plantId}/fertilize [post]
func (h *GardenHandler) FertilizePlant(c *gin.Context) {
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

	plantID, err := uuid.Parse(c.Param("plantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant ID"})
		return
	}

	var req FertilizePlantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if plant exists and belongs to user's garden
	var plant models.Plant
	if err := h.db.DB.Joins("JOIN gardens ON plants.garden_id = gardens.id").
		Where("plants.id = ? AND gardens.id = ? AND gardens.user_id = ?", plantID, gardenID, userID).
		First(&plant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant"})
		return
	}

	// Update fertilizer level
	now := time.Now()
	plant.LastFertilizedAt = &now

	if err := h.db.DB.Save(&plant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fertilize plant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plant": plant})
}

// HarvestPlant godoc
// @Summary Harvest a plant
// @Description Harvest a mature plant to get rewards
// @Tags plants
// @Accept json
// @Produce json
// @Security bearer
// @Param id path string true "Garden ID" example("123e4567-e89b-12d3-a456-426614174000")
// @Param plantId path string true "Plant ID" example("123e4567-e89b-12d3-a456-426614174001")
// @Success 200 {object} map[string]interface{} "Harvest results"
// @Failure 400 {object} map[string]interface{} "Bad Request - Plant not ready for harvest"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Plant not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /gardens/{id}/plants/{plantId}/harvest [post]
func (h *GardenHandler) HarvestPlant(c *gin.Context) {
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

	plantID, err := uuid.Parse(c.Param("plantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant ID"})
		return
	}

	// Check if plant exists and belongs to user's garden
	var plant models.Plant
	if err := h.db.DB.Joins("JOIN gardens ON plants.garden_id = gardens.id").
		Preload("PlantType").
		Where("plants.id = ? AND gardens.id = ? AND gardens.user_id = ?", plantID, gardenID, userID).
		First(&plant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant"})
		return
	}

	// Check if plant is harvestable
	if plant.Stage != models.PlantStageHarvestable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plant is not ready for harvest"})
		return
	}

	// Get user to update coins and experience
	var user models.User
	if err := h.db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Calculate harvest rewards
	coinsEarned := plant.PlantType.HarvestValue * plant.PlantType.Yield
	experienceEarned := plant.PlantType.ExperienceValue

	// Update user stats
	user.Coins += coinsEarned
	user.Experience += experienceEarned

	// Check for level up
	oldLevel := user.Level
	user.Level = calculateLevel(user.Experience)

	// Update plant status
	now := time.Now()
	plant.HarvestedAt = &now
	plant.Stage = models.PlantStageWithered

	// Save changes in a transaction
	tx := h.db.DB.Begin()
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	if err := tx.Save(&plant).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plant"})
		return
	}

	tx.Commit()

	response := gin.H{
		"plant": plant,
		"harvest": gin.H{
			"coins_earned":      coinsEarned,
			"experience_earned": experienceEarned,
			"level_up":          user.Level > oldLevel,
			"new_level":         user.Level,
		},
	}

	c.JSON(http.StatusOK, response)
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
// @Router /gardens/{id}/plants/{plantId} [delete]
func (h *GardenHandler) RemovePlant(c *gin.Context) {
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

	plantID, err := uuid.Parse(c.Param("plantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant ID"})
		return
	}

	// Check if plant exists and belongs to user's garden
	var plant models.Plant
	if err := h.db.DB.Joins("JOIN gardens ON plants.garden_id = gardens.id").
		Where("plants.id = ? AND gardens.id = ? AND gardens.user_id = ?", plantID, gardenID, userID).
		First(&plant).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plant not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plant"})
		return
	}

	// Delete plant
	if err := h.db.DB.Delete(&plant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove plant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plant removed successfully"})
}

// Helper function to calculate level from experience
func calculateLevel(experience int) int {
	// Simple level calculation: every 100 XP = 1 level
	return (experience / 100) + 1
}

// Helper function to get minimum value
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
