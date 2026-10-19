package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string            `json:"username" gorm:"uniqueIndex;not null"`
	Email        string            `json:"email" gorm:"uniqueIndex;not null"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	PasswordHash string            `json:"-" gorm:"not null"`
	Avatar       string            `json:"avatar"`
	Level        int               `json:"level" gorm:"default:1"`
	Experience   int               `json:"experience" gorm:"default:0"`
	Coins        int               `json:"coins" gorm:"default:100"`
	Timezone     string            `json:"timezone" gorm:"default:'UTC'"`
	Language     string            `json:"language" gorm:"default:'en'"`
	LastLoginAt  *time.Time        `json:"last_login_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Achievements []UserAchievement `gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserAchievement represents user achievements
type UserAchievement struct {
	ID            uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	AchievementID uuid.UUID   `json:"achievement_id" gorm:"type:uuid;not null"`
	UnlockedAt    time.Time   `json:"unlocked_at"`
	Achievement   Achievement `gorm:"foreignKey:AchievementID"`
}

func (ua *UserAchievement) BeforeCreate(tx *gorm.DB) error {
	if ua.ID == uuid.Nil {
		ua.ID = uuid.New()
	}
	return nil
}

// Achievement represents available achievements
type Achievement struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Points      int       `json:"points" gorm:"default:0"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a *Achievement) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
