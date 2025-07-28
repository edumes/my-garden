package auth

import (
	"github.com/gin-gonic/gin"
	pkgAuth "github.com/my-garden/api/pkg/auth"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *AuthHandler, jwtManager *pkgAuth.JWTManager) {
	authRoutes := rg.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.POST("/login", handler.Login)
		authRoutes.POST("/refresh", handler.RefreshToken)
		authRoutes.POST("/logout", AuthMiddleware(jwtManager), handler.Logout)
	}

	users := rg.Group("/users")
	users.Use(AuthMiddleware(jwtManager))
	{
		users.GET("/profile", handler.GetProfile)
		users.PUT("/profile", handler.UpdateProfile)
	}
}
