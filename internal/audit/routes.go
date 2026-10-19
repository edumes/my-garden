package audit

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, handler *AuditHandler, requireAuth gin.HandlerFunc) {
	audit := rg.Group("/audit")
	audit.Use(requireAuth)
	{
		audit.GET("/logs", handler.GetAuditLogs)
		audit.GET("/users/:user_id/activity", handler.GetUserActivity)
		audit.GET("/gardens/:garden_id/activity", handler.GetGardenActivity)
		audit.GET("/stats", handler.GetAuditStats)
	}
}
