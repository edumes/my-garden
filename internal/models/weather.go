package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Weather represents current weather conditions
type Weather struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Condition   WeatherCondition `json:"condition" gorm:"not null"`
	Temperature float64          `json:"temperature" gorm:"not null"`     // in Celsius
	Humidity    int              `json:"humidity" gorm:"not null"`        // 0-100
	WindSpeed   float64          `json:"wind_speed" gorm:"default:0"`     // km/h
	Pressure    float64          `json:"pressure" gorm:"default:1013.25"` // hPa

	// Effects on plants
	GrowthMultiplier     float64 `json:"growth_multiplier" gorm:"default:1.0"`
	WaterEvaporationRate float64 `json:"water_evaporation_rate" gorm:"default:1.0"`

	// Timestamps
	CreatedAt  time.Time `json:"created_at"`
	ValidUntil time.Time `json:"valid_until"`
}

func (w *Weather) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// WeatherCondition represents different weather types
type WeatherCondition string

const (
	WeatherSunny  WeatherCondition = "sunny"
	WeatherCloudy WeatherCondition = "cloudy"
	WeatherRainy  WeatherCondition = "rainy"
	WeatherStormy WeatherCondition = "stormy"
	WeatherFoggy  WeatherCondition = "foggy"
	WeatherWindy  WeatherCondition = "windy"
	WeatherSnowy  WeatherCondition = "snowy"
)

// WeatherForecast represents weather predictions
type WeatherForecast struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Condition   WeatherCondition `json:"condition" gorm:"not null"`
	Temperature float64          `json:"temperature" gorm:"not null"`
	Humidity    int              `json:"humidity" gorm:"not null"`
	Probability int              `json:"probability" gorm:"not null"` // 0-100

	// Timestamps
	ForecastFor time.Time `json:"forecast_for"`
	CreatedAt   time.Time `json:"created_at"`
}

func (wf *WeatherForecast) BeforeCreate(tx *gorm.DB) error {
	if wf.ID == uuid.Nil {
		wf.ID = uuid.New()
	}
	return nil
}

// Season represents the current season
type Season string

const (
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonAutumn Season = "autumn"
	SeasonWinter Season = "winter"
)

// GetSeason returns the current season based on date
func GetSeason(date time.Time) Season {
	month := date.Month()
	switch {
	case month >= 3 && month <= 5:
		return SeasonSpring
	case month >= 6 && month <= 8:
		return SeasonSummer
	case month >= 9 && month <= 11:
		return SeasonAutumn
	default:
		return SeasonWinter
	}
}

// GetWeatherEffects returns the effects of weather on plant growth
func GetWeatherEffects(condition WeatherCondition) (growthMultiplier, waterEvaporationRate float64) {
	switch condition {
	case WeatherSunny:
		return 1.2, 1.5 // Faster growth, more water evaporation
	case WeatherCloudy:
		return 1.0, 1.0 // Normal conditions
	case WeatherRainy:
		return 1.1, 0.3 // Slight growth boost, less water evaporation
	case WeatherStormy:
		return 0.8, 0.1 // Slower growth, minimal water evaporation
	case WeatherFoggy:
		return 0.9, 0.8 // Slightly slower growth
	case WeatherWindy:
		return 0.95, 1.3 // Slightly slower growth, more water evaporation
	case WeatherSnowy:
		return 0.5, 0.2 // Much slower growth, minimal water evaporation
	default:
		return 1.0, 1.0
	}
}
