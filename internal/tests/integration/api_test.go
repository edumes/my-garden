package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/audit"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/store"
	jwtauth "github.com/my-garden/api/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestServer struct {
	Router        *gin.Engine
	AuthHandler   *auth.AuthHandler
	GardenHandler *garden.GardenHandler
	StoreHandler  *store.StoreHandler
	DB            *database.Database
}

func setupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	// Setup database
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
			User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
			Password: getEnvOrDefault("TEST_DB_PASS", "postgres"),
			Name:     "my_garden_test",
		},
		JWT: config.JWTConfig{
			Secret: "test_secret",
			Expiry: 24 * time.Hour,
		},
	}

	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)

	// Setup audit service
	auditService := audit.NewAuditService(db)

	// Setup services
	jwtManager := jwtauth.NewJWTManager(cfg)
	authRepo := auth.NewRepository(db.DB)
	authService := auth.NewService(authRepo, jwtManager)
	gardenRepo := garden.NewRepository(db.DB)
	gardenService := garden.NewService(gardenRepo)
	storeRepo := store.NewRepository(db)
	storeService := store.NewService(storeRepo, auditService)

	// Setup handlers
	authHandler := auth.NewAuthHandler(authService, auditService)
	gardenHandler := garden.NewGardenHandler(gardenService, auditService, authService, nil)
	storeHandler := store.NewStoreHandler(storeService, auditService)

	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	router.POST("/auth/register", authHandler.Register)
	router.POST("/auth/login", authHandler.Login)

	// Protected routes
	protected := router.Group("/")
	protected.Use(auth.AuthMiddleware(jwtManager))
	{
		protected.GET("/gardens", gardenHandler.GetGardens)
		protected.POST("/gardens", gardenHandler.CreateGarden)
		protected.GET("/gardens/:id", gardenHandler.GetGarden)
		protected.PUT("/gardens/:id", gardenHandler.UpdateGarden)
		protected.DELETE("/gardens/:id", gardenHandler.DeleteGarden)
		protected.POST("/gardens/:id/plants", gardenHandler.PlantSeed)

		protected.GET("/store/inventory", storeHandler.GetStoreInventory)
		protected.POST("/store/buy", storeHandler.BuySeed)
	}

	return &TestServer{
		Router:        router,
		AuthHandler:   authHandler,
		GardenHandler: gardenHandler,
		StoreHandler:  storeHandler,
		DB:            db,
	}
}

func TestAuthFlow(t *testing.T) {
	server := setupTestServer(t)

	// Test registration
	registerPayload := auth.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	w := httptest.NewRecorder()
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResponse auth.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, registerResponse.Token)
	assert.Equal(t, registerPayload.Username, registerResponse.User.Username)

	// Test login
	loginPayload := auth.LoginRequest{
		Username: registerPayload.Username,
		Password: registerPayload.Password,
	}

	w = httptest.NewRecorder()
	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var loginResponse auth.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResponse.Token)
	assert.Equal(t, registerPayload.Username, loginResponse.User.Username)
}

func TestGardenFlow(t *testing.T) {
	server := setupTestServer(t)

	// First register and login a user
	user, token := createTestUser(t, server)

	// Test garden creation
	createGardenPayload := map[string]interface{}{
		"name":        "My Test Garden",
		"description": "A beautiful test garden",
	}

	w := httptest.NewRecorder()
	body, _ := json.Marshal(createGardenPayload)
	req := httptest.NewRequest("POST", "/gardens", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var gardenResponse map[string]garden.Garden
	err := json.Unmarshal(w.Body.Bytes(), &gardenResponse)
	require.NoError(t, err)

	createdGarden := gardenResponse["garden"]
	assert.Equal(t, createGardenPayload["name"], createdGarden.Name)
	assert.Equal(t, createGardenPayload["description"], createdGarden.Description)
	assert.Equal(t, user.ID, createdGarden.UserID)

	// Test get garden
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", fmt.Sprintf("/gardens/%s", createdGarden.ID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var getGardenResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &getGardenResponse)
	require.NoError(t, err)
}

func TestStoreFlow(t *testing.T) {
	server := setupTestServer(t)

	// First register and login a user
	_, token := createTestUser(t, server)

	// Create a test plant type
	plantType := &garden.PlantType{
		Name:        "Test Plant",
		Description: "A test plant",
		GrowthTime:  60,
		SeedPrice:   10,
	}
	err := server.DB.DB.Create(plantType).Error
	require.NoError(t, err)

	// Test getting store inventory
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/store/inventory", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var inventoryResponse map[string][]garden.PlantType
	err = json.Unmarshal(w.Body.Bytes(), &inventoryResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, inventoryResponse["inventory"])

	// Test buying seeds
	buySeedPayload := map[string]interface{}{
		"plant_type_id": plantType.ID,
		"quantity":      1,
	}

	w = httptest.NewRecorder()
	body, _ := json.Marshal(buySeedPayload)
	req = httptest.NewRequest("POST", "/store/buy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	server.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Helper function to create a test user and return the user and token
func createTestUser(t *testing.T, server *TestServer) (*auth.User, string) {
	registerPayload := auth.RegisterRequest{
		Username:  fmt.Sprintf("testuser_%d", time.Now().UnixNano()),
		Email:     fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}

	w := httptest.NewRecorder()
	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	server.Router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response auth.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	return &response.User, response.Token
}
