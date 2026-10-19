package weather

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"gorm.io/gorm"
)

type Repository struct {
	db *database.Database
}

func NewRepository(db *database.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetCurrentWeather() (*Weather, error) {
	var weather Weather
	err := r.db.DB.Order("created_at DESC").First(&weather).Error
	if err != nil {
		return nil, err
	}
	return &weather, nil
}

func (r *Repository) GetWeatherHistory(limit int) ([]Weather, error) {
	var weatherHistory []Weather
	err := r.db.DB.Order("created_at DESC").Limit(limit).Find(&weatherHistory).Error
	if err != nil {
		return nil, err
	}
	return weatherHistory, nil
}

// refreshLock is the advisory lock key that makes weather refreshes take
// turns across instances.
const refreshLock = 0x7765617468 // "weath"

// RefreshWeather returns the weather valid at now, first creating one from
// generate, valid for validFor, when the last one has expired. Instances take
// turns under a transaction-level advisory lock, so only one of them creates
// each new weather, and a restart keeps weather that is still valid.
func (r *Repository) RefreshWeather(now time.Time, validFor time.Duration, generate func() Weather) (*Weather, error) {
	var current Weather
	err := r.db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", refreshLock).Error; err != nil {
			return err
		}
		err := tx.Where("valid_until > ?", now).Order("created_at DESC").Take(&current).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		current = generate()
		current.ValidUntil = now.Add(validFor)
		return tx.Create(&current).Error
	})
	return &current, err
}

func (r *Repository) GetWeatherByID(id uuid.UUID) (*Weather, error) {
	var weather Weather
	err := r.db.DB.First(&weather, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &weather, nil
}

func (r *Repository) UpdateWeather(weather *Weather) error {
	return r.db.DB.Save(weather).Error
}

func (r *Repository) DeleteWeather(id uuid.UUID) error {
	return r.db.DB.Delete(&Weather{}, "id = ?", id).Error
}
