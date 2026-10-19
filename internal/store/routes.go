package store

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *StoreHandler, requireAuth gin.HandlerFunc) {
	store := rg.Group("/store")
	store.Use(requireAuth)
	{
		store.GET("/inventory", handler.GetStoreInventory)
		store.POST("/buy", handler.BuySeed)
		store.GET("/inventory/user", handler.GetUserSeedInventory)
	}
}
