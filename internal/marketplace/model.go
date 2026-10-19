package marketplace

import (
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
)

// DeliveryRequest asks other players to deliver a plant type for a coin reward.
type DeliveryRequest struct {
	ID          uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RequesterID uuid.UUID     `json:"requester_id" gorm:"type:uuid;not null"`
	PlantTypeID uuid.UUID     `json:"plant_type_id" gorm:"type:uuid;not null"`
	Reward      int           `json:"reward" gorm:"not null;check:chk_delivery_requests_reward,reward > 0"`
	Deadline    time.Time     `json:"deadline" gorm:"not null"`
	Status      RequestStatus `json:"status" gorm:"not null;default:'open'"`
	AcceptedBy  *uuid.UUID    `json:"accepted_by" gorm:"type:uuid"`
	CompletedAt *time.Time    `json:"completed_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	// Relationships
	Requester auth.PublicUser  `json:"requester" gorm:"foreignKey:RequesterID"`
	Acceptor  *auth.PublicUser `json:"acceptor" gorm:"foreignKey:AcceptedBy"`
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
)

// PlantListing offers a harvestable plant for a fixed price. A plant can have
// at most one active listing.
type PlantListing struct {
	ID        uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SellerID  uuid.UUID     `json:"seller_id" gorm:"type:uuid;not null"`
	PlantID   uuid.UUID     `json:"plant_id" gorm:"type:uuid;not null;uniqueIndex:idx_plant_listings_active_plant,where:status = 'active'"`
	Price     int           `json:"price" gorm:"not null;check:chk_plant_listings_price,price > 0"`
	Status    ListingStatus `json:"status" gorm:"not null;default:'active'"`
	BuyerID   *uuid.UUID    `json:"buyer_id" gorm:"type:uuid"`
	ExpiresAt time.Time     `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`

	// Relationships
	Seller auth.PublicUser `json:"seller" gorm:"foreignKey:SellerID"`
	Plant  garden.Plant    `json:"plant" gorm:"foreignKey:PlantID"`
}

func (pl *PlantListing) BeforeCreate(tx *gorm.DB) error {
	if pl.ID == uuid.Nil {
		pl.ID = uuid.New()
	}
	return nil
}

// ListingStatus represents the status of a plant listing
type ListingStatus string

const (
	ListingStatusActive  ListingStatus = "active"
	ListingStatusSold    ListingStatus = "sold"
	ListingStatusExpired ListingStatus = "expired"
)
