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
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/my-garden/api/docs"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/handlers"
	"github.com/my-garden/api/internal/middleware"
	"github.com/my-garden/api/pkg/auth"
	"github.com/my-garden/api/pkg/game"
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

	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg)

	// Initialize game engine
	gameEngine := game.NewGameEngine(db, rdb, cfg)
	gameEngine.Start()
	defer gameEngine.Stop()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtManager)
	gardenHandler := handlers.NewGardenHandler(db)
	weatherHandler := handlers.NewWeatherHandler(db, gameEngine)

	// Initialize router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// Swagger documentation
	router.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := router.Group("/api/v1")
	{
		// Authentication routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// User routes (protected)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/profile", authHandler.GetProfile)
			users.PUT("/profile", authHandler.UpdateProfile)
		}

		// Garden routes (protected)
		gardens := api.Group("/gardens")
		gardens.Use(middleware.AuthMiddleware(jwtManager))
		{
			gardens.GET("", gardenHandler.GetGardens)
			gardens.POST("", gardenHandler.CreateGarden)
			gardens.GET("/:id", gardenHandler.GetGarden)
			gardens.PUT("/:id", gardenHandler.UpdateGarden)
			gardens.DELETE("/:id", gardenHandler.DeleteGarden)

			// Plant routes
			gardens.POST("/:id/plants", gardenHandler.PlantSeed)
			gardens.PUT("/:id/plants/:plantId", gardenHandler.WaterPlant)
			gardens.POST("/:id/plants/:plantId/fertilize", gardenHandler.FertilizePlant)
			gardens.POST("/:id/plants/:plantId/harvest", gardenHandler.HarvestPlant)
			gardens.DELETE("/:id/plants/:plantId", gardenHandler.RemovePlant)
		}

		// Public routes
		api.GET("/plants", gardenHandler.ListPlantTypes)

		api.GET("/weather/current", weatherHandler.GetCurrentWeather)
		api.GET("/weather/forecast", weatherHandler.GetWeatherForecast)
		api.GET("/weather/history", weatherHandler.GetWeatherHistory)

		api.GET("/game/status", func(c *gin.Context) {
			// TODO: Implement game status endpoint
			c.JSON(http.StatusOK, gin.H{"message": "Game status endpoint - coming soon"})
		})

		api.GET("/game/leaderboard", func(c *gin.Context) {
			// TODO: Implement leaderboard endpoint
			c.JSON(http.StatusOK, gin.H{"message": "Leaderboard endpoint - coming soon"})
		})
	}

	// WebSocket routes (protected)
	ws := router.Group("/api/v1/ws")
	ws.Use(middleware.AuthMiddleware(jwtManager))
	{
		ws.GET("/garden/:gardenId", func(c *gin.Context) {
			// TODO: Implement WebSocket handler for real-time garden updates
			c.JSON(http.StatusOK, gin.H{"message": "WebSocket endpoint - coming soon"})
		})
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
