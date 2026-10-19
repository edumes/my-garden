package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/audit"
	auth "github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/marketplace"
	"github.com/my-garden/api/internal/middleware"
	"github.com/my-garden/api/internal/store"
	"github.com/my-garden/api/internal/weather"
	pkgAuth "github.com/my-garden/api/pkg/auth"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// newRouter wires every handler and middleware of the HTTP API.
func newRouter(db *database.Database, rdb *redis.Client, cfg *config.Config) *gin.Engine {
	// Initialize JWT manager
	jwtManager := pkgAuth.NewJWTManager(cfg)

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
	router := gin.New()

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

	return router
}
