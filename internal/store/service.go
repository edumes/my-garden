package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/database"
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

func (s *Service) BuySeed(userID uuid.UUID, plantTypeID uuid.UUID, quantity int) (*BuySeedResponse, error) {
	if quantity < 1 || quantity > 10 {
		return nil, fmt.Errorf("quantity must be between 1 and 10")
	}

	plantType, err := s.repo.GetPlantTypeByID(plantTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plant type: %w", err)
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	totalCost := plantType.SeedPrice * quantity
	if user.Coins < totalCost {
		return nil, fmt.Errorf("insufficient coins: required %d, available %d", totalCost, user.Coins)
	}

	seedInventory := &garden.SeedInventory{
		UserID:      userID,
		PlantTypeID: plantTypeID,
		Quantity:    quantity,
		PurchasedAt: time.Now(),
	}

	err = s.repo.ExecuteInTransaction(func(db *database.Database) error {
		if err := s.repo.UpdateUserCoins(userID, user.Coins-totalCost); err != nil {
			return fmt.Errorf("failed to update user coins: %w", err)
		}

		if err := s.repo.CreateSeedInventory(seedInventory); err != nil {
			return fmt.Errorf("failed to create seed inventory: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	return &BuySeedResponse{
		PlantType: *plantType,
		Quantity:  quantity,
		TotalCost: totalCost,
		UserCoins: updatedUser.Coins,
	}, nil
}

func (s *Service) GetUserSeedInventory(userID uuid.UUID) ([]garden.SeedInventory, error) {
	return s.repo.GetUserSeedInventory(userID)
}
