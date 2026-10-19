package marketplace

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
)

// DeliveryRequest represents a service order for plant delivery
type DeliveryRequest struct {
	ID          uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RequesterID uuid.UUID     `json:"requester_id" gorm:"type:uuid;not null"`
	PlantTypeID uuid.UUID     `json:"plant_type_id" gorm:"type:uuid;not null"`
	Reward      int           `json:"reward" gorm:"not null"`
	Deadline    time.Time     `json:"deadline" gorm:"not null"`
	Status      RequestStatus `json:"status" gorm:"default:'open'"`
	AcceptedBy  *uuid.UUID    `json:"accepted_by" gorm:"type:uuid"`
	CompletedAt *time.Time    `json:"completed_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// Relationships
	Requester auth.User        `json:"requester" gorm:"foreignKey:RequesterID"`
	Acceptor  *auth.User       `json:"acceptor" gorm:"foreignKey:AcceptedBy"`
	PlantType garden.PlantType `json:"plant_type" gorm:"foreignKey:PlantTypeID"`
}

func (dr *DeliveryRequest) BeforeCreate(tx *gorm.DB) error {
	if dr.ID == uuid.Nil {
		dr.ID = uuid.New()
	}
	return nil
}

// RequestStatus represents the status of a delivery request
type RequestStatus string

const (
	RequestStatusOpen      RequestStatus = "open"
	RequestStatusAccepted  RequestStatus = "accepted"
	RequestStatusCompleted RequestStatus = "completed"
	RequestStatusExpired   RequestStatus = "expired"
	RequestStatusCancelled RequestStatus = "cancelled"
)

// PlantListing represents a plant available for direct purchase
type PlantListing struct {
	ID        uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SellerID  uuid.UUID     `json:"seller_id" gorm:"type:uuid;not null"`
	PlantID   uuid.UUID     `json:"plant_id" gorm:"type:uuid;not null"`
	Price     int           `json:"price" gorm:"not null"`
	Status    ListingStatus `json:"status" gorm:"default:'active'"`
	ExpiresAt time.Time     `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`

	// Relationships
	Seller auth.User    `json:"seller" gorm:"foreignKey:SellerID"`
	Plant  garden.Plant `json:"plant" gorm:"foreignKey:PlantID"`
}

func (pl *PlantListing) BeforeCreate(tx *gorm.DB) error {
	if pl.ID == uuid.Nil {
		pl.ID = uuid.New()
	}
	if pl.ExpiresAt.IsZero() {
		pl.ExpiresAt = time.Now().AddDate(0, 0, 7) // 7 days default
	}
	return nil
}

// ListingStatus represents the status of a plant listing
type ListingStatus string

const (
	ListingStatusActive    ListingStatus = "active"
	ListingStatusSold      ListingStatus = "sold"
	ListingStatusExpired   ListingStatus = "expired"
	ListingStatusCancelled ListingStatus = "cancelled"
)

// Transaction represents a blockchain-style transaction
type Transaction struct {
	ID          uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Hash        string            `json:"hash" gorm:"uniqueIndex;not null"`
	SenderKey   string            `json:"sender_key" gorm:"not null"`
	ReceiverKey string            `json:"receiver_key" gorm:"not null"`
	PlantID     *uuid.UUID        `json:"plant_id" gorm:"type:uuid"`
	Amount      int               `json:"amount" gorm:"not null"`
	TxType      TransactionType   `json:"tx_type" gorm:"not null"`
	Nonce       int64             `json:"nonce" gorm:"not null"`
	Signature   string            `json:"signature" gorm:"not null"`
	Status      TransactionStatus `json:"status" gorm:"default:'pending'"`
	BlockHeight *int64            `json:"block_height"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	// Relationships
	Plant *garden.Plant `json:"plant" gorm:"foreignKey:PlantID"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Hash == "" {
		t.Hash = t.GenerateHash()
	}
	return nil
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypePurchase   TransactionType = "purchase"
	TransactionTypeReward     TransactionType = "reward"
	TransactionTypeCommission TransactionType = "commission"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusRejected  TransactionStatus = "rejected"
)

// Block represents a block in the blockchain
type Block struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Height      int64     `json:"height" gorm:"uniqueIndex;not null"`
	Hash        string    `json:"hash" gorm:"uniqueIndex;not null"`
	PrevHash    string    `json:"prev_hash"`
	MerkleRoot  string    `json:"merkle_root" gorm:"not null"`
	Timestamp   time.Time `json:"timestamp" gorm:"not null"`
	ValidatorID uuid.UUID `json:"validator_id" gorm:"type:uuid;not null"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Validator    auth.User     `json:"validator" gorm:"foreignKey:ValidatorID"`
	Transactions []Transaction `json:"transactions" gorm:"many2many:block_transactions;"`
}

func (b *Block) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Wallet represents a user's wallet with cryptographic keys
type Wallet struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	PublicKey    string    `json:"public_key" gorm:"not null"`
	EncryptedKey string    `json:"encrypted_key" gorm:"not null"` // AES-256-GCM encrypted private key
	Balance      int       `json:"balance" gorm:"default:100"`
	Reputation   float64   `json:"reputation" gorm:"default:0"`
	Nonce        int64     `json:"nonce" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	User auth.User `json:"user" gorm:"foreignKey:UserID"`
}

func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// ConsensusNode represents a node in the consensus network
type ConsensusNode struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NodeID    string    `json:"node_id" gorm:"uniqueIndex;not null"`
	PublicKey string    `json:"public_key" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cn *ConsensusNode) BeforeCreate(tx *gorm.DB) error {
	if cn.ID == uuid.Nil {
		cn.ID = uuid.New()
	}
	return nil
}

// GenerateHash generates a unique hash for the transaction
func (t *Transaction) GenerateHash() string {
	data := fmt.Sprintf("%s%s%s%d%s%d", t.SenderKey, t.ReceiverKey, t.PlantID, t.Amount, t.TxType, t.Nonce)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// SignTransaction signs a transaction with the given private key
func SignTransaction(privateKey *ecdsa.PrivateKey, transaction *Transaction) (string, error) {
	// Create the data to sign
	data := fmt.Sprintf("%s%s%s%d%s%d", transaction.SenderKey, transaction.ReceiverKey, transaction.PlantID, transaction.Amount, transaction.TxType, transaction.Nonce)

	// Hash the data
	hash := sha256.Sum256([]byte(data))

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", err
	}

	// Encode the signature
	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}

// VerifySignature verifies a transaction signature
func VerifySignature(publicKey *ecdsa.PublicKey, transaction *Transaction, signature string) bool {
	// Create the data that was signed
	data := fmt.Sprintf("%s%s%s%d%s%d", transaction.SenderKey, transaction.ReceiverKey, transaction.PlantID, transaction.Amount, transaction.TxType, transaction.Nonce)

	// Hash the data
	hash := sha256.Sum256([]byte(data))

	// Decode the signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	if len(sigBytes) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	// Verify the signature
	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// PublicKeyFromString converts a hex string to an ECDSA public key
func PublicKeyFromString(publicKeyStr string) (*ecdsa.PublicKey, error) {
	// This is a simplified implementation - in production you'd want proper key encoding
	// For now, we'll assume the public key is stored as a hex string of the coordinates
	keyBytes, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != 65 {
		return nil, fmt.Errorf("invalid public key length")
	}

	// Extract x and y coordinates
	x := new(big.Int).SetBytes(keyBytes[1:33])
	y := new(big.Int).SetBytes(keyBytes[33:])

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, nil
}

// PublicKeyToString converts an ECDSA public key to a hex string
func PublicKeyToString(publicKey *ecdsa.PublicKey) string {
	// This is a simplified implementation - in production you'd want proper key encoding
	// For now, we'll store as uncompressed format (0x04 + x + y)
	keyBytes := append([]byte{0x04}, publicKey.X.Bytes()...)
	keyBytes = append(keyBytes, publicKey.Y.Bytes()...)
	return hex.EncodeToString(keyBytes)
}
