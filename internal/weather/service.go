package weather

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/websocket"
	"github.com/my-garden/api/pkg/utils"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo        *Repository
	redisClient *redis.Client
	wsHandler   *websocket.Handler
}

func NewService(repo *Repository, redisClient *redis.Client, wsHandler *websocket.Handler) *Service {
	return &Service{
		repo:        repo,
		redisClient: redisClient,
		wsHandler:   wsHandler,
	}
}

func (s *Service) GetCurrentWeather() (*Weather, error) {
	return s.repo.GetCurrentWeather()
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

func (s *Service) CacheCurrentWeather(ctx context.Context, weather *Weather, duration time.Duration) error {
	key := "weather:current"
	// Note: In a real implementation, you'd serialize the weather struct to JSON
	// For now, we'll just store the condition
	return s.redisClient.Set(ctx, key, weather.Condition, duration).Err()
}

func (s *Service) UpdateWeather(weather *Weather) error {
	if err := s.repo.UpdateWeather(weather); err != nil {
		return err
	}

	// Broadcast weather update to all connected clients
	s.wsHandler.BroadcastEvent(websocket.EventWeather, uuid.Nil, uuid.Nil, weather)
	return nil
}
