package integration

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/migrations"
)

func TestMain(m *testing.M) {
	// Setup test database
	if err := setupTestDB(); err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup test database
	if err := cleanupTestDB(); err != nil {
		log.Printf("Failed to cleanup test database: %v", err)
	}

	os.Exit(code)
}

func setupTestDB() error {
	// Connect to default postgres database to create test database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		getEnvOrDefault("TEST_DB_HOST", "localhost"),
		getEnvOrDefault("TEST_DB_PORT", "5432"),
		getEnvOrDefault("TEST_DB_USER", "postgres"),
		getEnvOrDefault("TEST_DB_PASS", "postgres"),
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Drop test database if it exists
	_, err = db.Exec(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = 'my_garden_test'
		AND pid <> pg_backend_pid();
	`)
	if err != nil {
		log.Printf("Warning: Failed to terminate existing connections: %v", err)
	}

	_, err = db.Exec("DROP DATABASE IF EXISTS my_garden_test")
	if err != nil {
		return fmt.Errorf("failed to drop test database: %w", err)
	}

	// Create test database
	_, err = db.Exec("CREATE DATABASE my_garden_test")
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	// Connect to test database and run migrations
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
			Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
			User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
			Password: getEnvOrDefault("TEST_DB_PASS", "postgres"),
			Name:     "my_garden_test",
		},
	}

	testDB, err := database.NewDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	if err := migrations.AutoMigrate(testDB.DB); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func cleanupTestDB() error {
	// Connect to default postgres database to drop test database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		getEnvOrDefault("TEST_DB_HOST", "localhost"),
		getEnvOrDefault("TEST_DB_PORT", "5432"),
		getEnvOrDefault("TEST_DB_USER", "postgres"),
		getEnvOrDefault("TEST_DB_PASS", "postgres"),
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Terminate all connections to the test database
	_, err = db.Exec(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = 'my_garden_test'
		AND pid <> pg_backend_pid();
	`)
	if err != nil {
		log.Printf("Warning: Failed to terminate existing connections: %v", err)
	}

	// Drop test database
	_, err = db.Exec("DROP DATABASE IF EXISTS my_garden_test")
	if err != nil {
		return fmt.Errorf("failed to drop test database: %w", err)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
