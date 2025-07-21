package middleware

import (
	"bytes"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/models"
	"github.com/my-garden/api/internal/services"
)

type AuditMiddleware struct {
	auditService *services.AuditService
}

func NewAuditMiddleware(auditService *services.AuditService) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (m *AuditMiddleware) AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = writer

		c.Next()

		statusCode := c.Writer.Status()

		var userID *uuid.UUID
		if userIDVal, exists := c.Get("user_id"); exists {
			if id, ok := userIDVal.(uuid.UUID); ok {
				userID = &id
			}
		}

		action := m.determineAction(c)
		if action == models.AuditAction("unknown") {
			return
		}
		resource := m.determineResource(c)
		resourceID := m.extractResourceID(c)
		status := m.determineStatus(statusCode)

		m.auditService.LogEvent(c.Request.Context(), services.AuditEvent{
			UserID:     userID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Status:     status,
			Request:    c.Request,
		})
	}
}

func (m *AuditMiddleware) determineAction(c *gin.Context) models.AuditAction {
	path := c.Request.URL.Path
	method := c.Request.Method

	switch {
	case path == "/api/v1/gardens" && method == "POST":
		return models.AuditActionGardenCreate
	case path == "/api/v1/gardens" && method == "GET":
		return models.AuditActionGardenView
	case path == "/api/v1/gardens" && method == "DELETE":
		return models.AuditActionGardenDelete
	case path == "/api/v1/gardens" && method == "GET":
		return models.AuditActionGardenView
	case path == "/api/v1/gardens" && method == "POST":
		return models.AuditActionPlantSeed
	case path == "/api/v1/gardens" && method == "POST":
		return models.AuditActionPlantHarvest
	case path == "/api/v1/gardens" && method == "DELETE":
		return models.AuditActionPlantRemove
	case path == "/api/v1/garden-shares/access-links" && method == "POST":
		return models.AuditActionAccessLinkCreate
	case path == "/api/v1/garden-shares/join" && method == "POST":
		return models.AuditActionGardenJoin
	case path == "/api/v1/garden-shares" && method == "PUT":
		return models.AuditActionSharePermission
	case path == "/api/v1/garden-shares" && method == "DELETE":
		return models.AuditActionShareRemove
	case path == "/api/v1/store/buy" && method == "POST":
		return models.AuditActionSeedPurchase
	default:
		return models.AuditAction("unknown")
	}
}

func (m *AuditMiddleware) determineResource(c *gin.Context) models.AuditResource {
	path := c.Request.URL.Path

	switch {
	case path == "/api/v1/auth/register" || path == "/api/v1/auth/login" || path == "/api/v1/auth/logout" || path == "/api/v1/auth/refresh":
		return models.AuditResourceUser
	case path == "/api/v1/users/profile":
		return models.AuditResourceUser
	case path == "/api/v1/gardens":
		return models.AuditResourceGarden
	case path == "/api/v1/plants":
		return models.AuditResourcePlant
	case path == "/api/v1/garden-shares":
		return models.AuditResourceShare
	case path == "/api/v1/store":
		return models.AuditResourceStore
	default:
		return models.AuditResourceSystem
	}
}

func (m *AuditMiddleware) extractResourceID(c *gin.Context) *uuid.UUID {
	if id := c.Param("id"); id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			return &parsed
		}
	}
	if id := c.Param("garden_id"); id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			return &parsed
		}
	}
	if id := c.Param("plantId"); id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			return &parsed
		}
	}
	if id := c.Param("user_id"); id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			return &parsed
		}
	}
	return nil
}

func (m *AuditMiddleware) determineStatus(statusCode int) models.AuditStatus {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return models.AuditStatusSuccess
	case statusCode >= 400 && statusCode < 500:
		return models.AuditStatusFailure
	case statusCode >= 500:
		return models.AuditStatusError
	default:
		return models.AuditStatusFailure
	}
}
