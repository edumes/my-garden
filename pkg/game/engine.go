package game

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
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
}

func NewGameEngine(db *database.Database, redis *redis.Client, cfg *config.Config) *GameEngine {
	ctx, cancel := context.WithCancel(context.Background())

	return &GameEngine{
		db:            db,
		redis:         redis,
		config:        cfg,
		ctx:           ctx,
		cancel:        cancel,
		tickTicker:    time.NewTicker(cfg.Game.TickInterval),
		weatherTicker: time.NewTicker(cfg.Game.WeatherUpdateInterval),
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
	log.Println("Processing game tick...")

	// Get current weather
	var currentWeather models.Weather
	if err := g.db.DB.Order("created_at DESC").First(&currentWeather).Error; err != nil {
		log.Printf("Failed to get current weather: %v", err)
		return
	}

	// Process all plants
	var plants []models.Plant
	if err := g.db.DB.Preload("PlantType").Preload("Garden").Find(&plants).Error; err != nil {
		log.Printf("Failed to fetch plants: %v", err)
		return
	}

	for _, plant := range plants {
		g.processPlantGrowth(&plant, &currentWeather)
	}
}

func (g *GameEngine) processPlantGrowth(plant *models.Plant, weather *models.Weather) {
	// Skip if plant is already harvested or withered
	if plant.Stage == models.PlantStageHarvestable || plant.Stage == models.PlantStageWithered {
		return
	}

	// Calculate growth progress
	baseGrowthRate := 1.0 / float64(plant.PlantType.GrowthTime) // Growth per minute
	weatherMultiplier := weather.GrowthMultiplier

	// Apply water and fertilizer bonuses
	waterBonus := 1.0
	if plant.WaterLevel > 70 {
		waterBonus = 1.2
	} else if plant.WaterLevel < 30 {
		waterBonus = 0.8
	}

	// Calculate total growth for this tick
	tickDuration := g.config.Game.TickInterval.Minutes()
	growthIncrement := baseGrowthRate * weatherMultiplier * waterBonus * tickDuration

	plant.GrowthProgress += growthIncrement

	// Update plant stage based on growth progress
	g.updatePlantStage(plant)

	// Reduce water level due to evaporation
	evaporationRate := weather.WaterEvaporationRate * tickDuration / 60.0 // per minute
	plant.WaterLevel = max(0, plant.WaterLevel-int(evaporationRate*10))

	// Update plant health based on water level
	if plant.WaterLevel < 20 {
		plant.Health = max(0, plant.Health-5)
	} else if plant.WaterLevel > 80 {
		plant.Health = min(100, plant.Health+2)
	}

	// Check if plant has withered
	if plant.Health <= 0 {
		plant.Stage = models.PlantStageWithered
	}

	// Save plant changes
	if err := g.db.DB.Save(plant).Error; err != nil {
		log.Printf("Failed to save plant %s: %v", plant.ID, err)
	}
}

func (g *GameEngine) updatePlantStage(plant *models.Plant) {
	progress := plant.GrowthProgress

	switch {
	case progress < 20:
		plant.Stage = models.PlantStageSeed
	case progress < 40:
		plant.Stage = models.PlantStageSprout
	case progress < 70:
		plant.Stage = models.PlantStageGrowing
	case progress < 100:
		plant.Stage = models.PlantStageMature
	default:
		plant.Stage = models.PlantStageHarvestable
	}
}

func (g *GameEngine) updateWeather() {
	log.Println("Updating weather...")

	// Generate new weather conditions
	weather := g.generateWeather()

	// Save to database
	if err := g.db.DB.Create(&weather).Error; err != nil {
		log.Printf("Failed to save weather: %v", err)
		return
	}

	// Cache current weather in Redis
	g.cacheCurrentWeather(weather)

	log.Printf("Weather updated: %s, Temperature: %.1f°C, Growth Multiplier: %.2f",
		weather.Condition, weather.Temperature, weather.GrowthMultiplier)
}

func (g *GameEngine) generateWeather() models.Weather {
	// Get current season
	season := models.GetSeason(time.Now())

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
	growthMultiplier, waterEvaporationRate := models.GetWeatherEffects(selectedCondition)

	weather := models.Weather{
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

func (g *GameEngine) getWeatherConditionsForSeason(season models.Season) []models.WeatherCondition {
	switch season {
	case models.SeasonSpring:
		return []models.WeatherCondition{
			models.WeatherSunny, models.WeatherCloudy, models.WeatherRainy,
			models.WeatherFoggy, models.WeatherWindy,
		}
	case models.SeasonSummer:
		return []models.WeatherCondition{
			models.WeatherSunny, models.WeatherCloudy, models.WeatherStormy,
			models.WeatherWindy,
		}
	case models.SeasonAutumn:
		return []models.WeatherCondition{
			models.WeatherCloudy, models.WeatherRainy, models.WeatherFoggy,
			models.WeatherWindy, models.WeatherSunny,
		}
	case models.SeasonWinter:
		return []models.WeatherCondition{
			models.WeatherCloudy, models.WeatherSnowy, models.WeatherFoggy,
			models.WeatherWindy,
		}
	default:
		return []models.WeatherCondition{
			models.WeatherSunny, models.WeatherCloudy, models.WeatherRainy,
		}
	}
}

func (g *GameEngine) generateTemperature(season models.Season, condition models.WeatherCondition) float64 {
	baseTemp := g.getBaseTemperatureForSeason(season)

	// Adjust temperature based on weather condition
	switch condition {
	case models.WeatherSunny:
		baseTemp += rand.Float64()*5 + 2 // +2 to +7°C
	case models.WeatherCloudy:
		baseTemp += rand.Float64()*3 - 1 // -1 to +2°C
	case models.WeatherRainy:
		baseTemp += rand.Float64()*2 - 2 // -2 to 0°C
	case models.WeatherStormy:
		baseTemp += rand.Float64()*3 - 3 // -3 to 0°C
	case models.WeatherFoggy:
		baseTemp += rand.Float64()*2 - 1 // -1 to +1°C
	case models.WeatherWindy:
		baseTemp += rand.Float64()*2 - 1 // -1 to +1°C
	case models.WeatherSnowy:
		baseTemp += rand.Float64()*3 - 5 // -5 to -2°C
	}

	return baseTemp
}

func (g *GameEngine) getBaseTemperatureForSeason(season models.Season) float64 {
	switch season {
	case models.SeasonSpring:
		return 15.0
	case models.SeasonSummer:
		return 25.0
	case models.SeasonAutumn:
		return 15.0
	case models.SeasonWinter:
		return 5.0
	default:
		return 15.0
	}
}

func (g *GameEngine) cacheCurrentWeather(weather models.Weather) {
	// Cache weather in Redis for quick access
	key := "weather:current"
	// Note: In a real implementation, you'd serialize the weather struct to JSON
	// For now, we'll just store a simple string
	g.redis.Set(g.ctx, key, weather.Condition, g.config.Game.WeatherUpdateInterval)
}

func (g *GameEngine) GetCurrentWeather() (*models.Weather, error) {
	var weather models.Weather
	err := g.db.DB.Order("created_at DESC").First(&weather).Error
	if err != nil {
		return nil, err
	}
	return &weather, nil
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
