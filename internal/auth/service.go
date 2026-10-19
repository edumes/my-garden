package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUsernameTaken   = errors.New("username already taken")
	ErrEmailTaken      = errors.New("email already taken")
	ErrTokenGeneration = errors.New("failed to generate token")
)

type Service interface {
	Register(req RegisterRequest) (*AuthResponse, error)
	Login(req LoginRequest) (*AuthResponse, error)
	RefreshToken(userID uuid.UUID) (string, time.Time, error)
	Logout(userID uuid.UUID) error
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
		Username:     req.Username,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: string(hashedPassword),
		LastLoginAt:  &now,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}
	return s.issue(user)
}

// issue returns a token for the user's current token version.
func (s *service) issue(user *User) (*AuthResponse, error) {
	token, expiresAt, err := s.jwtManager.GenerateToken(user.ID, user.Username, user.Email, user.TokenVersion)
	if err != nil {
		return nil, ErrTokenGeneration
	}
	return &AuthResponse{Token: token, User: *user, ExpiresAt: expiresAt}, nil
}

func (s *service) Login(req LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindUserByUsername(req.Username)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.repo.RecordLogin(user.ID, now); err != nil {
		return nil, err
	}
	return s.issue(user)
}

// RefreshToken issues a fresh token to an authenticated user. The caller's
// token was already checked against the current token version.
func (s *service) RefreshToken(userID uuid.UUID) (string, time.Time, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return "", time.Time{}, ErrUserNotFound
	}
	response, err := s.issue(user)
	if err != nil {
		return "", time.Time{}, err
	}
	return response.Token, response.ExpiresAt, nil
}

// Logout revokes every token issued to the user, on every device.
func (s *service) Logout(userID uuid.UUID) error {
	return s.repo.RevokeTokens(userID)
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

	if err := s.repo.UpdateProfile(user); err != nil {
		return nil, err
	}

	return user, nil
}
