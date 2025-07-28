package audit

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
