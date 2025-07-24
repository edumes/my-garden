package tests

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/my-garden/api/internal/auth"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/garden"
	"github.com/my-garden/api/internal/store"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestDB represents a test database instance
type TestDB struct {
	DB         *gorm.DB
	AuthRepo   auth.Repository
	GardenRepo garden.Repository
	StoreRepo  store.Repository
}

// NewTestDB creates a new test database instance
func NewTestDB() *TestDB {
	// Use environment variables or default to test database
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	dbPass := getEnvOrDefault("TEST_DB_PASS", "postgres")
	dbName := getEnvOrDefault("TEST_DB_NAME", "my_garden_test")

	config := &config.Config{
		Database: config.DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPass,
			Name:     dbName,
		},
	}

	db, err := database.NewDatabase(config)
	if err != nil {
		panic(err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		panic(err)
	}

	return &TestDB{
		DB:         db,
		AuthRepo:   auth.NewRepository(db),
		GardenRepo: garden.NewRepository(db),
		StoreRepo:  store.NewRepository(db),
	}
}

// CleanupTestDB cleans up the test database
func (db *TestDB) CleanupTestDB(t *testing.T) {
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)

	// Drop all tables
	tables := []string{
		"users",
		"gardens",
		"plants",
		"plant_types",
		"garden_shares",
		"store_inventory",
		"audit_logs",
	}

	for _, table := range tables {
		_, err := sqlDB.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
		require.NoError(t, err)
	}
}

// CreateTestUser creates a test user
func (db *TestDB) CreateTestUser(t *testing.T, username, email, password string) *auth.User {
	user := &auth.User{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: "Test",
		LastName:  "User",
	}

	err := db.AuthRepo.CreateUser(user)
	require.NoError(t, err)
	return user
}

// CreateTestGarden creates a test garden
func (db *TestDB) CreateTestGarden(t *testing.T, userID uuid.UUID, name, description string) *garden.Garden {
	garden := &garden.Garden{
		UserID:      userID,
		Name:        name,
		Description: description,
	}

	err := db.GardenRepo.CreateGarden(garden)
	require.NoError(t, err)
	return garden
}

// CreateTestPlantType creates a test plant type
func (db *TestDB) CreateTestPlantType(t *testing.T, name string, growthTime int, price int) *store.PlantType {
	plantType := &store.PlantType{
		Name:       name,
		GrowthTime: growthTime,
		Price:      price,
	}

	err := db.StoreRepo.CreatePlantType(plantType)
	require.NoError(t, err)
	return plantType
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TruncateAllTables truncates all tables in the test database
func (db *TestDB) TruncateAllTables(t *testing.T) {
	sqlDB, err := db.DB.DB()
	require.NoError(t, err)

	tables := []string{
		"users",
		"gardens",
		"plants",
		"plant_types",
		"garden_shares",
		"store_inventory",
		"audit_logs",
	}

	for _, table := range tables {
		_, err := sqlDB.Exec("TRUNCATE TABLE " + table + " CASCADE")
		require.NoError(t, err)
	}
}

// WithTestTransaction runs a test within a transaction and rolls it back afterward
func (db *TestDB) WithTestTransaction(t *testing.T, fn func(tx *gorm.DB)) {
	tx := db.DB.Begin()
	require.NoError(t, tx.Error)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	fn(tx)

	tx.Rollback()
}

// WithTestDatabase creates a temporary test database, runs the test, and drops it afterward
func WithTestDatabase(t *testing.T, fn func(*TestDB)) {
	db := NewTestDB()
	defer db.CleanupTestDB(t)
	fn(db)
}
