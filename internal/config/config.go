package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Game     GameConfig
	API      APIConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type GameConfig struct {
	TickInterval          time.Duration
	WeatherUpdateInterval time.Duration
	PlantGrowthInterval   time.Duration
}

type APIConfig struct {
	CORSOrigin        string
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't return error if .env doesn't exist
		fmt.Println("No .env file found, using environment variables")
	}

	e := &env{}
	config := &Config{
		Server: ServerConfig{
			Port: e.get("PORT", "8080"),
			Env:  e.get("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     e.get("DB_HOST", "localhost"),
			Port:     e.get("DB_PORT", "5432"),
			User:     e.get("DB_USER", "postgres"),
			Password: e.get("DB_PASSWORD", "password"),
			Name:     e.get("DB_NAME", "my_garden"),
			SSLMode:  e.get("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     e.get("REDIS_HOST", "localhost"),
			Port:     e.get("REDIS_PORT", "6379"),
			Password: e.get("REDIS_PASSWORD", ""),
			DB:       e.getInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: e.get("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			Expiry: e.getDuration("JWT_EXPIRY", 24*time.Hour),
		},
		Game: GameConfig{
			TickInterval:          e.getDuration("GAME_TICK_INTERVAL", 5*time.Minute),
			WeatherUpdateInterval: e.getDuration("WEATHER_UPDATE_INTERVAL", 10*time.Minute),
			PlantGrowthInterval:   e.getDuration("PLANT_GROWTH_INTERVAL", 15*time.Minute),
		},
		API: APIConfig{
			CORSOrigin:        e.get("CORS_ORIGIN", "http://localhost:5173"),
			RateLimitRequests: e.getInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   e.getDuration("RATE_LIMIT_WINDOW", time.Minute),
		},
	}

	if err := errors.Join(e.errs...); err != nil {
		return nil, fmt.Errorf("invalid configuration:\n%w", err)
	}
	return config, nil
}

// env reads environment variables and collects every malformed value, so
// startup reports all configuration mistakes at once instead of silently
// falling back to defaults.
type env struct {
	errs []error
}

func (e *env) get(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (e *env) getInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s=%q is not an integer", key, value))
		return defaultValue
	}
	return intValue
}

func (e *env) getDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s=%q is not a duration (examples: 30s, 5m, 24h)", key, value))
		return defaultValue
	}
	return duration
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}
