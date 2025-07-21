package audit

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/pkg/db"
)

type AuditService struct {
	db         db.Database
	repository *AuditRepository
}

func NewAuditService(db db.Database) *AuditService {
	return &AuditService{
		db:         db,
		repository: NewAuditRepository(db),
	}
}

type AuditEvent struct {
	UserID     *uuid.UUID
	Action     AuditAction
	Resource   AuditResource
	ResourceID *uuid.UUID
	Status     AuditStatus
	Request    *http.Request
}

func (s *AuditService) LogEvent(ctx context.Context, event AuditEvent) error {
	// Only log known actions
	switch event.Action {
	case AuditActionUserRegister,
		AuditActionUserLogin,
		AuditActionUserLogout,
		AuditActionTokenRefresh,
		AuditActionProfileUpdate,
		AuditActionGardenCreate,
		AuditActionGardenUpdate,
		AuditActionGardenDelete,
		AuditActionGardenView,
		AuditActionPlantSeed,
		AuditActionPlantHarvest,
		AuditActionPlantRemove,
		AuditActionPlantWater,
		AuditActionPlantFertilize,
		AuditActionGardenShare,
		AuditActionGardenJoin,
		AuditActionSharePermission,
		AuditActionShareRemove,
		AuditActionAccessLinkCreate,
		AuditActionAccessLinkDeactivate,
		AuditActionSeedPurchase,
		AuditActionInventoryView,
		AuditActionError,
		AuditActionSecurity:
		// continue
	default:
		return nil
	}

	auditLog := AuditLog{
		UserID:     event.UserID,
		Action:     string(event.Action),
		Resource:   string(event.Resource),
		ResourceID: event.ResourceID,
		Status:     string(event.Status),
		CreatedAt:  time.Now(),
	}

	if event.Request != nil {
		auditLog.IPAddress = s.getClientIP(event.Request)
	}

	return s.repository.CreateLog(ctx, &auditLog)
}

func (s *AuditService) LogGinEvent(c *gin.Context, action AuditAction, resource AuditResource, resourceID *uuid.UUID, details interface{}, status AuditStatus) error {
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if id, ok := userIDVal.(uuid.UUID); ok {
			userID = &id
		}
	}

	event := AuditEvent{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Status:     status,
		Request:    c.Request,
	}

	return s.LogEvent(c.Request.Context(), event)
}

func (s *AuditService) LogSuccess(c *gin.Context, action AuditAction, resource AuditResource, resourceID *uuid.UUID, details interface{}) error {
	return s.LogGinEvent(c, action, resource, resourceID, details, AuditStatusSuccess)
}

func (s *AuditService) LogFailure(c *gin.Context, action AuditAction, resource AuditResource, resourceID *uuid.UUID, details interface{}) error {
	return s.LogGinEvent(c, action, resource, resourceID, details, AuditStatusFailure)
}

func (s *AuditService) LogError(c *gin.Context, action AuditAction, resource AuditResource, resourceID *uuid.UUID, details interface{}) error {
	return s.LogGinEvent(c, action, resource, resourceID, details, AuditStatusError)
}

func (s *AuditService) GetAuditLogs(ctx context.Context, filters AuditFilters) ([]AuditLog, int64, error) {
	return s.repository.GetLogs(ctx, filters)
}

func (s *AuditService) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int) ([]AuditLog, error) {
	return s.repository.GetUserActivity(ctx, userID, limit)
}

func (s *AuditService) GetGardenActivity(ctx context.Context, gardenID uuid.UUID, limit int) ([]AuditLog, error) {
	return s.repository.GetGardenActivity(ctx, gardenID, limit)
}

func (s *AuditService) getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
