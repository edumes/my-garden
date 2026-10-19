package game

import (
	"context"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/weather"
)

// GameEngine runs the game tick: every GAME_TICK_INTERVAL it draws new
// weather once the current one expires and grows plants up to the present.
// Ticks are safe to repeat and to run on several instances at once.
type GameEngine struct {
	db       *database.Database
	config   *config.Config
	weather  *weather.Service
	cancel   context.CancelFunc
	running  atomic.Bool
	lastTick atomic.Int64 // Unix nanoseconds of the last successful tick
}

func NewGameEngine(db *database.Database, cfg *config.Config) *GameEngine {
	return &GameEngine{
		db:      db,
		config:  cfg,
		weather: weather.NewService(weather.NewRepository(db)),
	}
}

// Start ticks right away, so a restart catches up at once, and then on every
// interval until Stop.
func (g *GameEngine) Start() {
	log.Println("Starting game engine...")
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel
	g.running.Store(true)

	go func() {
		ticker := time.NewTicker(g.config.Game.TickInterval)
		defer ticker.Stop()
		for {
			if err := g.Tick(time.Now()); err != nil {
				log.Printf("Game tick failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (g *GameEngine) Stop() {
	log.Println("Stopping game engine...")
	g.running.Store(false)
	if g.cancel != nil {
		g.cancel()
	}
}

// Tick draws new weather if the current one has expired and grows every
// plant up to now under the current weather.
func (g *GameEngine) Tick(now time.Time) error {
	// Without weather, plants still grow at the normal rate.
	multiplier := 1.0
	current, err := g.weather.RefreshWeather(now, g.config.Game.WeatherUpdateInterval, g.generateWeather)
	if err == nil {
		multiplier = current.GrowthMultiplier
	} else {
		log.Printf("Growing plants without weather effects: %v", err)
	}

	if err := garden.GrowPlants(g.db.DB, now, multiplier, g.config.Game.Speed); err != nil {
		return err
	}
	g.lastTick.Store(now.UnixNano())
	return nil
}

// Status describes the engine for readiness checks.
type Status struct {
	// State is "running", "stale" when no tick has succeeded for three
	// intervals, or "stopped".
	State    string     `json:"status"`
	LastTick *time.Time `json:"last_tick,omitempty"`
}

func (g *GameEngine) Status() Status {
	var status Status
	if nanos := g.lastTick.Load(); nanos != 0 {
		last := time.Unix(0, nanos)
		status.LastTick = &last
	}
	switch {
	case !g.running.Load():
		status.State = "stopped"
	case status.LastTick == nil || time.Since(*status.LastTick) > 3*g.config.Game.TickInterval:
		status.State = "stale"
	default:
		status.State = "running"
	}
	return status
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
