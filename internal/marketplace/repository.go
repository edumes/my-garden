package marketplace

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

// DeliveryRequest methods
func (r *Repository) CreateDeliveryRequest(request *DeliveryRequest) error {
	return r.db.Create(request).Error
}

func (r *Repository) GetDeliveryRequest(id uuid.UUID) (*DeliveryRequest, error) {
	var request DeliveryRequest
	err := r.db.Preload("Requester").Preload("Acceptor").Preload("PlantType").First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *Repository) GetOpenDeliveryRequests() ([]DeliveryRequest, error) {
	var requests []DeliveryRequest
	err := r.db.Preload("Requester").Preload("PlantType").
		Where("status = ? AND deadline > ?", RequestStatusOpen, time.Now()).
		Order("created_at DESC").Find(&requests).Error
	return requests, err
}

func (r *Repository) UpdateDeliveryRequest(request *DeliveryRequest) error {
	return r.db.Save(request).Error
}

func (r *Repository) GetUserDeliveryRequests(userID uuid.UUID) ([]DeliveryRequest, error) {
	var requests []DeliveryRequest
	err := r.db.Preload("Requester").Preload("Acceptor").Preload("PlantType").
		Where("requester_id = ? OR accepted_by = ?", userID, userID).
		Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// PlantListing methods
func (r *Repository) CreatePlantListing(listing *PlantListing) error {
	return r.db.Create(listing).Error
}

func (r *Repository) GetPlantListing(id uuid.UUID) (*PlantListing, error) {
	var listing PlantListing
	err := r.db.Preload("Seller").Preload("Plant").Preload("Plant.PlantType").
		First(&listing, id).Error
	if err != nil {
		return nil, err
	}
	return &listing, nil
}

func (r *Repository) GetActiveListings(filters map[string]interface{}) ([]PlantListing, error) {
	query := r.db.Preload("Seller").Preload("Plant").Preload("Plant.PlantType").
		Where("status = ? AND expires_at > ?", ListingStatusActive, time.Now())

	// Apply filters
	if plantTypeID, ok := filters["plant_type_id"].(uuid.UUID); ok {
		query = query.Joins("JOIN plants ON plant_listings.plant_id = plants.id").
			Where("plants.plant_type_id = ?", plantTypeID)
	}

	if maxPrice, ok := filters["max_price"].(int); ok {
		query = query.Where("price <= ?", maxPrice)
	}

	if minPrice, ok := filters["min_price"].(int); ok {
		query = query.Where("price >= ?", minPrice)
	}

	var listings []PlantListing
	err := query.Order("created_at DESC").Find(&listings).Error
	return listings, err
}

func (r *Repository) UpdatePlantListing(listing *PlantListing) error {
	return r.db.Save(listing).Error
}

func (r *Repository) GetUserListings(userID uuid.UUID) ([]PlantListing, error) {
	var listings []PlantListing
	err := r.db.Preload("Seller").Preload("Plant").Preload("Plant.PlantType").
		Where("seller_id = ?", userID).
		Order("created_at DESC").Find(&listings).Error
	return listings, err
}

// Transaction methods
func (r *Repository) CreateTransaction(transaction *Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *Repository) GetTransaction(hash string) (*Transaction, error) {
	var transaction Transaction
	err := r.db.Preload("Plant").First(&transaction, "hash = ?", hash).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *Repository) GetPendingTransactions() ([]Transaction, error) {
	var transactions []Transaction
	err := r.db.Preload("Plant").
		Where("status = ?", TransactionStatusPending).
		Order("created_at ASC").Find(&transactions).Error
	return transactions, err
}

func (r *Repository) UpdateTransaction(transaction *Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *Repository) GetTransactionHistory(limit, offset int) ([]Transaction, error) {
	var transactions []Transaction
	err := r.db.Preload("Plant").
		Where("status = ?", TransactionStatusConfirmed).
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&transactions).Error
	return transactions, err
}

// Block methods
func (r *Repository) CreateBlock(block *Block) error {
	return r.db.Create(block).Error
}

func (r *Repository) GetLatestBlock() (*Block, error) {
	var block Block
	err := r.db.Preload("Validator").Preload("Transactions").
		Order("height DESC").First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *Repository) GetBlockByHeight(height int64) (*Block, error) {
	var block Block
	err := r.db.Preload("Validator").Preload("Transactions").
		Where("height = ?", height).First(&block).Error
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *Repository) GetBlockchainLedger(limit, offset int) ([]Block, error) {
	var blocks []Block
	err := r.db.Preload("Validator").
		Order("height DESC").
		Limit(limit).Offset(offset).Find(&blocks).Error
	return blocks, err
}

// Wallet methods
func (r *Repository) CreateWallet(wallet *Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *Repository) GetWallet(userID uuid.UUID) (*Wallet, error) {
	var wallet Wallet
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *Repository) UpdateWallet(wallet *Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *Repository) GetWalletByPublicKey(publicKey string) (*Wallet, error) {
	var wallet Wallet
	err := r.db.Preload("User").Where("public_key = ?", publicKey).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// ConsensusNode methods
func (r *Repository) CreateConsensusNode(node *ConsensusNode) error {
	return r.db.Create(node).Error
}

func (r *Repository) GetActiveConsensusNodes() ([]ConsensusNode, error) {
	var nodes []ConsensusNode
	err := r.db.Where("is_active = ?", true).Find(&nodes).Error
	return nodes, err
}

func (r *Repository) UpdateConsensusNode(node *ConsensusNode) error {
	return r.db.Save(node).Error
}

// Utility methods
func (r *Repository) GetTransactionCount() (int64, error) {
	var count int64
	err := r.db.Model(&Transaction{}).Count(&count).Error
	return count, err
}

func (r *Repository) GetBlockCount() (int64, error) {
	var count int64
	err := r.db.Model(&Block{}).Count(&count).Error
	return count, err
}

func (r *Repository) GetWalletBalance(userID uuid.UUID) (int, error) {
	var wallet Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

func (r *Repository) UpdateWalletBalance(userID uuid.UUID, amount int) error {
	return r.db.Model(&Wallet{}).Where("user_id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (r *Repository) GetUserReputation(userID uuid.UUID) (float64, error) {
	var wallet Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return 0, err
	}
	return wallet.Reputation, nil
}

func (r *Repository) UpdateUserReputation(userID uuid.UUID, reputation float64) error {
	return r.db.Model(&Wallet{}).Where("user_id = ?", userID).Update("reputation", reputation).Error
}

// Marketplace stats methods
func (r *Repository) GetAllDeliveryRequests() ([]DeliveryRequest, error) {
	var requests []DeliveryRequest
	err := r.db.Preload("Requester").Preload("PlantType").
		Order("created_at DESC").Find(&requests).Error
	return requests, err
}

func (r *Repository) GetAllListings() ([]PlantListing, error) {
	var listings []PlantListing
	err := r.db.Preload("Seller").Preload("Plant").Preload("Plant.PlantType").
		Order("created_at DESC").Find(&listings).Error
	return listings, err
}

func (r *Repository) GetTransactionStats() (int, int, error) {
	var totalVolume int
	var totalTransactions int64

	// Get total volume from all completed transactions
	err := r.db.Model(&Transaction{}).
		Where("status = ?", TransactionStatusConfirmed).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalVolume).Error
	if err != nil {
		return 0, 0, err
	}

	// Get total transaction count
	err = r.db.Model(&Transaction{}).Count(&totalTransactions).Error
	if err != nil {
		return 0, 0, err
	}

	return totalVolume, int(totalTransactions), nil
}

// Cleanup expired listings and requests
func (r *Repository) CleanupExpiredItems() error {
	now := time.Now()

	// Update expired listings
	if err := r.db.Model(&PlantListing{}).
		Where("status = ? AND expires_at <= ?", ListingStatusActive, now).
		Update("status", ListingStatusExpired).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired listings: %w", err)
	}

	// Update expired delivery requests
	if err := r.db.Model(&DeliveryRequest{}).
		Where("status = ? AND deadline <= ?", RequestStatusOpen, now).
		Update("status", RequestStatusExpired).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired delivery requests: %w", err)
	}

	return nil
}
