package weather

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *WeatherHandler) {
	rg.GET("/weather/current", handler.GetCurrentWeather)
	rg.GET("/weather/forecast", handler.GetWeatherForecast)
	rg.GET("/weather/history", handler.GetWeatherHistory)
}
