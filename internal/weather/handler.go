package weather

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	service *Service
}

func NewWeatherHandler(service *Service) *WeatherHandler {
	return &WeatherHandler{
		service: service,
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
	weather, err := h.service.GetCurrentWeather()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weather"})
		return
	}

	// Get current season
	season := GetSeason(time.Now())

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
	forecasts, err := h.service.GetWeatherForecast()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate forecast"})
		return
	}

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

	weatherHistory, err := h.service.GetWeatherHistory(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weather history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": weatherHistory,
		"count":   len(weatherHistory),
	})
}
