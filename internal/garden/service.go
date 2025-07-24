package garden

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGardenNotFound = errors.New("garden not found")
	ErrPlantNotFound  = errors.New("plant not found")
)

type Service interface {
	CreateGarden(userID uuid.UUID, req CreateGardenRequest) (*Garden, error)
	GetGarden(id uuid.UUID) (*Garden, error)
	GetUserGardens(userID uuid.UUID) ([]Garden, error)
	PlantSeed(gardenID uuid.UUID, req PlantRequest) (*Plant, error)
	HarvestPlant(plantID uuid.UUID) (*Plant, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateGarden(userID uuid.UUID, req CreateGardenRequest) (*Garden, error) {
	garden := &Garden{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateGarden(garden); err != nil {
		return nil, err
	}

	return garden, nil
}

func (s *service) GetGarden(id uuid.UUID) (*Garden, error) {
	garden, err := s.repo.GetGarden(id)
	if err != nil {
		return nil, ErrGardenNotFound
	}
	return garden, nil
}

func (s *service) GetUserGardens(userID uuid.UUID) ([]Garden, error) {
	return s.repo.GetUserGardens(userID)
}

func (s *service) PlantSeed(gardenID uuid.UUID, req PlantRequest) (*Plant, error) {
	garden, err := s.repo.GetGarden(gardenID)
	if err != nil {
		return nil, ErrGardenNotFound
	}

	plant := &Plant{
		GardenID:    garden.ID,
		PlantTypeID: req.PlantTypeID,
		Position:    *req.Position,
		Stage:       PlantStageSeed,
		PlantedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.PlantSeed(plant); err != nil {
		return nil, err
	}

	return plant, nil
}

func (s *service) HarvestPlant(plantID uuid.UUID) (*Plant, error) {
	plant, err := s.repo.GetPlant(plantID)
	if err != nil {
		return nil, ErrPlantNotFound
	}

	if plant.Stage != PlantStageMature {
		return nil, errors.New("plant is not ready for harvest")
	}

	now := time.Now()
	plant.HarvestedAt = &now
	plant.UpdatedAt = now

	if err := s.repo.HarvestPlant(plant); err != nil {
		return nil, err
	}

	return plant, nil
}
