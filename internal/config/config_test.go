package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadRejectsMalformedValues(t *testing.T) {
	t.Setenv("REDIS_DB", "zero")
	t.Setenv("JWT_EXPIRY", "1 day")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for malformed values")
	}
	for _, key := range []string{"REDIS_DB", "JWT_EXPIRY"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not mention %s", err, key)
		}
	}
}

func TestLoadParsesValidValues(t *testing.T) {
	t.Setenv("REDIS_DB", "2")
	t.Setenv("JWT_EXPIRY", "1h")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Redis.DB != 2 || cfg.JWT.Expiry != time.Hour {
		t.Errorf("got REDIS_DB=%d JWT_EXPIRY=%s", cfg.Redis.DB, cfg.JWT.Expiry)
	}
}

// production sets a deployment environment with safe secrets.
func production(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("JWT_SECRET", strings.Repeat("s", minJWTSecretLength))
	t.Setenv("DB_PASSWORD", "a-real-password")
	t.Setenv("DB_SSL_MODE", "")
}

func TestLoadRejectsDevelopmentDefaultsElsewhere(t *testing.T) {
	production(t)
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DB_PASSWORD", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected development defaults to be rejected")
	}
	for _, key := range []string{"JWT_SECRET", "DB_PASSWORD"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not mention %s", err, key)
		}
	}

	t.Setenv("JWT_SECRET", "short")
	t.Setenv("DB_PASSWORD", "a-real-password")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("expected a short JWT_SECRET to be rejected, got %v", err)
	}
}

func TestLoadDefaultsToDatabaseTLSElsewhere(t *testing.T) {
	production(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("got DB_SSL_MODE=%q, want require", cfg.Database.SSLMode)
	}
	if !strings.Contains(cfg.GetDatabaseDSN(), "sslmode=require") {
		t.Errorf("DSN %q ignores the TLS mode", cfg.GetDatabaseDSN())
	}

	t.Setenv("DB_SSL_MODE", "disable")
	if cfg, err = Load(); err != nil || cfg.Database.SSLMode != "disable" {
		t.Errorf("an explicit DB_SSL_MODE must win, got %v, %v", cfg, err)
	}
}

func TestLoadKeepsDevelopmentDefaults(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_SSL_MODE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("got DB_SSL_MODE=%q, want disable", cfg.Database.SSLMode)
	}
}

func TestLoadRejectsUnusableValues(t *testing.T) {
	t.Setenv("DB_SSL_MODE", "sometimes")
	t.Setenv("RATE_LIMIT_REQUESTS", "0")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, proxy.internal")

	_, err := Load()
	if err == nil {
		t.Fatal("expected unusable values to be rejected")
	}
	for _, key := range []string{"DB_SSL_MODE", "RATE_LIMIT_REQUESTS", "proxy.internal"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("error %q does not mention %s", err, key)
		}
	}
	if strings.Contains(err.Error(), "10.0.0.0/8") {
		t.Errorf("error %q rejects a valid CIDR", err)
	}
}
