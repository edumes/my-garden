package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type E2ETestSuite struct {
	APIClient  *http.Client
	BaseURL    string
	AuthToken  string
	UserID     uuid.UUID
	TestUserID string
}

func setupE2ETest(t *testing.T) *E2ETestSuite {
	baseURL := "http://localhost:8080" // Change this to match your test environment
	testUserID := fmt.Sprintf("test_%d", time.Now().UnixNano())

	suite := &E2ETestSuite{
		APIClient:  &http.Client{Timeout: 10 * time.Second},
		BaseURL:    baseURL,
		TestUserID: testUserID,
	}

	// Register test user
	registerResp, err := suite.registerUser(testUserID)
	require.NoError(t, err)
	suite.AuthToken = registerResp.Token
	suite.UserID = registerResp.User.ID

	return suite
}

func TestCompleteGardenFlow(t *testing.T) {
	suite := setupE2ETest(t)

	// 1. Create a new garden
	garden, err := suite.createGarden("My First Garden", "A beautiful garden to grow vegetables")
	require.NoError(t, err)
	assert.NotNil(t, garden)
	assert.Equal(t, "My First Garden", garden.Name)

	// 2. Get store inventory
	inventory, err := suite.getStoreInventory()
	require.NoError(t, err)
	require.NotEmpty(t, inventory)

	// 3. Buy seeds
	seedPurchase, err := suite.buySeeds(inventory[0].ID, 3)
	require.NoError(t, err)
	assert.Equal(t, 3, seedPurchase.Quantity)

	// 4. Plant seeds in garden
	plant, err := suite.plantSeed(garden.ID, inventory[0].ID, 0)
	require.NoError(t, err)
	assert.NotNil(t, plant)
	assert.Equal(t, garden.PlantStageSeed, plant.Stage)

	// 5. Wait for plant growth (simulated)
	time.Sleep(2 * time.Second)

	// 6. Check garden status
	updatedGarden, err := suite.getGarden(garden.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedGarden)
	assert.NotEmpty(t, updatedGarden.Plants)

	// 7. Harvest plant when ready
	if updatedGarden.Plants[0].Stage == garden.PlantStageMature {
		harvestedPlant, err := suite.harvestPlant(garden.ID, updatedGarden.Plants[0].ID)
		require.NoError(t, err)
		assert.NotNil(t, harvestedPlant)
		assert.NotNil(t, harvestedPlant.HarvestedAt)
	}

	// 8. Share garden
	shareLink, err := suite.createGardenShareLink(garden.ID, []garden.GardenPermission{garden.GardenPermissionView})
	require.NoError(t, err)
	assert.NotEmpty(t, shareLink)

	// 9. Join shared garden (with another user)
	otherUser, err := suite.registerUser(fmt.Sprintf("other_%d", time.Now().UnixNano()))
	require.NoError(t, err)

	// Switch to other user's context
	originalToken := suite.AuthToken
	suite.AuthToken = otherUser.Token

	joinedGarden, err := suite.joinGarden(shareLink)
	require.NoError(t, err)
	assert.Equal(t, garden.ID, joinedGarden.ID)

	// Switch back to original user
	suite.AuthToken = originalToken

	// 10. Delete garden
	err = suite.deleteGarden(garden.ID)
	require.NoError(t, err)
}

// Helper methods for API calls

func (s *E2ETestSuite) registerUser(userID string) (*auth.AuthResponse, error) {
	url := fmt.Sprintf("%s/auth/register", s.BaseURL)
	payload := auth.RegisterRequest{
		Username:  fmt.Sprintf("user_%s", userID),
		Email:     fmt.Sprintf("user_%s@example.com", userID),
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	var response auth.AuthResponse
	err := s.makeRequest(http.MethodPost, url, payload, &response)
	return &response, err
}

func (s *E2ETestSuite) createGarden(name, description string) (*garden.Garden, error) {
	url := fmt.Sprintf("%s/gardens", s.BaseURL)
	payload := garden.CreateGardenRequest{
		Name:        name,
		Description: description,
	}

	var response map[string]garden.Garden
	err := s.makeRequest(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, err
	}
	return &response["garden"], nil
}

func (s *E2ETestSuite) getStoreInventory() ([]store.PlantType, error) {
	url := fmt.Sprintf("%s/store/inventory", s.BaseURL)
	var response map[string][]store.PlantType
	err := s.makeRequest(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, err
	}
	return response["inventory"], nil
}

func (s *E2ETestSuite) buySeeds(plantTypeID uuid.UUID, quantity int) (*store.BuySeedResponse, error) {
	url := fmt.Sprintf("%s/store/buy", s.BaseURL)
	payload := store.BuySeedRequest{
		PlantTypeID: plantTypeID,
		Quantity:    quantity,
	}

	var response store.BuySeedResponse
	err := s.makeRequest(http.MethodPost, url, payload, &response)
	return &response, err
}

func (s *E2ETestSuite) plantSeed(gardenID, plantTypeID uuid.UUID, position int) (*garden.Plant, error) {
	url := fmt.Sprintf("%s/gardens/%s/plants", s.BaseURL, gardenID)
	payload := garden.PlantRequest{
		PlantTypeID: plantTypeID,
		Position:    &position,
	}

	var response map[string]garden.Plant
	err := s.makeRequest(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, err
	}
	return &response["plant"], nil
}

func (s *E2ETestSuite) getGarden(gardenID uuid.UUID) (*garden.Garden, error) {
	url := fmt.Sprintf("%s/gardens/%s", s.BaseURL, gardenID)
	var response map[string]garden.Garden
	err := s.makeRequest(http.MethodGet, url, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response["garden"], nil
}

func (s *E2ETestSuite) harvestPlant(gardenID, plantID uuid.UUID) (*garden.Plant, error) {
	url := fmt.Sprintf("%s/gardens/%s/plants/%s/harvest", s.BaseURL, gardenID, plantID)
	var response map[string]garden.Plant
	err := s.makeRequest(http.MethodPost, url, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response["plant"], nil
}

func (s *E2ETestSuite) createGardenShareLink(gardenID uuid.UUID, permissions []garden.GardenPermission) (string, error) {
	url := fmt.Sprintf("%s/garden-shares/access-links", s.BaseURL)
	payload := garden.CreateAccessLinkRequest{
		GardenID:    gardenID,
		Permissions: permissions,
	}

	var response map[string]string
	err := s.makeRequest(http.MethodPost, url, payload, &response)
	if err != nil {
		return "", err
	}
	return response["token"], nil
}

func (s *E2ETestSuite) joinGarden(token string) (*garden.Garden, error) {
	url := fmt.Sprintf("%s/garden-shares/join", s.BaseURL)
	payload := garden.JoinGardenRequest{
		Token: token,
	}

	var response map[string]garden.Garden
	err := s.makeRequest(http.MethodPost, url, payload, &response)
	if err != nil {
		return nil, err
	}
	return &response["garden"], nil
}

func (s *E2ETestSuite) deleteGarden(gardenID uuid.UUID) error {
	url := fmt.Sprintf("%s/gardens/%s", s.BaseURL, gardenID)
	return s.makeRequest(http.MethodDelete, url, nil, nil)
}

// Generic request helper
func (s *E2ETestSuite) makeRequest(method, url string, payload interface{}, response interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if s.AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.AuthToken))
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
		// Add payload to request body
	}

	resp, err := s.APIClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
