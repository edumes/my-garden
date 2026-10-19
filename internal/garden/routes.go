package garden

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *GardenHandler, shareHandler *GardenShareHandler, requireAuth gin.HandlerFunc) {
	gardens := rg.Group("/gardens")
	gardens.Use(requireAuth)
	{
		gardens.GET("", handler.GetGardens)
		gardens.POST("", handler.CreateGarden)
		gardens.GET("/:id", handler.GetGarden)
		gardens.PUT("/:id", handler.UpdateGarden)
		gardens.DELETE("/:id", handler.DeleteGarden)
		gardens.POST("/:id/plants", handler.PlantSeed)
		gardens.POST("/:id/plants/:plantId/harvest", handler.HarvestPlant)
		gardens.DELETE("/:id/plants/:plantId", handler.RemovePlant)
	}

	gardenShares := rg.Group("/garden-shares")
	gardenShares.Use(requireAuth)
	{
		gardenShares.POST("/access-links", shareHandler.CreateAccessLink)
		gardenShares.POST("/join", shareHandler.JoinGarden)
		gardenShares.GET("/shared-with-me", shareHandler.GetSharedGardens)
		gardenShares.GET("/garden/:garden_id", shareHandler.GetGardenShares)
		gardenShares.PUT("/garden/:garden_id/permissions", shareHandler.UpdateSharePermissions)
		gardenShares.DELETE("/garden/:garden_id/user/:user_id", shareHandler.RemoveGardenShare)
		gardenShares.GET("/garden/:garden_id/access-links", shareHandler.GetAccessLinks)
		gardenShares.POST("/garden/:garden_id/access-links/:link_id/deactivate", shareHandler.DeactivateAccessLink)
	}

	rg.GET("/plants", handler.ListPlantTypes)
}
