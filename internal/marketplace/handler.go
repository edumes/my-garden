package marketplace

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
)

type Handler struct {
	service      *Service
	auditService *audit.AuditService
}

func NewHandler(service *Service, auditService *audit.AuditService) *Handler {
	return &Handler{
		service:      service,
		auditService: auditService,
	}
}

// Delivery Request Handlers

// CreateDeliveryRequest godoc
// @Summary Create a delivery request
// @Description Create a new delivery request for a specific plant type
// @Tags marketplace
// @Accept json
// @Produce json
// @Param request body CreateDeliveryRequestRequest true "Delivery request details"
// @Success 201 {object} DeliveryRequest
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /marketplace/delivery-requests [post]
func (h *Handler) CreateDeliveryRequest(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateDeliveryRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse deadline
	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format"})
		return
	}

	request, err := h.service.CreateDeliveryRequest(userID, req.PlantTypeID, req.Reward, deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// GetDeliveryRequests godoc
// @Summary Get open delivery requests
// @Description Get all open delivery requests
// @Tags marketplace
// @Produce json
// @Success 200 {object} DeliveryRequestsResponse
// @Router /marketplace/delivery-requests [get]
func (h *Handler) GetDeliveryRequests(c *gin.Context) {
	requests, err := h.service.GetOpenDeliveryRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"delivery_requests": requests,
		"total":             len(requests),
	})
}

// AcceptDeliveryRequest godoc
// @Summary Accept a delivery request
// @Description Accept an open delivery request
// @Tags marketplace
// @Param id path string true "Request ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /marketplace/delivery-requests/{id}/accept [post]
func (h *Handler) AcceptDeliveryRequest(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	if err := h.service.AcceptDeliveryRequest(requestID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request accepted successfully"})
}

// CompleteDeliveryRequest godoc
// @Summary Complete a delivery request
// @Description Complete an accepted delivery request
// @Tags marketplace
// @Param id path string true "Request ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /marketplace/delivery-requests/{id}/complete [post]
func (h *Handler) CompleteDeliveryRequest(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	if err := h.service.CompleteDeliveryRequest(requestID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request completed successfully"})
}

// Plant Listing Handlers

// CreatePlantListing godoc
// @Summary Create a plant listing
// @Description Create a new plant listing for sale
// @Tags marketplace
// @Accept json
// @Produce json
// @Param request body CreatePlantListingRequest true "Plant listing details"
// @Success 201 {object} PlantListing
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /marketplace/listings [post]
func (h *Handler) CreatePlantListing(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreatePlantListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	listing, err := h.service.CreatePlantListing(userID, req.PlantID, req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, listing)
}

// GetPlantListings godoc
// @Summary Get active plant listings
// @Description Get all active plant listings with optional filters
// @Tags marketplace
// @Produce json
// @Param plant_type_id query string false "Plant type ID filter"
// @Param min_price query int false "Minimum price filter"
// @Param max_price query int false "Maximum price filter"
// @Success 200 {object} PlantListingsResponse
// @Router /marketplace/listings [get]
func (h *Handler) GetPlantListings(c *gin.Context) {
	filters := make(map[string]interface{})

	if plantTypeIDStr := c.Query("plant_type_id"); plantTypeIDStr != "" {
		if plantTypeID, err := uuid.Parse(plantTypeIDStr); err == nil {
			filters["plant_type_id"] = plantTypeID
		}
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.Atoi(minPriceStr); err == nil {
			filters["min_price"] = minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.Atoi(maxPriceStr); err == nil {
			filters["max_price"] = maxPrice
		}
	}

	listings, err := h.service.GetActiveListings(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"listings": listings,
		"total":    len(listings),
	})
}

// PurchasePlantListing godoc
// @Summary Purchase a plant listing
// @Description Purchase a plant from a listing
// @Tags marketplace
// @Param id path string true "Listing ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /marketplace/listings/{id}/buy [post]
func (h *Handler) PurchasePlantListing(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	listingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid listing ID"})
		return
	}

	if err := h.service.PurchasePlantListing(listingID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plant purchased successfully"})
}

// Blockchain Handlers

// SubmitTransaction godoc
// @Summary Submit a blockchain transaction
// @Description Submit a signed transaction to the blockchain
// @Tags blockchain
// @Accept json
// @Produce json
// @Param transaction body Transaction true "Transaction details"
// @Success 200 {object} Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /blockchain/transactions [post]
func (h *Handler) SubmitTransaction(c *gin.Context) {
	var transaction Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SubmitTransaction(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// GetTransactionStatus godoc
// @Summary Get transaction status
// @Description Get the status of a transaction by hash
// @Tags blockchain
// @Produce json
// @Param hash path string true "Transaction hash"
// @Success 200 {object} Transaction
// @Failure 404 {object} ErrorResponse
// @Router /blockchain/transactions/{hash} [get]
func (h *Handler) GetTransactionStatus(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction hash is required"})
		return
	}

	transaction, err := h.service.GetTransactionStatus(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// GetBlockchainLedger godoc
// @Summary Get blockchain ledger
// @Description Get the blockchain ledger with pagination
// @Tags blockchain
// @Produce json
// @Param limit query int false "Number of blocks to return (default 10)"
// @Param offset query int false "Number of blocks to skip (default 0)"
// @Success 200 {array} Block
// @Router /blockchain/ledger [get]
func (h *Handler) GetBlockchainLedger(c *gin.Context) {
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	blocks, err := h.service.GetBlockchainLedger(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, blocks)
}

// Wallet Handlers

// CreateWallet godoc
// @Summary Create a wallet
// @Description Create a new wallet for the authenticated user
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body CreateWalletRequest true "Wallet creation details"
// @Success 201 {object} Wallet
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /wallet [post]
func (h *Handler) CreateWallet(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := h.service.CreateWallet(userID, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

// GetWalletBalance godoc
// @Summary Get wallet balance
// @Description Get the current user's wallet balance and details
// @Tags wallet
// @Produce json
// @Success 200 {object} WalletResponse
// @Failure 401 {object} ErrorResponse
// @Router /wallet/balance [get]
func (h *Handler) GetWalletBalance(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	wallet, err := h.service.repo.GetWallet(userID)
	if err != nil {
		// If wallet doesn't exist, return empty wallet
		c.JSON(http.StatusOK, gin.H{"wallet": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"wallet": wallet})
}

// TransferCurrency godoc
// @Summary Transfer currency
// @Description Transfer currency to another user's wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body TransferCurrencyRequest true "Transfer details"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /wallet/transfer [post]
func (h *Handler) TransferCurrency(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req TransferCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.TransferCurrency(userID, req.ToPublicKey, req.Amount, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transfer completed successfully"})
}

// SignTransaction godoc
// @Summary Sign a transaction
// @Description Generate a signature for a transaction (offline operation)
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body SignTransactionRequest true "Transaction to sign"
// @Success 200 {object} SignTransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /wallet/sign [post]
func (h *Handler) SignTransaction(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SignTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user's wallet
	wallet, err := h.service.repo.GetWallet(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wallet not found"})
		return
	}

	// Decrypt private key
	privateKey, err := h.service.decryptPrivateKey(wallet.EncryptedKey, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	// Create transaction object
	transaction := &Transaction{
		SenderKey:   wallet.PublicKey,
		ReceiverKey: req.ReceiverKey,
		PlantID:     req.PlantID,
		Amount:      req.Amount,
		TxType:      req.TxType,
		Nonce:       req.Nonce,
	}

	// Generate signature
	signature, err := SignTransaction(privateKey, transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"signature": signature,
		"hash":      transaction.GenerateHash(),
	})
}

// GetMarketplaceStats godoc
// @Summary Get marketplace statistics
// @Description Get overall marketplace statistics including active requests, listings, and volume
// @Tags marketplace
// @Produce json
// @Success 200 {object} MarketplaceStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /marketplace/stats [get]
func (h *Handler) GetMarketplaceStats(c *gin.Context) {
	stats, err := h.service.GetMarketplaceStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Helper functions

func getUserIDFromContext(c *gin.Context) uuid.UUID {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}

	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}

	return userID
}

// Request/Response structures

type CreateDeliveryRequestRequest struct {
	PlantTypeID uuid.UUID `json:"plant_type_id" binding:"required"`
	Reward      int       `json:"reward" binding:"required,min=10"`
	Deadline    string    `json:"deadline" binding:"required"`
}

type CreatePlantListingRequest struct {
	PlantID uuid.UUID `json:"plant_id" binding:"required"`
	Price   int       `json:"price" binding:"required,min=1"`
}

type CreateWalletRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

type TransferCurrencyRequest struct {
	ToPublicKey string `json:"to_public_key" binding:"required"`
	Amount      int    `json:"amount" binding:"required,min=1"`
	Password    string `json:"password" binding:"required"`
}

type SignTransactionRequest struct {
	ReceiverKey string          `json:"receiver_key" binding:"required"`
	PlantID     *uuid.UUID      `json:"plant_id"`
	Amount      int             `json:"amount" binding:"required,min=1"`
	TxType      TransactionType `json:"tx_type" binding:"required"`
	Nonce       int64           `json:"nonce" binding:"required"`
	Password    string          `json:"password" binding:"required"`
}

type BalanceResponse struct {
	Balance int `json:"balance"`
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
