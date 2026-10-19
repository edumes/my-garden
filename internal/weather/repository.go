package weather

import (
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
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

func (r *Repository) CreateWeather(weather *Weather) error {
	return r.db.DB.Create(weather).Error
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
