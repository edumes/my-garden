package audit

import (
	"context"
	"gorm.io/gorm"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/pkg/db"
)

type AuditRepository struct {
	db db.Database
}

func NewAuditRepository(db db.Database) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) CreateLog(ctx context.Context, log *AuditLog) error {
	return r.db.GetDB().WithContext(ctx).Create(log).Error
}

func (r *AuditRepository) GetLogs(ctx context.Context, filters AuditFilters) ([]AuditLog, int64, error) {
	query := r.db.GetDB().WithContext(ctx).Model(&AuditLog{})

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

	var logs []AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *AuditRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	err := r.db.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *AuditRepository) GetGardenActivity(ctx context.Context, userID, gardenID uuid.UUID, limit int) ([]AuditLog, error) {
	var logs []AuditLog
	err := r.db.GetDB().WithContext(ctx).
		Where("user_id = ? AND resource_id = ? AND resource = ?", userID, gardenID, AuditResourceGarden).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetAuditStats summarizes one user's audit logs since startDate.
func (r *AuditRepository) GetAuditStats(ctx context.Context, userID uuid.UUID, startDate time.Time) (AuditStats, error) {
	var stats AuditStats
	logs := func() *gorm.DB {
		return r.db.GetDB().WithContext(ctx).Model(&AuditLog{}).
			Where("user_id = ? AND created_at >= ?", userID, startDate)
	}

	if err := logs().Count(&stats.TotalActions).Error; err != nil {
		return stats, err
	}
	if err := logs().Where("status = ?", string(AuditStatusSuccess)).Count(&stats.SuccessfulActions).Error; err != nil {
		return stats, err
	}
	if err := logs().Where("status = ?", string(AuditStatusFailure)).Count(&stats.FailedActions).Error; err != nil {
		return stats, err
	}
	if err := logs().Where("status = ?", string(AuditStatusError)).Count(&stats.ErrorActions).Error; err != nil {
		return stats, err
	}
	if err := logs().Select("action, COUNT(*) as count").Group("action").Order("count DESC").Limit(10).
		Find(&stats.TopActions).Error; err != nil {
		return stats, err
	}
	if err := logs().Select("resource, COUNT(*) as count").Group("resource").Order("count DESC").Limit(10).
		Find(&stats.TopResources).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
