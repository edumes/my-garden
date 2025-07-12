package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
	"github.com/my-garden/api/pkg/game"
	"gorm.io/gorm"
)

type WeatherHandler struct {
	db         *database.Database
	gameEngine *game.GameEngine
}

func NewWeatherHandler(db *database.Database, gameEngine *game.GameEngine) *WeatherHandler {
	return &WeatherHandler{
		db:         db,
		gameEngine: gameEngine,
	}
}

// GetCurrentWeather godoc
// @Summary Get current weather
// @Description Get current weather conditions and season
// @Tags weather
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Current weather and season"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /weather/current [get]
func (h *WeatherHandler) GetCurrentWeather(c *gin.Context) {
	weather, err := h.gameEngine.GetCurrentWeather()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Note: updateWeather is private, so we'll just return an error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No weather data available"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weather"})
			return
		}
	}

	// Get current season
	season := models.GetSeason(time.Now())

	response := gin.H{
		"weather":    weather,
		"season":     season,
		"updated_at": weather.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetWeatherForecast godoc
// @Summary Get weather forecast
// @Description Get weather predictions for next 24 hours
// @Tags weather
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Weather forecasts"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /weather/forecast [get]
func (h *WeatherHandler) GetWeatherForecast(c *gin.Context) {
	// Get forecast for next 24 hours (4 periods of 6 hours each)
	var forecasts []models.WeatherForecast

	// Generate forecast data (in a real app, this would use a weather API)
	forecasts = h.generateForecast()

	c.JSON(http.StatusOK, gin.H{
		"forecasts":    forecasts,
		"generated_at": time.Now(),
	})
}

// GetWeatherHistory godoc
// @Summary Get weather history
// @Description Get historical weather data for the last 24 records
// @Tags weather
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Weather history"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /weather/history [get]
func (h *WeatherHandler) GetWeatherHistory(c *gin.Context) {
	limit := 24 // Last 24 weather records

	var weatherHistory []models.Weather
	if err := h.db.DB.Order("created_at DESC").Limit(limit).Find(&weatherHistory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": weatherHistory,
		"count":   len(weatherHistory),
	})
}

// generateForecast creates a simple weather forecast
func (h *WeatherHandler) generateForecast() []models.WeatherForecast {
	var forecasts []models.WeatherForecast

	// Get current weather to base forecast on
	currentWeather, err := h.gameEngine.GetCurrentWeather()
	if err != nil {
		// Use default weather if current weather not available
		currentWeather = &models.Weather{
			Condition:   models.WeatherCloudy,
			Temperature: 20.0,
			Humidity:    50,
		}
	}

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
			humidity = max(30, humidity-10)
		case hour >= 13 && hour <= 18: // Afternoon
			temperature += 5
			humidity = max(20, humidity-20)
		case hour >= 19 && hour <= 23: // Evening
			temperature -= 2
			humidity = minInt(80, humidity+10)
		default: // Night
			temperature -= 5
			humidity = minInt(90, humidity+20)
		}

		forecast := models.WeatherForecast{
			Condition:   condition,
			Temperature: temperature,
			Humidity:    humidity,
			Probability: 85, // 85% confidence
			ForecastFor: forecastTime,
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
