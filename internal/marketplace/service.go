package marketplace

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
)

type Service struct {
	repo         *Repository
	auditService *audit.AuditService
}

func NewService(repo *Repository, auditService *audit.AuditService) *Service {
	return &Service{
		repo:         repo,
		auditService: auditService,
	}
}

// DeliveryRequest methods
func (s *Service) CreateDeliveryRequest(userID uuid.UUID, plantTypeID uuid.UUID, reward int, deadline time.Time) (*DeliveryRequest, error) {
	// Validate reward amount
	if reward < 10 {
		return nil, fmt.Errorf("reward must be at least 10 coins")
	}

	// Validate deadline
	if deadline.Before(time.Now().Add(time.Hour)) {
		return nil, fmt.Errorf("deadline must be at least 1 hour in the future")
	}

	request := &DeliveryRequest{
		RequesterID: userID,
		PlantTypeID: plantTypeID,
		Reward:      reward,
		Deadline:    deadline,
		Status:      RequestStatusOpen,
	}

	if err := s.repo.CreateDeliveryRequest(request); err != nil {
		return nil, fmt.Errorf("failed to create delivery request: %w", err)
	}

	return request, nil
}

func (s *Service) GetOpenDeliveryRequests() ([]DeliveryRequest, error) {
	return s.repo.GetOpenDeliveryRequests()
}

func (s *Service) AcceptDeliveryRequest(requestID uuid.UUID, acceptorID uuid.UUID) error {
	request, err := s.repo.GetDeliveryRequest(requestID)
	if err != nil {
		return fmt.Errorf("failed to get delivery request: %w", err)
	}

	if request.Status != RequestStatusOpen {
		return fmt.Errorf("request is not open for acceptance")
	}

	if request.RequesterID == acceptorID {
		return fmt.Errorf("cannot accept your own request")
	}

	// Check if user has enough balance for the reward - create wallet if it doesn't exist
	wallet, err := s.repo.GetWallet(acceptorID)
	if err != nil {
		// If wallet doesn't exist, create one with default balance
		if err == gorm.ErrRecordNotFound {
			// Generate key pair for new wallet
			_, publicKey, err := GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair for new wallet: %w", err)
			}

			// Create wallet with default balance (no encryption for auto-created wallet)
			wallet = &Wallet{
				UserID:     acceptorID,
				PublicKey:  PublicKeyToString(publicKey),
				Balance:    100, // Default starting balance
				Reputation: 0,
				Nonce:      0,
			}

			if err := s.repo.CreateWallet(wallet); err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get wallet: %w", err)
		}
	}

	if wallet.Balance < request.Reward {
		return fmt.Errorf("insufficient balance to accept request")
	}

	request.Status = RequestStatusAccepted
	request.AcceptedBy = &acceptorID

	if err := s.repo.UpdateDeliveryRequest(request); err != nil {
		return fmt.Errorf("failed to update delivery request: %w", err)
	}

	return nil
}

func (s *Service) CompleteDeliveryRequest(requestID uuid.UUID, userID uuid.UUID) error {
	request, err := s.repo.GetDeliveryRequest(requestID)
	if err != nil {
		return fmt.Errorf("failed to get delivery request: %w", err)
	}

	if request.Status != RequestStatusAccepted {
		return fmt.Errorf("request is not in accepted status")
	}

	if request.AcceptedBy == nil || *request.AcceptedBy != userID {
		return fmt.Errorf("only the acceptor can complete the request")
	}

	// Check if deadline has passed
	if time.Now().After(request.Deadline) {
		// Apply penalty
		penalty := int(float64(request.Reward) * 0.2) // 20% penalty
		request.Reward -= penalty

		// Update reputation
		currentRep, _ := s.repo.GetUserReputation(userID)
		s.repo.UpdateUserReputation(userID, currentRep-0.5)
	}

	request.Status = RequestStatusCompleted
	now := time.Now()
	request.CompletedAt = &now

	if err := s.repo.UpdateDeliveryRequest(request); err != nil {
		return fmt.Errorf("failed to update delivery request: %w", err)
	}

	// Get acceptor wallet for reward - create wallet if it doesn't exist
	acceptorWallet, err := s.repo.GetWallet(userID)
	if err != nil {
		// If wallet doesn't exist, create one with default balance
		if err == gorm.ErrRecordNotFound {
			// Generate key pair for new wallet
			_, publicKey, err := GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair for new wallet: %w", err)
			}

			// Create wallet with default balance (no encryption for auto-created wallet)
			acceptorWallet = &Wallet{
				UserID:     userID,
				PublicKey:  PublicKeyToString(publicKey),
				Balance:    100, // Default starting balance
				Reputation: 0,
				Nonce:      0,
			}

			if err := s.repo.CreateWallet(acceptorWallet); err != nil {
				return fmt.Errorf("failed to create acceptor wallet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get acceptor wallet: %w", err)
		}
	}

	// Create reward transaction
	rewardTx := &Transaction{
		SenderKey:   "system", // System wallet
		ReceiverKey: acceptorWallet.PublicKey,
		Amount:      request.Reward,
		TxType:      TransactionTypeReward,
		Nonce:       time.Now().UnixNano(),
		Status:      TransactionStatusPending,
	}

	if err := s.repo.CreateTransaction(rewardTx); err != nil {
		return fmt.Errorf("failed to create reward transaction: %w", err)
	}

	// Update balances
	s.repo.UpdateWalletBalance(request.RequesterID, -request.Reward)
	s.repo.UpdateWalletBalance(userID, request.Reward)

	// Update reputation for successful completion
	currentRep, _ := s.repo.GetUserReputation(userID)
	s.repo.UpdateUserReputation(userID, currentRep+0.1)

	return nil
}

// PlantListing methods
func (s *Service) CreatePlantListing(userID uuid.UUID, plantID uuid.UUID, price int) (*PlantListing, error) {
	// Validate price
	if price < 1 {
		return nil, fmt.Errorf("price must be at least 1 coin")
	}

	// Check if user owns the plant
	var plant garden.Plant
	if err := s.repo.db.Preload("Garden").First(&plant, plantID).Error; err != nil {
		return nil, fmt.Errorf("plant not found: %w", err)
	}

	if plant.Garden.UserID != userID {
		return nil, fmt.Errorf("you can only list your own plants")
	}

	// Check if plant is harvestable
	if plant.Stage != garden.PlantStageHarvestable {
		return nil, fmt.Errorf("only harvestable plants can be listed")
	}

	listing := &PlantListing{
		SellerID:  userID,
		PlantID:   plantID,
		Price:     price,
		Status:    ListingStatusActive,
		ExpiresAt: time.Now().AddDate(0, 0, 7), // 7 days
	}

	if err := s.repo.CreatePlantListing(listing); err != nil {
		return nil, fmt.Errorf("failed to create plant listing: %w", err)
	}

	return listing, nil
}

func (s *Service) GetActiveListings(filters map[string]interface{}) ([]PlantListing, error) {
	return s.repo.GetActiveListings(filters)
}

func (s *Service) PurchasePlantListing(listingID uuid.UUID, buyerID uuid.UUID) error {
	listing, err := s.repo.GetPlantListing(listingID)
	if err != nil {
		return fmt.Errorf("failed to get plant listing: %w", err)
	}

	if listing.Status != ListingStatusActive {
		return fmt.Errorf("listing is not active")
	}

	if listing.SellerID == buyerID {
		return fmt.Errorf("cannot purchase your own listing")
	}

	// Check buyer's balance - create wallet if it doesn't exist
	buyerWallet, err := s.repo.GetWallet(buyerID)
	if err != nil {
		// If wallet doesn't exist, create one with default balance
		if err == gorm.ErrRecordNotFound {
			// Generate key pair for new wallet
			_, publicKey, err := GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair for new wallet: %w", err)
			}

			// Create wallet with default balance (no encryption for auto-created wallet)
			buyerWallet = &Wallet{
				UserID:     buyerID,
				PublicKey:  PublicKeyToString(publicKey),
				Balance:    100, // Default starting balance
				Reputation: 0,
				Nonce:      0,
			}

			if err := s.repo.CreateWallet(buyerWallet); err != nil {
				return fmt.Errorf("failed to create buyer wallet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get buyer wallet: %w", err)
		}
	}

	if buyerWallet.Balance < listing.Price {
		return fmt.Errorf("insufficient balance")
	}

	// Check seller's balance - create wallet if it doesn't exist
	sellerWallet, err := s.repo.GetWallet(listing.SellerID)
	if err != nil {
		// If wallet doesn't exist, create one with default balance
		if err == gorm.ErrRecordNotFound {
			// Generate key pair for new wallet
			_, publicKey, err := GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate key pair for seller wallet: %w", err)
			}

			// Create wallet with default balance (no encryption for auto-created wallet)
			sellerWallet = &Wallet{
				UserID:     listing.SellerID,
				PublicKey:  PublicKeyToString(publicKey),
				Balance:    100, // Default starting balance
				Reputation: 0,
				Nonce:      0,
			}

			if err := s.repo.CreateWallet(sellerWallet); err != nil {
				return fmt.Errorf("failed to create seller wallet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get seller wallet: %w", err)
		}
	}

	// Calculate commission (0.5%)
	commission := int(float64(listing.Price) * 0.005)
	sellerAmount := listing.Price - commission

	// Create purchase transaction
	purchaseTx := &Transaction{
		SenderKey:   buyerWallet.PublicKey,
		ReceiverKey: sellerWallet.PublicKey,
		PlantID:     &listing.PlantID,
		Amount:      listing.Price,
		TxType:      TransactionTypePurchase,
		Nonce:       time.Now().UnixNano(),
		Status:      TransactionStatusPending,
	}

	if err := s.repo.CreateTransaction(purchaseTx); err != nil {
		return fmt.Errorf("failed to create purchase transaction: %w", err)
	}

	// Update balances
	s.repo.UpdateWalletBalance(buyerID, -listing.Price)
	s.repo.UpdateWalletBalance(listing.SellerID, sellerAmount)

	// Transfer plant ownership - find or create a garden for the buyer
	var buyerGarden garden.Garden
	if err := s.repo.db.Where("user_id = ?", buyerID).First(&buyerGarden).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a default garden for the buyer if they don't have one
			buyerGarden = garden.Garden{
				UserID:      buyerID,
				Name:        "My Garden",
				Description: "Default garden for purchased plants",
			}
			if err := s.repo.db.Create(&buyerGarden).Error; err != nil {
				return fmt.Errorf("failed to create garden for buyer: %w", err)
			}
		} else {
			return fmt.Errorf("failed to find buyer's garden: %w", err)
		}
	}

	// Update plant's garden_id to the buyer's garden
	if err := s.repo.db.Model(&garden.Plant{}).Where("id = ?", listing.PlantID).
		Update("garden_id", buyerGarden.ID).Error; err != nil {
		return fmt.Errorf("failed to transfer plant ownership: %w", err)
	}

	// Update listing status
	listing.Status = ListingStatusSold
	if err := s.repo.UpdatePlantListing(listing); err != nil {
		return fmt.Errorf("failed to update listing status: %w", err)
	}

	// Update reputations
	buyerRep, _ := s.repo.GetUserReputation(buyerID)
	sellerRep, _ := s.repo.GetUserReputation(listing.SellerID)
	s.repo.UpdateUserReputation(buyerID, buyerRep+0.1)
	s.repo.UpdateUserReputation(listing.SellerID, sellerRep+0.1)

	return nil
}

// Blockchain methods
func (s *Service) SubmitTransaction(transaction *Transaction) error {
	// Verify transaction signature
	senderWallet, err := s.repo.GetWalletByPublicKey(transaction.SenderKey)
	if err != nil {
		return fmt.Errorf("sender wallet not found: %w", err)
	}

	publicKey, err := PublicKeyFromString(senderWallet.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	if !VerifySignature(publicKey, transaction, transaction.Signature) {
		return fmt.Errorf("invalid transaction signature")
	}

	// Check nonce
	if transaction.Nonce <= senderWallet.Nonce {
		return fmt.Errorf("invalid nonce")
	}

	// Check balance for transfers
	if transaction.TxType == TransactionTypeTransfer {
		if senderWallet.Balance < transaction.Amount {
			return fmt.Errorf("insufficient balance")
		}
	}

	// Create transaction
	if err := s.repo.CreateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update sender nonce
	senderWallet.Nonce = transaction.Nonce
	if err := s.repo.UpdateWallet(senderWallet); err != nil {
		return fmt.Errorf("failed to update sender nonce: %w", err)
	}

	return nil
}

func (s *Service) GetTransactionStatus(hash string) (*Transaction, error) {
	return s.repo.GetTransaction(hash)
}

func (s *Service) GetBlockchainLedger(limit, offset int) ([]Block, error) {
	return s.repo.GetBlockchainLedger(limit, offset)
}

// Wallet methods
func (s *Service) CreateWallet(userID uuid.UUID, password string) (*Wallet, error) {
	// Generate key pair
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Encrypt private key
	encryptedKey, err := s.encryptPrivateKey(privateKey, password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	wallet := &Wallet{
		UserID:       userID,
		PublicKey:    PublicKeyToString(publicKey),
		EncryptedKey: encryptedKey,
		Balance:      100, // Starting balance
		Reputation:   0,
		Nonce:        0,
	}

	if err := s.repo.CreateWallet(wallet); err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

func (s *Service) GetWalletBalance(userID uuid.UUID) (int, error) {
	return s.repo.GetWalletBalance(userID)
}

func (s *Service) TransferCurrency(fromUserID uuid.UUID, toPublicKey string, amount int, password string) error {
	// Get sender wallet
	senderWallet, err := s.repo.GetWallet(fromUserID)
	if err != nil {
		return fmt.Errorf("sender wallet not found: %w", err)
	}

	// Decrypt private key
	privateKey, err := s.decryptPrivateKey(senderWallet.EncryptedKey, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// Get receiver wallet
	receiverWallet, err := s.repo.GetWalletByPublicKey(toPublicKey)
	if err != nil {
		return fmt.Errorf("receiver wallet not found: %w", err)
	}

	// Check balance
	if senderWallet.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	// Create transaction
	transaction := &Transaction{
		SenderKey:   senderWallet.PublicKey,
		ReceiverKey: receiverWallet.PublicKey,
		Amount:      amount,
		TxType:      TransactionTypeTransfer,
		Nonce:       senderWallet.Nonce + 1,
		Status:      TransactionStatusPending,
	}

	// Sign transaction
	signature, err := SignTransaction(privateKey, transaction)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	transaction.Signature = signature

	// Submit transaction
	if err := s.SubmitTransaction(transaction); err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	// Update balances
	s.repo.UpdateWalletBalance(fromUserID, -amount)
	s.repo.UpdateWalletBalance(receiverWallet.UserID, amount)

	// Update sender nonce
	senderWallet.Nonce++
	s.repo.UpdateWallet(senderWallet)

	return nil
}

// Helper methods for encryption/decryption
func (s *Service) encryptPrivateKey(privateKey *ecdsa.PrivateKey, password string) (string, error) {
	// Convert private key to bytes
	privateKeyBytes := privateKey.D.Bytes()

	// Generate key from password
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, privateKeyBytes, nil)
	return hex.EncodeToString(ciphertext), nil
}

func (s *Service) decryptPrivateKey(encryptedKey string, password string) (*ecdsa.PrivateKey, error) {
	// Decode hex string
	ciphertext, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return nil, err
	}

	// Generate key from password
	key := sha256.Sum256([]byte(password))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Reconstruct private key
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
		},
		D: new(big.Int).SetBytes(plaintext),
	}

	return privateKey, nil
}

// Consensus methods
func (s *Service) ValidateTransaction(transaction *Transaction) bool {
	// Basic validation
	if transaction.Amount <= 0 {
		return false
	}

	// Check if transaction already exists
	existingTx, err := s.repo.GetTransaction(transaction.Hash)
	if err == nil && existingTx != nil {
		return false // Transaction already exists
	}

	// Verify signature
	senderWallet, err := s.repo.GetWalletByPublicKey(transaction.SenderKey)
	if err != nil {
		return false
	}

	publicKey, err := PublicKeyFromString(senderWallet.PublicKey)
	if err != nil {
		return false
	}

	return VerifySignature(publicKey, transaction, transaction.Signature)
}

func (s *Service) ProcessPendingTransactions() error {
	transactions, err := s.repo.GetPendingTransactions()
	if err != nil {
		return fmt.Errorf("failed to get pending transactions: %w", err)
	}

	for _, tx := range transactions {
		if s.ValidateTransaction(&tx) {
			tx.Status = TransactionStatusConfirmed
			if err := s.repo.UpdateTransaction(&tx); err != nil {
				return fmt.Errorf("failed to update transaction: %w", err)
			}
		} else {
			tx.Status = TransactionStatusRejected
			if err := s.repo.UpdateTransaction(&tx); err != nil {
				return fmt.Errorf("failed to update transaction: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) GetMarketplaceStats() (*MarketplaceStatsResponse, error) {
	// Get active delivery requests count
	activeRequests, err := s.repo.GetOpenDeliveryRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to get active delivery requests: %w", err)
	}

	// Get all delivery requests count
	allRequests, err := s.repo.GetAllDeliveryRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to get all delivery requests: %w", err)
	}

	// Get active listings count
	activeListings, err := s.repo.GetActiveListings(map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get active listings: %w", err)
	}

	// Get all listings count
	allListings, err := s.repo.GetAllListings()
	if err != nil {
		return nil, fmt.Errorf("failed to get all listings: %w", err)
	}

	// Get total volume and transaction count
	totalVolume, totalTransactions, err := s.repo.GetTransactionStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction stats: %w", err)
	}

	return &MarketplaceStatsResponse{
		ActiveDeliveryRequests: len(activeRequests),
		TotalDeliveryRequests:  len(allRequests),
		ActiveListings:         len(activeListings),
		TotalListings:          len(allListings),
		TotalVolume:            totalVolume,
		TotalTransactions:      totalTransactions,
	}, nil
}
