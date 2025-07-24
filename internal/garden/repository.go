package garden

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateGarden(garden *Garden) error
	GetGarden(id uuid.UUID) (*Garden, error)
	GetUserGardens(userID uuid.UUID) ([]Garden, error)
	PlantSeed(plant *Plant) error
	GetPlant(id uuid.UUID) (*Plant, error)
	HarvestPlant(plant *Plant) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateGarden(garden *Garden) error {
	return r.db.Create(garden).Error
}

func (r *repository) GetGarden(id uuid.UUID) (*Garden, error) {
	var garden Garden
	if err := r.db.Preload("Plants").First(&garden, id).Error; err != nil {
		return nil, err
	}
	return &garden, nil
}

func (r *repository) GetUserGardens(userID uuid.UUID) ([]Garden, error) {
	var gardens []Garden
	if err := r.db.Where("user_id = ?", userID).Preload("Plants").Find(&gardens).Error; err != nil {
		return nil, err
	}
	return gardens, nil
}

func (r *repository) PlantSeed(plant *Plant) error {
	return r.db.Create(plant).Error
}

func (r *repository) GetPlant(id uuid.UUID) (*Plant, error) {
	var plant Plant
	if err := r.db.First(&plant, id).Error; err != nil {
		return nil, err
	}
	return &plant, nil
}

func (r *repository) HarvestPlant(plant *Plant) error {
	return r.db.Save(plant).Error
}
