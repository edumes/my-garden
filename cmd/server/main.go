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
	"github.com/my-garden/api/internal/audit"
	auth "github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/marketplace"
	"github.com/my-garden/api/internal/middleware"
	"github.com/my-garden/api/internal/migrations"
	"github.com/my-garden/api/internal/store"
	"github.com/my-garden/api/internal/weather"
	pkgAuth "github.com/my-garden/api/pkg/auth"
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

	// Run migrations
	gormDB := db.GetDB()
	if err := migrations.AutoMigrate(gormDB); err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	// Run seeds
	if err := migrations.Seed(gormDB); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

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
	jwtManager := pkgAuth.NewJWTManager(cfg)

	// Initialize game engine
	gameEngine := game.NewGameEngine(db, rdb, cfg)
	gameEngine.Start()
	defer gameEngine.Stop()

	// Initialize services
	auditService := audit.NewAuditService(db)

	// Initialize auth components
	authRepo := auth.NewRepository(db.GetDB())
	authService := auth.NewService(authRepo, jwtManager)
	authHandler := auth.NewAuthHandler(authService, auditService)

	// Initialize store components
	storeRepo := store.NewRepository(db)
	storeService := store.NewService(storeRepo, auditService)
	storeHandler := store.NewStoreHandler(storeService, auditService)

	// Initialize weather components
	weatherRepo := weather.NewRepository(db)
	weatherService := weather.NewService(weatherRepo, rdb)
	weatherHandler := weather.NewWeatherHandler(weatherService)

	// Initialize marketplace components
	marketplaceRepo := marketplace.NewRepository(db.GetDB())
	marketplaceService := marketplace.NewService(marketplaceRepo, auditService)
	marketplaceHandler := marketplace.NewHandler(marketplaceService, auditService)

	// Initialize other handlers
	gardenHandler := garden.NewGardenHandler(db, auditService)
	gardenShareHandler := garden.NewGardenShareHandler(db)
	auditHandler := audit.NewAuditHandler(db, auditService)

	// Initialize router
	router := gin.Default()

	// Initialize middleware
	auditMiddleware := audit.NewAuditMiddleware(auditService)

	// Add middleware
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(auditMiddleware.AuditLog())

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

	// Setup API v1 routes
	api := router.Group("/api/v1")
	auth.RegisterRoutes(api, authHandler, jwtManager)
	garden.RegisterRoutes(api, gardenHandler, gardenShareHandler, jwtManager)
	store.RegisterRoutes(api, storeHandler, auditService, jwtManager)
	weather.RegisterRoutes(api, weatherHandler)
	marketplace.RegisterRoutes(api, marketplaceHandler, auditService, jwtManager)
	audit.RegisterRoutes(api, auditHandler, jwtManager, auth.AuthMiddleware(jwtManager))

	// Game and WebSocket endpoints (if not domain-specific, keep here)
	api.GET("/game/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Game status endpoint - coming soon"})
	})
	api.GET("/game/leaderboard", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Leaderboard endpoint - coming soon"})
	})

	ws := router.Group("/api/v1/ws")
	ws.Use(auth.AuthMiddleware(jwtManager))
	{
		ws.GET("/garden/:gardenId", func(c *gin.Context) {
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
