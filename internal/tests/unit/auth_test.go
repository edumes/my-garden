package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/types"
	jwtauth "github.com/my-garden/api/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// MockAuthRepository is a mock implementation of auth.Repository
type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) CreateUser(user *auth.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockAuthRepository) FindUserByUsername(username string) (*auth.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthRepository) FindUserByEmail(email string) (*auth.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthRepository) FindUserByID(id uuid.UUID) (*auth.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthRepository) UpdateUser(user *auth.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test_secret",
			Expiry: 24 * time.Hour,
		},
	}
	jwtManager := jwtauth.NewJWTManager(cfg)
	service := auth.NewService(mockRepo, jwtManager)

	tests := []struct {
		name          string
		request       auth.RegisterRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful registration",
			request: auth.RegisterRequest{
				Username:  "testuser",
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup: func() {
				mockRepo.On("FindUserByUsername", "testuser").Return(nil, auth.ErrUserNotFound)
				mockRepo.On("FindUserByEmail", "test@example.com").Return(nil, auth.ErrUserNotFound)
				mockRepo.On("CreateUser", mock.AnythingOfType("*auth.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "username taken",
			request: auth.RegisterRequest{
				Username:  "existinguser",
				Email:     "new@example.com",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup: func() {
				mockRepo.On("FindUserByUsername", "existinguser").Return(&auth.User{}, nil)
			},
			expectedError: auth.ErrUsernameTaken,
		},
		{
			name: "email taken",
			request: auth.RegisterRequest{
				Username:  "newuser",
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup: func() {
				mockRepo.On("FindUserByUsername", "newuser").Return(nil, auth.ErrUserNotFound)
				mockRepo.On("FindUserByEmail", "existing@example.com").Return(&auth.User{}, nil)
			},
			expectedError: auth.ErrEmailTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			_, err := service.Register(tt.request)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockAuthRepository)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test_secret",
			Expiry: 24 * time.Hour,
		},
	}
	jwtManager := jwtauth.NewJWTManager(cfg)
	service := auth.NewService(mockRepo, jwtManager)

	// Create a proper bcrypt hash of the test password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	testUser := &auth.User{
		User: types.User{
			ID:        uuid.New(),
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
		},
		PasswordHash: string(hashedPassword),
		LastLoginAt:  &time.Time{},
	}

	tests := []struct {
		name          string
		request       auth.LoginRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful login",
			request: auth.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.On("FindUserByUsername", "testuser").Return(testUser, nil)
				mockRepo.On("UpdateUser", mock.AnythingOfType("*auth.User")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			request: auth.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			mockSetup: func() {
				mockRepo.On("FindUserByUsername", "nonexistent").Return(nil, auth.ErrUserNotFound)
			},
			expectedError: auth.ErrUserNotFound,
		},
		{
			name: "invalid password",
			request: auth.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockRepo.On("FindUserByUsername", "testuser").Return(testUser, nil)
			},
			expectedError: auth.ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			_, err := service.Login(tt.request)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
