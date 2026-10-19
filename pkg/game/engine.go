package game

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/weather"
	"github.com/redis/go-redis/v9"
)

type GameEngine struct {
	db            *database.Database
	redis         *redis.Client
	config        *config.Config
	ctx           context.Context
	cancel        context.CancelFunc
	tickTicker    *time.Ticker
	weatherTicker *time.Ticker
	weatherRepo   *weather.Repository
	weatherSvc    *weather.Service
}

func NewGameEngine(db *database.Database, redis *redis.Client, cfg *config.Config) *GameEngine {
	ctx, cancel := context.WithCancel(context.Background())
	weatherRepo := weather.NewRepository(db)
	weatherSvc := weather.NewService(weatherRepo, redis)

	return &GameEngine{
		db:            db,
		redis:         redis,
		config:        cfg,
		ctx:           ctx,
		cancel:        cancel,
		tickTicker:    time.NewTicker(cfg.Game.TickInterval),
		weatherTicker: time.NewTicker(cfg.Game.WeatherUpdateInterval),
		weatherRepo:   weatherRepo,
		weatherSvc:    weatherSvc,
	}
}

func (g *GameEngine) Start() {
	log.Println("Starting game engine...")

	// Start game tick loop
	go g.gameTickLoop()

	// Start weather update loop
	go g.weatherUpdateLoop()

	// Initialize current weather
	g.updateWeather()
}

func (g *GameEngine) Stop() {
	log.Println("Stopping game engine...")
	g.cancel()
	g.tickTicker.Stop()
	g.weatherTicker.Stop()
}

func (g *GameEngine) gameTickLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-g.tickTicker.C:
			g.processGameTick()
		}
	}
}

func (g *GameEngine) weatherUpdateLoop() {
	for {
		select {
		case <-g.ctx.Done():
			return
		case <-g.weatherTicker.C:
			g.updateWeather()
		}
	}
}

func (g *GameEngine) processGameTick() {
	// Without current weather, plants grow at the normal rate.
	multiplier := 1.0
	if current, err := g.weatherSvc.GetCurrentWeather(); err == nil {
		multiplier = current.GrowthMultiplier
	} else {
		log.Printf("Growing plants without weather effects: %v", err)
	}

	if err := garden.GrowPlants(g.db.DB, time.Now(), multiplier, g.config.Game.Speed); err != nil {
		log.Printf("Failed to grow plants: %v", err)
	}
}

func (g *GameEngine) updateWeather() {
	log.Println("Updating weather...")

	// Generate new weather conditions
	weather := g.generateWeather()

	// Save to database
	if err := g.weatherRepo.CreateWeather(&weather); err != nil {
		log.Printf("Failed to save weather: %v", err)
		return
	}

	// Cache current weather in Redis
	if err := g.weatherSvc.CacheCurrentWeather(g.ctx, &weather, g.config.Game.WeatherUpdateInterval); err != nil {
		log.Printf("Failed to cache weather: %v", err)
	}

	log.Printf("Weather updated: %s, Temperature: %.1f°C, Growth Multiplier: %.2f",
		weather.Condition, weather.Temperature, weather.GrowthMultiplier)
}

func (g *GameEngine) generateWeather() weather.Weather {
	// Get current season
	season := weather.GetSeason(time.Now())

	// Define weather probabilities based on season
	weatherConditions := g.getWeatherConditionsForSeason(season)

	// Select random weather condition
	selectedCondition := weatherConditions[rand.Intn(len(weatherConditions))]

	// Generate temperature based on season and weather
	temperature := g.generateTemperature(season, selectedCondition)

	// Generate humidity
	humidity := rand.Intn(40) + 30 // 30-70%

	// Generate wind speed
	windSpeed := rand.Float64() * 20 // 0-20 km/h

	// Get weather effects
	growthMultiplier, waterEvaporationRate := weather.GetWeatherEffects(selectedCondition)

	weather := weather.Weather{
		Condition:            selectedCondition,
		Temperature:          temperature,
		Humidity:             humidity,
		WindSpeed:            windSpeed,
		Pressure:             1013.25, // Standard atmospheric pressure
		GrowthMultiplier:     growthMultiplier,
		WaterEvaporationRate: waterEvaporationRate,
		ValidUntil:           time.Now().Add(g.config.Game.WeatherUpdateInterval),
	}

	return weather
}

func (g *GameEngine) getWeatherConditionsForSeason(season weather.Season) []weather.WeatherCondition {
	switch season {
	case weather.SeasonSpring:
		return []weather.WeatherCondition{
			weather.WeatherSunny, weather.WeatherCloudy, weather.WeatherRainy,
			weather.WeatherFoggy, weather.WeatherWindy,
		}
	case weather.SeasonSummer:
		return []weather.WeatherCondition{
			weather.WeatherSunny, weather.WeatherCloudy, weather.WeatherStormy,
			weather.WeatherWindy,
		}
	case weather.SeasonAutumn:
		return []weather.WeatherCondition{
			weather.WeatherCloudy, weather.WeatherRainy, weather.WeatherFoggy,
			weather.WeatherWindy, weather.WeatherSunny,
		}
	case weather.SeasonWinter:
		return []weather.WeatherCondition{
			weather.WeatherCloudy, weather.WeatherSnowy, weather.WeatherFoggy,
			weather.WeatherWindy,
		}
	default:
		return []weather.WeatherCondition{
			weather.WeatherSunny, weather.WeatherCloudy, weather.WeatherRainy,
		}
	}
}

func (g *GameEngine) generateTemperature(season weather.Season, condition weather.WeatherCondition) float64 {
	baseTemp := g.getBaseTemperatureForSeason(season)

	// Adjust temperature based on weather condition
	switch condition {
	case weather.WeatherSunny:
		baseTemp += rand.Float64()*5 + 2 // +2 to +7°C
	case weather.WeatherCloudy:
		baseTemp += rand.Float64()*3 - 1 // -1 to +2°C
	case weather.WeatherRainy:
		baseTemp += rand.Float64()*2 - 2 // -2 to 0°C
	case weather.WeatherStormy:
		baseTemp += rand.Float64()*3 - 3 // -3 to 0°C
	case weather.WeatherFoggy:
		baseTemp += rand.Float64()*2 - 1 // -1 to +1°C
	case weather.WeatherWindy:
		baseTemp += rand.Float64()*2 - 1 // -1 to +1°C
	case weather.WeatherSnowy:
		baseTemp += rand.Float64()*3 - 5 // -5 to -2°C
	}

	return baseTemp
}

func (g *GameEngine) getBaseTemperatureForSeason(season weather.Season) float64 {
	switch season {
	case weather.SeasonSpring:
		return 15.0
	case weather.SeasonSummer:
		return 25.0
	case weather.SeasonAutumn:
		return 15.0
	case weather.SeasonWinter:
		return 5.0
	default:
		return 15.0
	}
}

func (g *GameEngine) GetCurrentWeather() (*weather.Weather, error) {
	return g.weatherSvc.GetCurrentWeather()
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
