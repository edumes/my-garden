// @title Virtual Garden Management System API
// @version 1.0
// @description A robust RESTful API for a virtual garden management game built with Go.
// @BasePath /api/v1
// @securityDefinitions.apikey bearer
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"

package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/middleware"
	"github.com/my-garden/api/internal/store"
	"github.com/my-garden/api/internal/weather"
	"github.com/my-garden/api/internal/websocket"
	pkgAuth "github.com/my-garden/api/pkg/auth"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Initialize JWT manager
	jwtManager := pkgAuth.NewJWTManager(cfg)

	// Initialize WebSocket hub and handler
	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewHandler(wsHub)

	// Initialize services and handlers
	authRepo := auth.NewRepository(db.GetDB())
	authService := auth.NewService(authRepo, jwtManager)
	auditService := audit.NewAuditService(db)
	authHandler := auth.NewAuthHandler(authService, auditService)
	weatherRepo := weather.NewRepository(db)
	weatherService := weather.NewService(weatherRepo, redisClient, wsHandler)
	weatherHandler := weather.NewWeatherHandler(weatherService)
	storeRepo := store.NewRepository(db)
	storeService := store.NewService(storeRepo, auditService)
	storeHandler := store.NewStoreHandler(storeService, auditService)
	gardenHandler := garden.NewGardenHandler(db, auditService, authService, wsHandler)
	gardenShareHandler := garden.NewGardenShareHandler(db)
	auditHandler := audit.NewAuditHandler(db, auditService)

	// Initialize router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg))

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		api.Use(auth.AuthMiddleware(jwtManager))
		{
			// Garden routes
			api.GET("/gardens", gardenHandler.GetGardens)
			api.POST("/gardens", gardenHandler.CreateGarden)
			api.GET("/gardens/:id", gardenHandler.GetGarden)
			api.PUT("/gardens/:id", gardenHandler.UpdateGarden)
			api.DELETE("/gardens/:id", gardenHandler.DeleteGarden)
			api.POST("/gardens/:id/plants", gardenHandler.PlantSeed)
			api.POST("/gardens/:id/plants/:plantId/harvest", gardenHandler.HarvestPlant)
			api.DELETE("/gardens/:id/plants/:plantId", gardenHandler.RemovePlant)

			// Garden sharing routes
			api.POST("/gardens/:id/access-links", gardenShareHandler.CreateAccessLink)
			api.GET("/gardens/:id/shares", gardenShareHandler.GetGardenShares)
			api.DELETE("/gardens/:id/shares/:userId", gardenShareHandler.RemoveGardenShare)

			// Store routes
			api.GET("/store/inventory", storeHandler.GetStoreInventory)
			api.POST("/store/buy", storeHandler.BuySeed)
			api.GET("/store/inventory/user", storeHandler.GetUserSeedInventory)

			// Weather routes
			api.GET("/weather/current", weatherHandler.GetCurrentWeather)
			api.GET("/weather/forecast", weatherHandler.GetWeatherForecast)
			api.GET("/weather/history", weatherHandler.GetWeatherHistory)

			// Audit routes
			api.GET("/audit/logs", auditHandler.GetAuditLogs)
			api.GET("/audit/activity", auditHandler.GetUserActivity)

			// WebSocket endpoint
			api.GET("/ws", wsHandler.HandleWebSocket)

			// Leaderboard endpoint (placeholder)
			api.GET("/leaderboard", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Leaderboard endpoint - coming soon"})
			})
		}
	}

	// Swagger documentation
	router.GET("api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
