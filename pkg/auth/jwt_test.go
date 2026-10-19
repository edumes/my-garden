package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/config"
)

const testSecret = "test-secret-with-at-least-32-characters"

func newManager() *JWTManager {
	return NewJWTManager(&config.Config{JWT: config.JWTConfig{Secret: testSecret, Expiry: time.Hour}})
}

func TestValidateTokenAcceptsIssuedTokens(t *testing.T) {
	userID := uuid.New()
	token, expiresAt, err := newManager().GenerateToken(userID, "gardener", "g@example.com", 3)
	if err != nil {
		t.Fatal(err)
	}
	if time.Until(expiresAt) < 59*time.Minute {
		t.Errorf("expiry %s does not follow the configured hour", expiresAt)
	}

	claims, err := newManager().ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != userID || claims.TokenVersion != 3 {
		t.Errorf("got user %s version %d", claims.UserID, claims.TokenVersion)
	}
}

func TestValidateTokenRejectsForeignTokens(t *testing.T) {
	valid := func() jwt.RegisteredClaims {
		return jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}
	}
	sign := func(method jwt.SigningMethod, key any, registered jwt.RegisteredClaims) string {
		token, err := jwt.NewWithClaims(method, &Claims{UserID: uuid.New(), RegisteredClaims: registered}).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}

	otherIssuer := valid()
	otherIssuer.Issuer = "someone-else"
	noExpiry := valid()
	noExpiry.ExpiresAt = nil
	expired := valid()
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	cases := map[string]string{
		"other algorithm": sign(jwt.SigningMethodHS512, []byte(testSecret), valid()),
		"no signature":    sign(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, valid()),
		"other secret":    sign(jwt.SigningMethodHS256, []byte("another-secret-of-at-least-32-chars"), valid()),
		"other issuer":    sign(jwt.SigningMethodHS256, []byte(testSecret), otherIssuer),
		"no expiry":       sign(jwt.SigningMethodHS256, []byte(testSecret), noExpiry),
		"expired":         sign(jwt.SigningMethodHS256, []byte(testSecret), expired),
	}
	for name, token := range cases {
		if _, err := newManager().ValidateToken(token); err == nil {
			t.Errorf("%s: token was accepted", name)
		}
	}
}
