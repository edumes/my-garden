package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/my-garden/api/internal/config"
	"github.com/my-garden/api/pkg/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

var _ db.Database = (*Database)(nil)

func NewDatabase(cfg *config.Config) (*Database, error) {
	gormConfig := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	}

	gormDB, err := gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL at %s:%s (database %q, user %q); check the DB_* settings or start local services with `docker compose up -d --wait`: %w",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.User, err)
	}

	return &Database{DB: gormDB}, nil
}
