package store

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// ErrInsufficientCoins means the user cannot afford the purchase.
var ErrInsufficientCoins = errors.New("insufficient coins")

// BuySeeds charges the user and adds the seeds to their inventory in one
// transaction, returning the coins left. Checking the balance and charging it
// is a single conditional update, so concurrent purchases cannot overspend.
func (r *Repository) BuySeeds(userID uuid.UUID, plantType garden.PlantType, quantity int) (int, error) {
	totalCost := plantType.SeedPrice * quantity
	var coins int
	err := r.db.DB.Transaction(func(tx *gorm.DB) error {
		charge := tx.Model(&auth.User{}).
			Where("id = ? AND coins >= ?", userID, totalCost).
			Update("coins", gorm.Expr("coins - ?", totalCost))
		if charge.Error != nil {
			return charge.Error
		}
		if charge.RowsAffected == 0 {
			return ErrInsufficientCoins
		}

		seeds := garden.SeedInventory{
			UserID:      userID,
			PlantTypeID: plantType.ID,
			Quantity:    quantity,
			PurchasedAt: time.Now(),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "plant_type_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"quantity":     gorm.Expr("seed_inventories.quantity + EXCLUDED.quantity"),
				"purchased_at": gorm.Expr("EXCLUDED.purchased_at"),
			}),
		}).Create(&seeds).Error; err != nil {
			return err
		}

		return tx.Model(&auth.User{}).Select("coins").Where("id = ?", userID).Scan(&coins).Error
	})
	return coins, err
}

func (r *Repository) GetUserSeedInventory(userID uuid.UUID) ([]garden.SeedInventory, error) {
	var seedInventory []garden.SeedInventory
	if err := r.db.DB.Preload("PlantType").
		Where("user_id = ? AND quantity > 0", userID).
		Order("purchased_at DESC").
		Find(&seedInventory).Error; err != nil {
		return nil, err
	}
	return seedInventory, nil
}
