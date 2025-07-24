package unit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/garden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGardenRepository is a mock implementation of garden repository
type MockGardenRepository struct {
	mock.Mock
}

func (m *MockGardenRepository) CreateGarden(garden *garden.Garden) error {
	args := m.Called(garden)
	return args.Error(0)
}

func (m *MockGardenRepository) GetGarden(id uuid.UUID) (*garden.Garden, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*garden.Garden), args.Error(1)
}

func (m *MockGardenRepository) GetUserGardens(userID uuid.UUID) ([]garden.Garden, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]garden.Garden), args.Error(1)
}

func (m *MockGardenRepository) PlantSeed(plant *garden.Plant) error {
	args := m.Called(plant)
	return args.Error(0)
}

func (m *MockGardenRepository) GetPlant(id uuid.UUID) (*garden.Plant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*garden.Plant), args.Error(1)
}

func (m *MockGardenRepository) HarvestPlant(plant *garden.Plant) error {
	args := m.Called(plant)
	return args.Error(0)
}

func TestCreateGarden(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := garden.NewService(mockRepo)

	tests := []struct {
		name          string
		request       garden.CreateGardenRequest
		userID        uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful garden creation",
			request: garden.CreateGardenRequest{
				Name:        "Test Garden",
				Description: "A test garden",
			},
			userID: uuid.New(),
			mockSetup: func() {
				mockRepo.On("CreateGarden", mock.AnythingOfType("*garden.Garden")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			request: garden.CreateGardenRequest{
				Name:        "Test Garden",
				Description: "A test garden",
			},
			userID: uuid.New(),
			mockSetup: func() {
				mockRepo.On("CreateGarden", mock.AnythingOfType("*garden.Garden")).Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			_, err := service.CreateGarden(tt.userID, tt.request)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestPlantSeed(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := garden.NewService(mockRepo)

	testGarden := &garden.Garden{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Name:        "Test Garden",
		Description: "A test garden",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Plants:      []garden.Plant{},
	}

	tests := []struct {
		name          string
		gardenID      uuid.UUID
		request       garden.PlantRequest
		mockSetup     func()
		expectedError error
	}{
		{
			name:     "successful planting",
			gardenID: testGarden.ID,
			request: garden.PlantRequest{
				PlantTypeID: uuid.New(),
				Position:    new(int),
			},
			mockSetup: func() {
				mockRepo.On("GetGarden", testGarden.ID).Return(testGarden, nil)
				mockRepo.On("PlantSeed", mock.AnythingOfType("*garden.Plant")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:     "garden not found",
			gardenID: uuid.New(),
			request: garden.PlantRequest{
				PlantTypeID: uuid.New(),
				Position:    new(int),
			},
			mockSetup: func() {
				mockRepo.On("GetGarden", mock.AnythingOfType("uuid.UUID")).Return(nil, garden.ErrGardenNotFound)
			},
			expectedError: garden.ErrGardenNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			_, err := service.PlantSeed(tt.gardenID, tt.request)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestHarvestPlant(t *testing.T) {
	mockRepo := new(MockGardenRepository)
	service := garden.NewService(mockRepo)

	plantID := uuid.New()
	testPlant := &garden.Plant{
		ID:             plantID,
		GardenID:       uuid.New(),
		PlantTypeID:    uuid.New(),
		Position:       0,
		Stage:          garden.PlantStageMature,
		GrowthProgress: 100,
		PlantedAt:      time.Now().Add(-24 * time.Hour),
		CreatedAt:      time.Now().Add(-24 * time.Hour),
		UpdatedAt:      time.Now(),
	}

	tests := []struct {
		name          string
		plantID       uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name:    "successful harvest",
			plantID: plantID,
			mockSetup: func() {
				mockRepo.On("GetPlant", plantID).Return(testPlant, nil)
				mockRepo.On("HarvestPlant", mock.AnythingOfType("*garden.Plant")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:    "plant not found",
			plantID: uuid.New(),
			mockSetup: func() {
				mockRepo.On("GetPlant", mock.AnythingOfType("uuid.UUID")).Return(nil, garden.ErrPlantNotFound)
			},
			expectedError: garden.ErrPlantNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			_, err := service.HarvestPlant(tt.plantID)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
