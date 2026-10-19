package marketplace

import (
	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	pkgAuth "github.com/my-garden/api/pkg/auth"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, auditService *audit.AuditService, jwtManager *pkgAuth.JWTManager) {
	// Marketplace routes
	marketplace := router.Group("/marketplace")
	marketplace.Use(auth.AuthMiddleware(jwtManager))
	{
		// Marketplace stats
		marketplace.GET("/stats", handler.GetMarketplaceStats)

		// Delivery requests
		deliveryRequests := marketplace.Group("/delivery-requests")
		{
			deliveryRequests.POST("", handler.CreateDeliveryRequest)
			deliveryRequests.GET("", handler.GetDeliveryRequests)
			deliveryRequests.POST("/:id/accept", handler.AcceptDeliveryRequest)
			deliveryRequests.POST("/:id/complete", handler.CompleteDeliveryRequest)
		}

		// Plant listings
		listings := marketplace.Group("/listings")
		{
			listings.POST("", handler.CreatePlantListing)
			listings.GET("", handler.GetPlantListings)
			listings.POST("/:id/buy", handler.PurchasePlantListing)
		}
	}

	// Blockchain routes
	blockchain := router.Group("/blockchain")
	{
		blockchain.POST("/transactions", handler.SubmitTransaction)
		blockchain.GET("/transactions/:hash", handler.GetTransactionStatus)
		blockchain.GET("/ledger", handler.GetBlockchainLedger)
	}

	// Wallet routes
	wallet := router.Group("/wallet")
	wallet.Use(auth.AuthMiddleware(jwtManager))
	{
		wallet.POST("", handler.CreateWallet)
		wallet.GET("/balance", handler.GetWalletBalance)
		wallet.POST("/transfer", handler.TransferCurrency)
		wallet.POST("/sign", handler.SignTransaction)
	}
}
