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
