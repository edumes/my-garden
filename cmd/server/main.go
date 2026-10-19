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
	"github.com/my-garden/api/internal/migrations"
	"github.com/my-garden/api/pkg/game"
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

	// Initialize game engine
	gameEngine := game.NewGameEngine(db, cfg)
	gameEngine.Start()
	defer gameEngine.Stop()

	router := newRouter(db, gameEngine, cfg)

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
