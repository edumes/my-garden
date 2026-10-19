package marketplace

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// respondError maps service errors to HTTP responses.
func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrInvalidReward), errors.Is(err, ErrInvalidDeadline), errors.Is(err, ErrInvalidPrice),
		errors.Is(err, ErrOwnItem), errors.Is(err, ErrNotHarvestable), errors.Is(err, ErrInsufficientCoins):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotYourPlant):
		status = http.StatusForbidden
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrNotAvailable), errors.Is(err, ErrAlreadyListed), errors.Is(err, ErrRequesterBroke):
		status = http.StatusConflict
	}
	message := err.Error()
	if status == http.StatusInternalServerError {
		message = "Marketplace operation failed"
	}
	c.JSON(status, gin.H{"error": message})
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return uuid.Nil, false
	}
	return id, true
}

func currentUser(c *gin.Context) uuid.UUID {
	return c.MustGet("user_id").(uuid.UUID)
}

// CreateDeliveryRequest godoc
// @Summary Create a delivery request
// @Description Ask other players to deliver a plant type for a coin reward, charged on completion
// @Tags marketplace
// @Accept json
// @Produce json
// @Security bearer
// @Param request body CreateDeliveryRequestRequest true "Delivery request data"
// @Success 201 {object} DeliveryRequest
// @Failure 400 {object} ErrorResponse "Invalid reward or deadline, or not enough coins for the reward"
// @Failure 404 {object} ErrorResponse "Plant type not found"
// @Router /marketplace/delivery-requests [post]
func (h *Handler) CreateDeliveryRequest(c *gin.Context) {
	var req CreateDeliveryRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format, use RFC3339"})
		return
	}

	request, err := h.service.CreateDeliveryRequest(currentUser(c), req.PlantTypeID, req.Reward, deadline)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, request)
}

// GetDeliveryRequests godoc
// @Summary Get open delivery requests
// @Description Get all open delivery requests whose deadline has not passed
// @Tags marketplace
// @Produce json
// @Security bearer
// @Success 200 {object} DeliveryRequestsResponse
// @Router /marketplace/delivery-requests [get]
func (h *Handler) GetDeliveryRequests(c *gin.Context) {
	requests, err := h.service.GetOpenDeliveryRequests()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, DeliveryRequestsResponse{DeliveryRequests: requests, Total: len(requests)})
}

// AcceptDeliveryRequest godoc
// @Summary Accept a delivery request
// @Description Commit to delivering an open request before its deadline
// @Tags marketplace
// @Produce json
// @Security bearer
// @Param id path string true "Delivery request ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Own request"
// @Failure 404 {object} ErrorResponse "Request not found"
// @Failure 409 {object} ErrorResponse "Request already accepted or expired"
// @Router /marketplace/delivery-requests/{id}/accept [post]
func (h *Handler) AcceptDeliveryRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.AcceptDeliveryRequest(id, currentUser(c)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Message: "Request accepted successfully"})
}

// CompleteDeliveryRequest godoc
// @Summary Complete a delivery request
// @Description Complete an accepted request and receive the reward, minus 20% after the deadline
// @Tags marketplace
// @Produce json
// @Security bearer
// @Param id path string true "Delivery request ID"
// @Success 200 {object} CompleteDeliveryResponse
// @Failure 404 {object} ErrorResponse "Request not found"
// @Failure 409 {object} ErrorResponse "Not accepted by the caller, already completed, or the requester cannot pay"
// @Router /marketplace/delivery-requests/{id}/complete [post]
func (h *Handler) CompleteDeliveryRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	paid, err := h.service.CompleteDeliveryRequest(id, currentUser(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, CompleteDeliveryResponse{Message: "Request completed successfully", RewardPaid: paid})
}

// CreatePlantListing godoc
// @Summary List a plant for sale
// @Description List a harvestable plant from one of your gardens for 7 days
// @Tags marketplace
// @Accept json
// @Produce json
// @Security bearer
// @Param request body CreatePlantListingRequest true "Listing data"
// @Success 201 {object} PlantListing
// @Failure 400 {object} ErrorResponse "Invalid price or plant not harvestable"
// @Failure 403 {object} ErrorResponse "Plant is not in one of your gardens"
// @Failure 404 {object} ErrorResponse "Plant not found"
// @Failure 409 {object} ErrorResponse "Plant already listed"
// @Router /marketplace/listings [post]
func (h *Handler) CreatePlantListing(c *gin.Context) {
	var req CreatePlantListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	listing, err := h.service.CreatePlantListing(currentUser(c), req.PlantID, req.Price)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, listing)
}

// GetPlantListings godoc
// @Summary Get active plant listings
// @Description Get active, unexpired plant listings with optional filters
// @Tags marketplace
// @Produce json
// @Security bearer
// @Param plant_type_id query string false "Plant type ID filter"
// @Param min_price query int false "Minimum price filter"
// @Param max_price query int false "Maximum price filter"
// @Success 200 {object} PlantListingsResponse
// @Failure 400 {object} ErrorResponse "Invalid filter"
// @Router /marketplace/listings [get]
func (h *Handler) GetPlantListings(c *gin.Context) {
	var filters ListingFilters
	if raw := c.Query("plant_type_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plant_type_id"})
			return
		}
		filters.PlantTypeID = &id
	}
	for name, target := range map[string]**int{"min_price": &filters.MinPrice, "max_price": &filters.MaxPrice} {
		if raw := c.Query(name); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + name})
				return
			}
			*target = &value
		}
	}

	listings, err := h.service.GetActiveListings(filters)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PlantListingsResponse{Listings: listings, Total: len(listings)})
}

// PurchasePlantListing godoc
// @Summary Buy a plant listing
// @Description Pay for a listing with coins; the plant moves to a free cell in your gardens
// @Tags marketplace
// @Produce json
// @Security bearer
// @Param id path string true "Listing ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse "Own listing or not enough coins"
// @Failure 404 {object} ErrorResponse "Listing not found"
// @Failure 409 {object} ErrorResponse "Listing sold, expired, or its plant is gone"
// @Router /marketplace/listings/{id}/buy [post]
func (h *Handler) PurchasePlantListing(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.PurchasePlantListing(id, currentUser(c)); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, SuccessResponse{Message: "Plant purchased successfully"})
}

// GetMarketplaceStats godoc
// @Summary Get marketplace statistics
// @Description Counts of requests and listings, and the coin volume of completed sales
// @Tags marketplace
// @Produce json
// @Security bearer
// @Success 200 {object} MarketplaceStatsResponse
// @Router /marketplace/stats [get]
func (h *Handler) GetMarketplaceStats(c *gin.Context) {
	stats, err := h.service.GetMarketplaceStats()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

type CreateDeliveryRequestRequest struct {
	PlantTypeID uuid.UUID `json:"plant_type_id" binding:"required"`
	Reward      int       `json:"reward" binding:"required,min=10"`
	Deadline    string    `json:"deadline" binding:"required" example:"2026-01-15T18:00:00Z"`
}

type CreatePlantListingRequest struct {
	PlantID uuid.UUID `json:"plant_id" binding:"required"`
	Price   int       `json:"price" binding:"required,min=1"`
}

type DeliveryRequestsResponse struct {
	DeliveryRequests []DeliveryRequest `json:"delivery_requests"`
	Total            int               `json:"total"`
}

type PlantListingsResponse struct {
	Listings []PlantListing `json:"listings"`
	Total    int            `json:"total"`
}

type CompleteDeliveryResponse struct {
	Message    string `json:"message"`
	RewardPaid int    `json:"reward_paid"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MarketplaceStatsResponse struct {
	ActiveDeliveryRequests int `json:"active_delivery_requests"`
	TotalDeliveryRequests  int `json:"total_delivery_requests"`
	ActiveListings         int `json:"active_listings"`
	TotalListings          int `json:"total_listings"`
	TotalVolume            int `json:"total_volume"`
	TotalTransactions      int `json:"total_transactions"`
}
