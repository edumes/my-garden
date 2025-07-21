package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
)

type AuditService struct {
	db *database.Database
}

func NewAuditService(db *database.Database) *AuditService {
	return &AuditService{db: db}
}

type AuditEvent struct {
	UserID     *uuid.UUID
	Action     models.AuditAction
	Resource   models.AuditResource
	ResourceID *uuid.UUID
	Status     models.AuditStatus
	Request    *http.Request
}

func (s *AuditService) LogEvent(ctx context.Context, event AuditEvent) error {
	// Only log known actions
	switch event.Action {
	case models.AuditActionUserRegister,
		models.AuditActionUserLogin,
		models.AuditActionUserLogout,
		models.AuditActionTokenRefresh,
		models.AuditActionProfileUpdate,
		models.AuditActionGardenCreate,
		models.AuditActionGardenUpdate,
		models.AuditActionGardenDelete,
		models.AuditActionGardenView,
		models.AuditActionPlantSeed,
		models.AuditActionPlantHarvest,
		models.AuditActionPlantRemove,
		models.AuditActionPlantWater,
		models.AuditActionPlantFertilize,
		models.AuditActionGardenShare,
		models.AuditActionGardenJoin,
		models.AuditActionSharePermission,
		models.AuditActionShareRemove,
		models.AuditActionAccessLinkCreate,
		models.AuditActionAccessLinkDeactivate,
		models.AuditActionSeedPurchase,
		models.AuditActionInventoryView,
		models.AuditActionError,
		models.AuditActionSecurity:
		// continue
	default:
		return nil
	}

	auditLog := models.AuditLog{
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

	return s.db.DB.WithContext(ctx).Create(&auditLog).Error
}

func (s *AuditService) LogGinEvent(c *gin.Context, action models.AuditAction, resource models.AuditResource, resourceID *uuid.UUID, details interface{}, status models.AuditStatus) error {
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

func (s *AuditService) LogSuccess(c *gin.Context, action models.AuditAction, resource models.AuditResource, resourceID *uuid.UUID, details interface{}) error {
	return s.LogGinEvent(c, action, resource, resourceID, details, models.AuditStatusSuccess)
}

func (s *AuditService) LogFailure(c *gin.Context, action models.AuditAction, resource models.AuditResource, resourceID *uuid.UUID, details interface{}) error {
	return s.LogGinEvent(c, action, resource, resourceID, details, models.AuditStatusFailure)
}

func (s *AuditService) LogError(c *gin.Context, action models.AuditAction, resource models.AuditResource, resourceID *uuid.UUID, details interface{}) error {
	return s.LogGinEvent(c, action, resource, resourceID, details, models.AuditStatusError)
}

func (s *AuditService) GetAuditLogs(ctx context.Context, filters AuditFilters) ([]models.AuditLog, int64, error) {
	query := s.db.DB.WithContext(ctx).Model(&models.AuditLog{}).Preload("User")

	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}

	if filters.Resource != "" {
		query = query.Where("resource = ?", filters.Resource)
	}

	if filters.ResourceID != nil {
		query = query.Where("resource_id = ?", *filters.ResourceID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if !filters.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filters.StartDate)
	}

	if !filters.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filters.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

type AuditFilters struct {
	UserID     *uuid.UUID
	Action     string
	Resource   string
	ResourceID *uuid.UUID
	Status     string
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	Offset     int
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

func (s *AuditService) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Preload("User").
		Find(&logs).Error
	return logs, err
}

func (s *AuditService) GetGardenActivity(ctx context.Context, gardenID uuid.UUID, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.DB.WithContext(ctx).
		Where("resource_id = ? AND resource = ?", gardenID, models.AuditResourceGarden).
		Order("created_at DESC").
		Limit(limit).
		Preload("User").
		Find(&logs).Error
	return logs, err
}
