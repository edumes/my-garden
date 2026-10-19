package marketplace

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test cases for core functionality

func TestGenerateKeyPair(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()

	assert.NoError(t, err)
	assert.NotNil(t, privateKey)
	assert.NotNil(t, publicKey)
	assert.Equal(t, privateKey.PublicKey, *publicKey)
}

func TestSignAndVerifyTransaction(t *testing.T) {
	// Generate key pair
	privateKey, publicKey, err := GenerateKeyPair()
	assert.NoError(t, err)

	// Create transaction
	transaction := &Transaction{
		SenderKey:   "sender_key",
		ReceiverKey: "receiver_key",
		Amount:      100,
		TxType:      TransactionTypeTransfer,
		Nonce:       1234567890,
	}

	// Sign transaction
	signature, err := SignTransaction(privateKey, transaction)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Verify signature
	isValid := VerifySignature(publicKey, transaction, signature)
	assert.True(t, isValid)

	// Test invalid signature
	isValid = VerifySignature(publicKey, transaction, "invalid_signature")
	assert.False(t, isValid)
}

func TestTransactionHashGeneration(t *testing.T) {
	transaction := &Transaction{
		SenderKey:   "sender_key",
		ReceiverKey: "receiver_key",
		Amount:      100,
		TxType:      TransactionTypeTransfer,
		Nonce:       1234567890,
	}

	hash1 := transaction.GenerateHash()
	hash2 := transaction.GenerateHash()

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2) // Same transaction should generate same hash

	// Different transaction should generate different hash
	transaction2 := &Transaction{
		SenderKey:   "sender_key",
		ReceiverKey: "receiver_key",
		Amount:      200, // Different amount
		TxType:      TransactionTypeTransfer,
		Nonce:       1234567890,
	}

	hash3 := transaction2.GenerateHash()
	assert.NotEqual(t, hash1, hash3)
}

func TestPublicKeyConversion(t *testing.T) {
	// Generate key pair
	_, publicKey, err := GenerateKeyPair()
	assert.NoError(t, err)

	// Convert to string
	publicKeyStr := PublicKeyToString(publicKey)
	assert.NotEmpty(t, publicKeyStr)

	// Convert back from string
	recoveredPublicKey, err := PublicKeyFromString(publicKeyStr)
	assert.NoError(t, err)
	assert.NotNil(t, recoveredPublicKey)

	// Verify they're the same
	assert.Equal(t, publicKey.X, recoveredPublicKey.X)
	assert.Equal(t, publicKey.Y, recoveredPublicKey.Y)
	assert.Equal(t, publicKey.Curve, recoveredPublicKey.Curve)
}

func TestWalletCreation(t *testing.T) {
	// Test wallet creation
	wallet := &Wallet{
		UserID:     uuid.New(),
		PublicKey:  "test_public_key",
		Balance:    100,
		Reputation: 0,
		Nonce:      0,
	}

	assert.NotNil(t, wallet)
	assert.Equal(t, 100, wallet.Balance)
	assert.Equal(t, 0.0, wallet.Reputation)
	assert.Equal(t, int64(0), wallet.Nonce)
}

func TestTransactionTypes(t *testing.T) {
	// Test transaction type constants
	assert.Equal(t, TransactionType("transfer"), TransactionTypeTransfer)
	assert.Equal(t, TransactionType("purchase"), TransactionTypePurchase)
	assert.Equal(t, TransactionType("reward"), TransactionTypeReward)
	assert.Equal(t, TransactionType("commission"), TransactionTypeCommission)
}

func TestRequestStatuses(t *testing.T) {
	// Test request status constants
	assert.Equal(t, RequestStatus("open"), RequestStatusOpen)
	assert.Equal(t, RequestStatus("accepted"), RequestStatusAccepted)
	assert.Equal(t, RequestStatus("completed"), RequestStatusCompleted)
	assert.Equal(t, RequestStatus("expired"), RequestStatusExpired)
	assert.Equal(t, RequestStatus("cancelled"), RequestStatusCancelled)
}

func TestListingStatuses(t *testing.T) {
	// Test listing status constants
	assert.Equal(t, ListingStatus("active"), ListingStatusActive)
	assert.Equal(t, ListingStatus("sold"), ListingStatusSold)
	assert.Equal(t, ListingStatus("expired"), ListingStatusExpired)
	assert.Equal(t, ListingStatus("cancelled"), ListingStatusCancelled)
}

func TestTransactionStatuses(t *testing.T) {
	// Test transaction status constants
	assert.Equal(t, TransactionStatus("pending"), TransactionStatusPending)
	assert.Equal(t, TransactionStatus("confirmed"), TransactionStatusConfirmed)
	assert.Equal(t, TransactionStatus("rejected"), TransactionStatusRejected)
}

func TestDeliveryRequestValidation(t *testing.T) {
	// Test delivery request validation
	userID := uuid.New()
	plantTypeID := uuid.New()
	validDeadline := time.Now().Add(time.Hour * 2)
	invalidDeadline := time.Now().Add(time.Minute * 30)

	// Test valid request
	request := &DeliveryRequest{
		RequesterID: userID,
		PlantTypeID: plantTypeID,
		Reward:      100,
		Deadline:    validDeadline,
		Status:      RequestStatusOpen,
	}

	assert.Equal(t, userID, request.RequesterID)
	assert.Equal(t, plantTypeID, request.PlantTypeID)
	assert.Equal(t, 100, request.Reward)
	assert.Equal(t, RequestStatusOpen, request.Status)

	// Test invalid reward
	request.Reward = 5
	assert.Less(t, request.Reward, 10)

	// Test invalid deadline
	assert.True(t, invalidDeadline.Before(time.Now().Add(time.Hour)))
}

func TestPlantListingValidation(t *testing.T) {
	// Test plant listing validation
	userID := uuid.New()
	plantID := uuid.New()
	validPrice := 50
	invalidPrice := 0

	// Test valid listing
	listing := &PlantListing{
		SellerID:  userID,
		PlantID:   plantID,
		Price:     validPrice,
		Status:    ListingStatusActive,
		ExpiresAt: time.Now().AddDate(0, 0, 7),
	}

	assert.Equal(t, userID, listing.SellerID)
	assert.Equal(t, plantID, listing.PlantID)
	assert.Equal(t, validPrice, listing.Price)
	assert.Equal(t, ListingStatusActive, listing.Status)

	// Test invalid price
	assert.Less(t, invalidPrice, 1)
}

func TestTransactionValidation(t *testing.T) {
	// Test transaction validation
	transaction := &Transaction{
		SenderKey:   "sender_key",
		ReceiverKey: "receiver_key",
		Amount:      100,
		TxType:      TransactionTypeTransfer,
		Nonce:       1234567890,
		Status:      TransactionStatusPending,
	}

	assert.Equal(t, "sender_key", transaction.SenderKey)
	assert.Equal(t, "receiver_key", transaction.ReceiverKey)
	assert.Equal(t, 100, transaction.Amount)
	assert.Equal(t, TransactionTypeTransfer, transaction.TxType)
	assert.Equal(t, TransactionStatusPending, transaction.Status)

	// Test invalid amount
	transaction.Amount = 0
	assert.LessOrEqual(t, transaction.Amount, 0)
}

func TestBlockCreation(t *testing.T) {
	// Test block creation
	block := &Block{
		Height:      1,
		Hash:        "test_hash",
		PrevHash:    "prev_hash",
		MerkleRoot:  "merkle_root",
		Timestamp:   time.Now(),
		ValidatorID: uuid.New(),
	}

	assert.Equal(t, int64(1), block.Height)
	assert.Equal(t, "test_hash", block.Hash)
	assert.Equal(t, "prev_hash", block.PrevHash)
	assert.Equal(t, "merkle_root", block.MerkleRoot)
	assert.NotNil(t, block.ValidatorID)
}

func TestConsensusNodeCreation(t *testing.T) {
	// Test consensus node creation
	node := &ConsensusNode{
		NodeID:    "node_1",
		PublicKey: "public_key_1",
		IsActive:  true,
		LastSeen:  time.Now(),
	}

	assert.Equal(t, "node_1", node.NodeID)
	assert.Equal(t, "public_key_1", node.PublicKey)
	assert.True(t, node.IsActive)
	assert.NotZero(t, node.LastSeen)
}
