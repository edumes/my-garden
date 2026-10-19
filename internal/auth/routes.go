package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *AuthHandler, requireAuth gin.HandlerFunc) {
	authRoutes := rg.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.POST("/login", handler.Login)
		authRoutes.POST("/refresh", requireAuth, handler.RefreshToken)
		authRoutes.POST("/logout", requireAuth, handler.Logout)
	}

	users := rg.Group("/users")
	users.Use(requireAuth)
	{
		users.GET("/profile", handler.GetProfile)
		users.PUT("/profile", handler.UpdateProfile)
	}
}
