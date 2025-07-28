package audit

import (
	"github.com/gin-gonic/gin"
	pkgAuth "github.com/my-garden/api/pkg/auth"
)

type AuthMiddlewareFunc func() gin.HandlerFunc

func RegisterRoutes(rg *gin.RouterGroup, handler *AuditHandler, jwtManager *pkgAuth.JWTManager, authMiddleware gin.HandlerFunc) {
	audit := rg.Group("/audit")
	audit.Use(authMiddleware)
	{
		audit.GET("/logs", handler.GetAuditLogs)
		audit.GET("/users/:user_id/activity", handler.GetUserActivity)
		audit.GET("/gardens/:garden_id/activity", handler.GetGardenActivity)
		audit.GET("/stats", handler.GetAuditStats)
	}
}
