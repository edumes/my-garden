package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/pkg/game"
)

// readiness answers 503 while the database is unreachable, since no request
// can be served without it. The game engine's state is reported as
// information: a stalled engine stops plants from growing, not requests.
func readiness(db *database.Database, engine *game.GameEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		status, databaseState := http.StatusOK, "up"
		sqlDB, err := db.DB.DB()
		if err == nil {
			err = sqlDB.PingContext(ctx)
		}
		if err != nil {
			log.Printf("Readiness: database unreachable: %v", err)
			status, databaseState = http.StatusServiceUnavailable, "down"
		}

		overall := "ready"
		if status != http.StatusOK {
			overall = "not_ready"
		}
		c.JSON(status, gin.H{
			"status": overall,
			"checks": gin.H{
				"database":    gin.H{"status": databaseState},
				"game_engine": engine.Status(),
			},
		})
	}
}
