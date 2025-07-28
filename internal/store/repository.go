package store

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPlantTypes() ([]garden.PlantType, error) {
	var plantTypes []garden.PlantType
	if err := r.db.DB.Find(&plantTypes).Error; err != nil {
		return nil, err
	}
	return plantTypes, nil
}

func (r *Repository) GetPlantTypeByID(id uuid.UUID) (*garden.PlantType, error) {
	var plantType garden.PlantType
	if err := r.db.DB.First(&plantType, id).Error; err != nil {
		return nil, err
	}
	return &plantType, nil
}

func (r *Repository) GetUserByID(id uuid.UUID) (*auth.User, error) {
	var user auth.User
	if err := r.db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserCoins(userID uuid.UUID, newCoins int) error {
	return r.db.DB.Model(&auth.User{}).Where("id = ?", userID).Update("coins", newCoins).Error
}

func (r *Repository) CreateSeedInventory(inventory *garden.SeedInventory) error {
	return r.db.DB.Create(inventory).Error
}

func (r *Repository) GetUserSeedInventory(userID uuid.UUID) ([]garden.SeedInventory, error) {
	var seedInventory []garden.SeedInventory
	if err := r.db.DB.Preload("PlantType").
		Where("user_id = ?", userID).
		Order("purchased_at DESC").
		Find(&seedInventory).Error; err != nil {
		return nil, err
	}
	fmt.Println(seedInventory)
	return seedInventory, nil
}

func (r *Repository) ExecuteInTransaction(fn func(*database.Database) error) error {
	tx := r.db.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txDB := &database.Database{DB: tx}
	if err := fn(txDB); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
