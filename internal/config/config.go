package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
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
	// Speed multiplies plant growth: 1 is real time, 60 grows an hour's worth
	// each minute. It changes pacing only, never the rules.
	Speed float64
}

type APIConfig struct {
	CORSOrigin        string
	RateLimitRequests int
	RateLimitWindow   time.Duration
	// TrustedProxies lists the proxy IPs or CIDRs whose X-Forwarded-For
	// header names the client. Empty trusts none, so the client is the peer.
	TrustedProxies []string
}

// Development defaults that are unsafe anywhere else.
const (
	developmentJWTSecret  = "your-super-secret-jwt-key-change-in-production"
	developmentDBPassword = "password"
	minJWTSecretLength    = 32
)

var sslModes = []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't return error if .env doesn't exist
		fmt.Println("No .env file found, using environment variables")
	}

	e := &env{}
	environment := e.get("ENV", "development")
	// Outside development, an unset TLS mode must not silently mean plaintext.
	defaultSSLMode := "require"
	if environment == "development" {
		defaultSSLMode = "disable"
	}
	config := &Config{
		Server: ServerConfig{
			Port: e.get("PORT", "8080"),
			Env:  environment,
		},
		Database: DatabaseConfig{
			Host:     e.get("DB_HOST", "localhost"),
			Port:     e.get("DB_PORT", "5432"),
			User:     e.get("DB_USER", "postgres"),
			Password: e.get("DB_PASSWORD", developmentDBPassword),
			Name:     e.get("DB_NAME", "my_garden"),
			SSLMode:  e.get("DB_SSL_MODE", defaultSSLMode),
		},
		Redis: RedisConfig{
			Host:     e.get("REDIS_HOST", "localhost"),
			Port:     e.get("REDIS_PORT", "6379"),
			Password: e.get("REDIS_PASSWORD", ""),
			DB:       e.getInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret: e.get("JWT_SECRET", developmentJWTSecret),
			Expiry: e.getDuration("JWT_EXPIRY", 24*time.Hour),
		},
		Game: GameConfig{
			TickInterval:          e.getDuration("GAME_TICK_INTERVAL", 5*time.Minute),
			WeatherUpdateInterval: e.getDuration("WEATHER_UPDATE_INTERVAL", 10*time.Minute),
			Speed:                 e.getFloat("GAME_SPEED", 1),
		},
		API: APIConfig{
			CORSOrigin:        e.get("CORS_ORIGIN", "http://localhost:5173"),
			RateLimitRequests: e.getInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   e.getDuration("RATE_LIMIT_WINDOW", time.Minute),
			TrustedProxies:    e.getList("TRUSTED_PROXIES"),
		},
	}

	if err := errors.Join(append(e.errs, config.validate()...)...); err != nil {
		return nil, fmt.Errorf("invalid configuration:\n%w", err)
	}
	return config, nil
}

// validate rejects values the API cannot run with, and, outside development,
// the development defaults that would make a deployment unsafe.
func (c *Config) validate() []error {
	var errs []error
	if !slices.Contains(sslModes, c.Database.SSLMode) {
		errs = append(errs, fmt.Errorf("DB_SSL_MODE=%q is not one of %s", c.Database.SSLMode, strings.Join(sslModes, ", ")))
	}
	if c.Game.Speed <= 0 {
		errs = append(errs, errors.New("GAME_SPEED must be positive"))
	}
	if c.JWT.Expiry <= 0 {
		errs = append(errs, errors.New("JWT_EXPIRY must be positive"))
	}
	if c.API.RateLimitRequests < 1 || c.API.RateLimitWindow <= 0 {
		errs = append(errs, errors.New("RATE_LIMIT_REQUESTS and RATE_LIMIT_WINDOW must be positive"))
	}
	for _, proxy := range c.API.TrustedProxies {
		if _, _, err := net.ParseCIDR(proxy); err != nil && net.ParseIP(proxy) == nil {
			errs = append(errs, fmt.Errorf("TRUSTED_PROXIES entry %q is not an IP or CIDR", proxy))
		}
	}

	if c.Server.Env == "development" {
		return errs
	}
	if c.JWT.Secret == developmentJWTSecret || len(c.JWT.Secret) < minJWTSecretLength {
		errs = append(errs, fmt.Errorf("JWT_SECRET must be a random value of at least %d characters when ENV=%q", minJWTSecretLength, c.Server.Env))
	}
	if c.Database.Password == developmentDBPassword {
		errs = append(errs, fmt.Errorf("DB_PASSWORD must not be the development default when ENV=%q", c.Server.Env))
	}
	return errs
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

// getList splits a comma-separated value, ignoring blank entries.
func (e *env) getList(key string) []string {
	var values []string
	for value := range strings.SplitSeq(os.Getenv(key), ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
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

func (e *env) getFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s=%q is not a number", key, value))
		return defaultValue
	}
	return floatValue
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
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
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
