package marketplace

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, requireAuth gin.HandlerFunc) {
	marketplace := router.Group("/marketplace")
	marketplace.Use(requireAuth)
	{
		marketplace.GET("/stats", handler.GetMarketplaceStats)

		deliveryRequests := marketplace.Group("/delivery-requests")
		{
			deliveryRequests.POST("", handler.CreateDeliveryRequest)
			deliveryRequests.GET("", handler.GetDeliveryRequests)
			deliveryRequests.POST("/:id/accept", handler.AcceptDeliveryRequest)
			deliveryRequests.POST("/:id/complete", handler.CompleteDeliveryRequest)
		}

		listings := marketplace.Group("/listings")
		{
			listings.POST("", handler.CreatePlantListing)
			listings.GET("", handler.GetPlantListings)
			listings.POST("/:id/buy", handler.PurchasePlantListing)
		}
	}
}
