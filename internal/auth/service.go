package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/types"
	"github.com/my-garden/api/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUsernameTaken   = errors.New("username already taken")
	ErrEmailTaken      = errors.New("email already taken")
	ErrTokenGeneration = errors.New("failed to generate token")
	ErrInvalidToken    = errors.New("invalid or expired token")
)

type Service interface {
	Register(req RegisterRequest) (*AuthResponse, error)
	Login(req LoginRequest) (*AuthResponse, error)
	RefreshToken(tokenString string) (string, time.Time, error)
	GetProfile(userID uuid.UUID) (*User, error)
	UpdateProfile(userID uuid.UUID, firstName, lastName, avatar, timezone, language string) (*User, error)
}

type service struct {
	repo       Repository
	jwtManager *auth.JWTManager
}

func NewService(repo Repository, jwtManager *auth.JWTManager) Service {
	return &service{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *service) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if username exists
	if _, err := s.repo.FindUserByUsername(req.Username); err == nil {
		return nil, ErrUsernameTaken
	}

	// Check if email exists
	if _, err := s.repo.FindUserByEmail(req.Email); err == nil {
		return nil, ErrEmailTaken
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now()
	user := &User{
		User: types.User{
			Username:  req.Username,
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
		},
		PasswordHash: string(hashedPassword),
		LastLoginAt:  &now,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	return &AuthResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *service) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByUsername(req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		return nil, ErrTokenGeneration
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token:     token,
		User:      *user,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

func (s *service) RefreshToken(tokenString string) (string, time.Time, error) {
	newToken, err := s.jwtManager.RefreshToken(tokenString)
	if err != nil {
		return "", time.Time{}, ErrInvalidToken
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	return newToken, expiresAt, nil
}

func (s *service) GetProfile(userID uuid.UUID) (*User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) UpdateProfile(userID uuid.UUID, firstName, lastName, avatar, timezone, language string) (*User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if firstName != "" {
		user.FirstName = firstName
	}
	if lastName != "" {
		user.LastName = lastName
	}
	if avatar != "" {
		user.Avatar = avatar
	}
	if timezone != "" {
		user.Timezone = timezone
	}
	if language != "" {
		user.Language = language
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}
