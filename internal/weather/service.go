package weather

import (
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/pkg/utils"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCurrentWeather() (*Weather, error) {
	return s.repo.GetCurrentWeather()
}

// RefreshWeather returns the current weather, drawing a new one with
// generate when the last has expired.
func (s *Service) RefreshWeather(now time.Time, validFor time.Duration, generate func() Weather) (*Weather, error) {
	return s.repo.RefreshWeather(now, validFor, generate)
}

func (s *Service) GetWeatherHistory(limit int) ([]Weather, error) {
	return s.repo.GetWeatherHistory(limit)
}

func (s *Service) GetWeatherByID(id uuid.UUID) (*Weather, error) {
	return s.repo.GetWeatherByID(id)
}

func (s *Service) GetWeatherForecast() ([]WeatherForecast, error) {
	// Get current weather to base forecast on
	currentWeather, err := s.repo.GetCurrentWeather()
	if err != nil {
		// Use default weather if current weather not available
		currentWeather = &Weather{
			Condition:   WeatherCloudy,
			Temperature: 20.0,
			Humidity:    50,
		}
	}

	var forecasts []WeatherForecast

	// Generate 4 forecast periods (6 hours each)
	for i := 1; i <= 4; i++ {
		forecastTime := time.Now().Add(time.Duration(i*6) * time.Hour)

		// Simple forecast logic - vary conditions slightly
		condition := currentWeather.Condition
		temperature := currentWeather.Temperature
		humidity := currentWeather.Humidity

		// Add some variation based on time of day
		hour := forecastTime.Hour()
		switch {
		case hour >= 6 && hour <= 12: // Morning
			temperature += 2
			humidity = utils.Max(30, humidity-10)
		case hour >= 13 && hour <= 18: // Afternoon
			temperature += 5
			humidity = utils.Max(20, humidity-20)
		case hour >= 19 && hour <= 23: // Evening
			temperature -= 2
			humidity = utils.Min(80, humidity+10)
		default: // Night
			temperature -= 5
			humidity = utils.Min(90, humidity+20)
		}

		forecast := WeatherForecast{
			Condition:   condition,
			Temperature: temperature,
			Humidity:    humidity,
			Probability: 85, // 85% confidence
			ForecastFor: forecastTime,
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}
