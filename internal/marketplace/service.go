package marketplace

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Marketplace rules.
const (
	minReward          = 10
	minDeadlineLead    = time.Hour
	listingLifetime    = 7 * 24 * time.Hour
	commissionPerMille = 5 // 0.5% of each sale, withheld from the seller
	latePenaltyDivisor = 5 // late deliveries lose 20% of the reward
)

// Errors the handler maps to client responses.
var (
	ErrInvalidReward     = errors.New("reward must be at least 10 coins")
	ErrInvalidDeadline   = errors.New("deadline must be at least 1 hour in the future")
	ErrInvalidPrice      = errors.New("price must be at least 1 coin")
	ErrNotFound          = errors.New("not found")
	ErrOwnItem           = errors.New("you cannot accept or buy your own request or listing")
	ErrNotAvailable      = errors.New("no longer available")
	ErrAlreadyListed     = errors.New("this plant is already listed")
	ErrNotYourPlant      = errors.New("you can only list plants you grow")
	ErrNotHarvestable    = errors.New("only harvestable plants can be listed")
	ErrInsufficientCoins = errors.New("insufficient coins")
	ErrRequesterBroke    = errors.New("the requester can no longer pay the reward")
	ErrNoFreePlot        = errors.New("you need a free plot in one of your gardens to receive this plant")
)

// Service runs marketplace commands. Every command that moves coins or plants
// is one transaction, and claims its request or listing with a conditional
// update first, so concurrent or repeated commands cannot apply twice.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// chargeCoins deducts amount from the user only if they can afford it.
func chargeCoins(tx *gorm.DB, userID uuid.UUID, amount int) (bool, error) {
	res := tx.Model(&auth.User{}).
		Where("id = ? AND coins >= ?", userID, amount).
		Update("coins", gorm.Expr("coins - ?", amount))
	return res.RowsAffected == 1, res.Error
}

func payCoins(tx *gorm.DB, userID uuid.UUID, amount int) error {
	return tx.Model(&auth.User{}).Where("id = ?", userID).
		Update("coins", gorm.Expr("coins + ?", amount)).Error
}

func (s *Service) CreateDeliveryRequest(userID, plantTypeID uuid.UUID, reward int, deadline time.Time) (*DeliveryRequest, error) {
	if reward < minReward {
		return nil, ErrInvalidReward
	}
	if deadline.Before(time.Now().Add(minDeadlineLead)) {
		return nil, ErrInvalidDeadline
	}

	var plantType garden.PlantType
	if err := s.db.Take(&plantType, "id = ?", plantTypeID).Error; err != nil {
		return nil, notFoundOr(err)
	}

	// The reward is charged on completion; refusing requests the requester
	// cannot pay today avoids promising rewards that do not exist.
	var requester auth.User
	if err := s.db.Take(&requester, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	if requester.Coins < reward {
		return nil, ErrInsufficientCoins
	}

	request := &DeliveryRequest{
		RequesterID: userID,
		PlantTypeID: plantTypeID,
		Reward:      reward,
		Deadline:    deadline,
		Status:      RequestStatusOpen,
	}
	if err := s.db.Create(request).Error; err != nil {
		return nil, err
	}
	request.PlantType = plantType
	return request, nil
}

func (s *Service) GetOpenDeliveryRequests() ([]DeliveryRequest, error) {
	var requests []DeliveryRequest
	err := s.db.Preload("Requester").Preload("PlantType").
		Where("status = ? AND deadline > ?", RequestStatusOpen, time.Now()).
		Order("created_at DESC").
		Find(&requests).Error
	return requests, err
}

func (s *Service) AcceptDeliveryRequest(requestID, acceptorID uuid.UUID) error {
	res := s.db.Model(&DeliveryRequest{}).
		Where("id = ? AND status = ? AND deadline > ? AND requester_id <> ?",
			requestID, RequestStatusOpen, time.Now(), acceptorID).
		Updates(map[string]any{"status": RequestStatusAccepted, "accepted_by": acceptorID})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 1 {
		return nil
	}
	return s.explainRequest(requestID, func(r DeliveryRequest) bool { return r.RequesterID == acceptorID })
}

// CompleteDeliveryRequest pays the acceptor and returns the reward paid,
// which drops by 20% after the deadline.
func (s *Service) CompleteDeliveryRequest(requestID, acceptorID uuid.UUID) (int, error) {
	var paid int
	err := s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var request DeliveryRequest
		if err := tx.Raw(`UPDATE delivery_requests SET status = ?, completed_at = ?, updated_at = ?
			WHERE id = ? AND status = ? AND accepted_by = ?
			RETURNING *`,
			RequestStatusCompleted, now, now, requestID, RequestStatusAccepted, acceptorID,
		).Scan(&request).Error; err != nil {
			return err
		}
		if request.ID == uuid.Nil {
			return s.explainRequest(requestID, func(DeliveryRequest) bool { return false })
		}

		paid = request.Reward
		if now.After(request.Deadline) {
			paid -= paid / latePenaltyDivisor
		}
		charged, err := chargeCoins(tx, request.RequesterID, paid)
		if err != nil {
			return err
		}
		if !charged {
			return ErrRequesterBroke
		}
		if err := payCoins(tx, acceptorID, paid); err != nil {
			return err
		}
		_, _, err = auth.GrantExperience(tx, acceptorID, auth.TradeExperience)
		return err
	})
	return paid, err
}

// explainRequest tells why a delivery request command matched nothing.
func (s *Service) explainRequest(requestID uuid.UUID, own func(DeliveryRequest) bool) error {
	var request DeliveryRequest
	if err := s.db.Take(&request, "id = ?", requestID).Error; err != nil {
		return notFoundOr(err)
	}
	if own(request) {
		return ErrOwnItem
	}
	return ErrNotAvailable
}

func (s *Service) CreatePlantListing(sellerID, plantID uuid.UUID, price int) (*PlantListing, error) {
	if price < 1 {
		return nil, ErrInvalidPrice
	}

	var plant garden.Plant
	if err := s.db.Take(&plant, "id = ?", plantID).Error; err != nil {
		return nil, notFoundOr(err)
	}
	if plant.GrowerID != sellerID {
		return nil, ErrNotYourPlant
	}
	if plant.Stage != garden.PlantStageHarvestable {
		return nil, ErrNotHarvestable
	}

	listing := &PlantListing{
		SellerID:  sellerID,
		PlantID:   plantID,
		Price:     price,
		Status:    ListingStatusActive,
		ExpiresAt: time.Now().Add(listingLifetime),
	}
	// An expired listing gives up the plant's one active listing.
	if err := s.db.Model(&PlantListing{}).
		Where("plant_id = ? AND status = ? AND expires_at <= ?", plantID, ListingStatusActive, time.Now()).
		Update("status", ListingStatusExpired).Error; err != nil {
		return nil, err
	}
	// The partial unique index allows one active listing per plant.
	res := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(listing)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrAlreadyListed
	}
	listing.Plant = plant
	return listing, nil
}

// ListingFilters narrows GetActiveListings; nil fields do not filter.
type ListingFilters struct {
	PlantTypeID *uuid.UUID
	MinPrice    *int
	MaxPrice    *int
}

func (s *Service) GetActiveListings(filters ListingFilters) ([]PlantListing, error) {
	query := s.db.Preload("Seller").Preload("Plant.PlantType").
		Where("plant_listings.status = ? AND plant_listings.expires_at > ?", ListingStatusActive, time.Now())
	if filters.PlantTypeID != nil {
		query = query.Joins("JOIN plants ON plants.id = plant_listings.plant_id").
			Where("plants.plant_type_id = ?", *filters.PlantTypeID)
	}
	if filters.MinPrice != nil {
		query = query.Where("plant_listings.price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("plant_listings.price <= ?", *filters.MaxPrice)
	}

	var listings []PlantListing
	err := query.Order("plant_listings.created_at DESC").Find(&listings).Error
	return listings, err
}

// PurchasePlantListing charges the buyer, pays the seller minus commission
// plus trade experience, and moves the plant into the first free plot of the
// buyer's gardens, making the buyer its grower.
func (s *Service) PurchasePlantListing(listingID, buyerID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var listing PlantListing
		if err := tx.Raw(`UPDATE plant_listings SET status = ?, buyer_id = ?, updated_at = ?
			WHERE id = ? AND status = ? AND expires_at > ? AND seller_id <> ?
			RETURNING *`,
			ListingStatusSold, buyerID, time.Now(), listingID, ListingStatusActive, time.Now(), buyerID,
		).Scan(&listing).Error; err != nil {
			return err
		}
		if listing.ID == uuid.Nil {
			return s.explainListing(listingID, buyerID)
		}

		// Charging locks the buyer's row, so the buyer's concurrent purchases
		// pick free cells one at a time.
		charged, err := chargeCoins(tx, buyerID, listing.Price)
		if err != nil {
			return err
		}
		if !charged {
			return ErrInsufficientCoins
		}
		if err := payCoins(tx, listing.SellerID, listing.Price-listing.Price*commissionPerMille/1000); err != nil {
			return err
		}
		if _, _, err := auth.GrantExperience(tx, listing.SellerID, auth.TradeExperience); err != nil {
			return err
		}

		gardenID, position, err := freeCell(tx, buyerID)
		if err != nil {
			return err
		}
		moved := tx.Model(&garden.Plant{}).
			Where("id = ? AND grower_id = ?", listing.PlantID, listing.SellerID).
			Updates(map[string]any{"garden_id": gardenID, "position": position, "grower_id": buyerID})
		if moved.Error != nil {
			return moved.Error
		}
		if moved.RowsAffected == 0 {
			// The seller removed or gave away the plant after listing it.
			return ErrNotAvailable
		}
		return nil
	})
}

// freeCell returns the first empty plot across the user's gardens.
func freeCell(tx *gorm.DB, userID uuid.UUID) (uuid.UUID, int, error) {
	var cell struct {
		GardenID uuid.UUID
		Position int
	}
	err := tx.Raw(`SELECT g.id AS garden_id, cell.position
		FROM gardens g CROSS JOIN generate_series(0, 8) AS cell(position)
		WHERE g.user_id = ?
		AND NOT EXISTS (SELECT 1 FROM plants p WHERE p.garden_id = g.id AND p.position = cell.position)
		ORDER BY g.created_at, cell.position
		LIMIT 1`, userID).Scan(&cell).Error
	if err == nil && cell.GardenID == uuid.Nil {
		err = ErrNoFreePlot
	}
	return cell.GardenID, cell.Position, err
}

// explainListing tells why a purchase matched no active listing.
func (s *Service) explainListing(listingID, buyerID uuid.UUID) error {
	var listing PlantListing
	if err := s.db.Take(&listing, "id = ?", listingID).Error; err != nil {
		return notFoundOr(err)
	}
	if listing.SellerID == buyerID {
		return ErrOwnItem
	}
	return ErrNotAvailable
}

func (s *Service) GetMarketplaceStats() (*MarketplaceStatsResponse, error) {
	var stats MarketplaceStatsResponse
	now := time.Now()
	err := s.db.Raw(`SELECT
		(SELECT COUNT(*) FROM delivery_requests WHERE status = ? AND deadline > ?) AS active_delivery_requests,
		(SELECT COUNT(*) FROM delivery_requests) AS total_delivery_requests,
		(SELECT COUNT(*) FROM plant_listings WHERE status = ? AND expires_at > ?) AS active_listings,
		(SELECT COUNT(*) FROM plant_listings) AS total_listings,
		(SELECT COALESCE(SUM(price), 0) FROM plant_listings WHERE status = ?) AS total_volume,
		(SELECT COUNT(*) FROM plant_listings WHERE status = ?) +
		(SELECT COUNT(*) FROM delivery_requests WHERE status = ?) AS total_transactions`,
		RequestStatusOpen, now, ListingStatusActive, now, ListingStatusSold, ListingStatusSold, RequestStatusCompleted,
	).Scan(&stats).Error
	return &stats, err
}

func notFoundOr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
