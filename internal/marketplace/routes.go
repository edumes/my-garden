package marketplace

import (
	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/auth"
	pkgAuth "github.com/my-garden/api/pkg/auth"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, jwtManager *pkgAuth.JWTManager) {
	marketplace := router.Group("/marketplace")
	marketplace.Use(auth.AuthMiddleware(jwtManager))
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
