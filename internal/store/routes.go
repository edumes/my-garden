package store

import (
	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	pkgAuth "github.com/my-garden/api/pkg/auth"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *StoreHandler, auditService *audit.AuditService, jwtManager *pkgAuth.JWTManager) {
	store := rg.Group("/store")
	store.Use(auth.AuthMiddleware(jwtManager))
	{
		store.GET("/inventory", handler.GetStoreInventory)
		store.POST("/buy", handler.BuySeed)
		store.GET("/inventory/user", handler.GetUserSeedInventory)
	}
}
