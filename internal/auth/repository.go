package auth

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	FindUserByUsername(username string) (*User, error)
	FindUserByEmail(email string) (*User, error)
	FindUserByID(id uuid.UUID) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
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

func (r *repository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}
