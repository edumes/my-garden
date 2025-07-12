package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Avatar       string    `json:"avatar"`

	// Game progression
	Level      int `json:"level" gorm:"default:1"`
	Experience int `json:"experience" gorm:"default:0"`
	Coins      int `json:"coins" gorm:"default:100"`

	// Game settings
	Timezone string `json:"timezone" gorm:"default:'UTC'"`
	Language string `json:"language" gorm:"default:'en'"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"`

	// Relationships
	Gardens      []Garden          `json:"gardens,omitempty" gorm:"foreignKey:UserID"`
	Achievements []UserAchievement `json:"achievements,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserAchievement represents user achievements
type UserAchievement struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	AchievementID uuid.UUID `json:"achievement_id" gorm:"type:uuid;not null"`
	UnlockedAt    time.Time `json:"unlocked_at"`

	// Relationships
	User        User        `json:"user" gorm:"foreignKey:UserID"`
	Achievement Achievement `json:"achievement" gorm:"foreignKey:AchievementID"`
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

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a *Achievement) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
