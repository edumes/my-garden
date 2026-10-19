package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	FindUserByUsername(username string) (*User, error)
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id uuid.UUID) (*User, error)
	CreateUser(user *User) error
	UpdateProfile(user *User) error
	RecordLogin(id uuid.UUID, at time.Time) error
	TokenVersion(id uuid.UUID) (int, error)
	RevokeTokens(id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindUserByUsername(username string) (*User, error) {
	var user User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByID(id uuid.UUID) (*User, error) {
	var user User
	if err := r.db.Preload("Achievements.Achievement").First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

// UpdateProfile writes only the editable profile fields, so it cannot undo
// concurrent changes to coins or the token version.
func (r *repository) UpdateProfile(user *User) error {
	return r.db.Model(user).Select("first_name", "last_name", "avatar", "timezone", "language").Updates(user).Error
}

func (r *repository) RecordLogin(id uuid.UUID, at time.Time) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("last_login_at", at).Error
}

func (r *repository) TokenVersion(id uuid.UUID) (int, error) {
	var user User
	err := r.db.Select("token_version").Take(&user, "id = ?", id).Error
	return user.TokenVersion, err
}

// RevokeTokens invalidates every token issued to the user so far.
func (r *repository) RevokeTokens(id uuid.UUID) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("token_version", gorm.Expr("token_version + 1")).Error
}
