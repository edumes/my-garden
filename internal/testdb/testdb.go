// Package testdb gives each integration test its own PostgreSQL database.
package testdb

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/migrations"
	"gorm.io/gorm/logger"
)

// New creates a randomly named database on the PostgreSQL server described by
// the DB_* variables (falling back to the repository's .env), applies the
// migrations and seed data, and drops the database when the test ends.
// Variables already set in the environment win over .env.
func New(tb testing.TB) (*database.Database, *config.Config) {
	tb.Helper()

	_, file, _, _ := runtime.Caller(0)
	_ = godotenv.Load(filepath.Join(filepath.Dir(file), "..", "..", ".env"))

	cfg, err := config.Load()
	if err != nil {
		tb.Fatal(err)
	}

	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		tb.Fatal(err)
	}
	name := "my_garden_test_" + hex.EncodeToString(suffix)

	admin := connect(tb, cfg, "postgres")
	if err := admin.DB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, name)).Error; err != nil {
		tb.Fatalf("create test database: %v", err)
	}

	testCfg := *cfg
	testCfg.Database.Name = name
	db := connect(tb, &testCfg, name)

	tb.Cleanup(func() {
		if sqlDB, err := db.DB.DB(); err == nil {
			sqlDB.Close()
		}
		if err := admin.DB.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name)).Error; err != nil {
			tb.Errorf("drop test database %s: %v", name, err)
		}
		if sqlDB, err := admin.DB.DB(); err == nil {
			sqlDB.Close()
		}
	})

	if err := migrations.AutoMigrate(db.DB); err != nil {
		tb.Fatalf("migrate test database: %v", err)
	}
	if err := migrations.Seed(db.DB); err != nil {
		tb.Fatalf("seed test database: %v", err)
	}

	return db, &testCfg
}

func connect(tb testing.TB, cfg *config.Config, name string) *database.Database {
	tb.Helper()

	dbCfg := *cfg
	dbCfg.Database.Name = name
	db, err := database.NewDatabase(&dbCfg)
	if err != nil {
		tb.Fatal(err)
	}
	db.DB.Logger = logger.Discard
	return db
}
