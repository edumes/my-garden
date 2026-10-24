package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Avatar       string    `json:"avatar"`
	Level        int       `json:"level" gorm:"default:1"`
	Experience   int       `json:"experience" gorm:"default:0"`
	// Experience at which the current level started and the next one starts.
	LevelExperience     int               `json:"level_experience" gorm:"-"`
	NextLevelExperience int               `json:"next_level_experience" gorm:"-"`
	GardenLimit         int               `json:"garden_limit" gorm:"-"`
	NextGardenLevel     int               `json:"next_garden_level" gorm:"-"`
	Coins               int               `json:"coins" gorm:"default:100;check:chk_users_coins,coins >= 0"`
	Timezone            string            `json:"timezone" gorm:"default:'UTC'"`
	Language            string            `json:"language" gorm:"default:'en'"`
	LastLoginAt         *time.Time        `json:"last_login_at"`
	TokenVersion        int               `json:"-" gorm:"not null;default:0"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	Achievements        []UserAchievement `gorm:"foreignKey:UserID"`
}

// PublicUser is the part of a profile other players may see. Use it for
// relations that expose someone else's account, such as a garden's owner.
type PublicUser struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	Level    int       `json:"level"`
}

func (PublicUser) TableName() string { return "users" }

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
