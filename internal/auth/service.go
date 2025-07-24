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
	AddExperience(userID uuid.UUID, xp int) (*User, error)
	GetLevelProgress(experience int) float64
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

// XP thresholds for each level
var levelThresholds = []int{
	0,     // Level 1
	100,   // Level 2
	300,   // Level 3
	600,   // Level 4
	1000,  // Level 5
	1500,  // Level 6
	2100,  // Level 7
	2800,  // Level 8
	3600,  // Level 9
	4500,  // Level 10
	5500,  // Level 11
	6600,  // Level 12
	7800,  // Level 13
	9100,  // Level 14
	10500, // Level 15
}

// XP rewards for different actions
const (
	XPPlantSeed    = 10
	XPHarvestPlant = 25
	XPCreateGarden = 50
	XPShareGarden  = 15
	XPJoinGarden   = 20
	XPDailyLogin   = 5
	XPFirstHarvest = 100 // Bonus for first harvest of each plant type
	XPGardenMaster = 200 // Bonus for harvesting all plant types
)

// AddExperience adds XP to a user and handles level ups
func (s *service) AddExperience(userID uuid.UUID, xp int) (*User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	oldLevel := user.Level
	user.Experience += xp

	// Calculate new level
	newLevel := 1
	for i, threshold := range levelThresholds {
		if user.Experience >= threshold {
			newLevel = i + 1
		} else {
			break
		}
	}

	user.Level = newLevel

	// Add coins bonus for level up
	if newLevel > oldLevel {
		coinsBonus := (newLevel - oldLevel) * 100
		user.Coins += coinsBonus
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetLevelProgress returns the current level progress as a percentage
func (s *service) GetLevelProgress(experience int) float64 {
	currentLevel := 1
	for i, threshold := range levelThresholds {
		if experience >= threshold {
			currentLevel = i + 1
		} else {
			break
		}
	}

	if currentLevel >= len(levelThresholds) {
		return 100.0
	}

	currentThreshold := levelThresholds[currentLevel-1]
	nextThreshold := levelThresholds[currentLevel]

	progress := float64(experience-currentThreshold) / float64(nextThreshold-currentThreshold) * 100
	if progress > 100 {
		progress = 100
	}

	return progress
}
