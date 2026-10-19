package store

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/garden"
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

func (s *Service) GetStoreInventory() ([]garden.PlantType, error) {
	return s.repo.GetPlantTypes()
}

// ErrInvalidQuantity rejects purchases outside the allowed range.
var ErrInvalidQuantity = errors.New("quantity must be between 1 and 10")

func (s *Service) BuySeed(userID uuid.UUID, plantTypeID uuid.UUID, quantity int) (*BuySeedResponse, error) {
	if quantity < 1 || quantity > 10 {
		return nil, ErrInvalidQuantity
	}

	plantType, err := s.repo.GetPlantTypeByID(plantTypeID)
	if err != nil {
		return nil, err
	}

	totalCost := plantType.SeedPrice * quantity
	coins, err := s.repo.BuySeeds(userID, *plantType, quantity)
	switch {
	case errors.Is(err, ErrInsufficientCoins):
		return nil, fmt.Errorf("%w: these seeds cost %d coins", err, totalCost)
	case errors.Is(err, ErrLocked):
		return nil, fmt.Errorf("%w: %s unlocks at level %d", err, plantType.Name, plantType.UnlockLevel)
	}
	if err != nil {
		return nil, err
	}

	return &BuySeedResponse{
		PlantType: *plantType,
		Quantity:  quantity,
		TotalCost: totalCost,
		UserCoins: coins,
	}, nil
}

func (s *Service) GetUserSeedInventory(userID uuid.UUID) ([]garden.SeedInventory, error) {
	return s.repo.GetUserSeedInventory(userID)
}
