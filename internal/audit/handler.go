package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/pkg/db"
)

type AuditHandler struct {
	db           db.Database
	auditService *AuditService
}

func NewAuditHandler(db db.Database, auditService *AuditService) *AuditHandler {
	return &AuditHandler{
		db:           db,
		auditService: auditService,
	}
}

type GetAuditLogsRequest struct {
	UserID     *uuid.UUID `form:"user_id"`
	Action     string     `form:"action"`
	Resource   string     `form:"resource"`
	ResourceID *uuid.UUID `form:"resource_id"`
	Status     string     `form:"status"`
	StartDate  string     `form:"start_date"`
	EndDate    string     `form:"end_date"`
	Limit      int        `form:"limit,default=50"`
	Offset     int        `form:"offset,default=0"`
}

// GetAuditLogs godoc
// @Summary Get audit logs
// @Description Get audit logs with optional filtering
// @Tags audit
// @Accept json
// @Produce json
// @Security bearer
// @Param user_id query string false "User ID filter"
// @Param action query string false "Action filter"
// @Param resource query string false "Resource filter"
// @Param resource_id query string false "Resource ID filter"
// @Param status query string false "Status filter"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset results" default(0)
// @Success 200 {object} map[string]interface{} "Audit logs with pagination"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /audit/logs [get]
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	var req GetAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filters := AuditFilters{
		UserID:     req.UserID,
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			filters.StartDate = startDate
		}
	}

	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			filters.EndDate = endDate.Add(24 * time.Hour)
		}
	}

	logs, total, err := h.auditService.GetAuditLogs(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"total":    total,
		"limit":    req.Limit,
		"offset":   req.Offset,
		"has_more": total > int64(req.Offset+req.Limit),
	})
}

// GetUserActivity godoc
// @Summary Get user activity
// @Description Get recent activity for a specific user
// @Tags audit
// @Accept json
// @Produce json
// @Security bearer
// @Param user_id path string true "User ID"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {object} map[string]interface{} "User activity logs"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /audit/users/{user_id}/activity [get]
func (h *AuditHandler) GetUserActivity(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	logs, err := h.auditService.GetUserActivity(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"activity": logs,
		"count":    len(logs),
	})
}

// GetGardenActivity godoc
// @Summary Get garden activity
// @Description Get recent activity for a specific garden
// @Tags audit
// @Accept json
// @Produce json
// @Security bearer
// @Param garden_id path string true "Garden ID"
// @Param limit query int false "Limit results" default(20)
// @Success 200 {object} map[string]interface{} "Garden activity logs"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /audit/gardens/{garden_id}/activity [get]
func (h *AuditHandler) GetGardenActivity(c *gin.Context) {
	gardenID, err := uuid.Parse(c.Param("garden_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid garden ID"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	logs, err := h.auditService.GetGardenActivity(c.Request.Context(), gardenID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch garden activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"garden_id": gardenID,
		"activity":  logs,
		"count":     len(logs),
	})
}

// GetAuditStats godoc
// @Summary Get audit statistics
// @Description Get audit statistics for the system
// @Tags audit
// @Accept json
// @Produce json
// @Security bearer
// @Param days query int false "Number of days to analyze" default(30)
// @Success 200 {object} map[string]interface{} "Audit statistics"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /audit/stats [get]
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if parsed, err := strconv.Atoi(daysStr); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)

	stats, err := h.auditService.repository.GetAuditStats(c.Request.Context(), startDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days": days,
		"start_date":  startDate,
		"end_date":    time.Now(),
		"stats":       stats,
	})
}
