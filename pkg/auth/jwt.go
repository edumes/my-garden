package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/config"
)

const issuer = "my-garden-api"

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	// TokenVersion must match the user's current token version. Logging out
	// increments it, which revokes every token issued before.
	TokenVersion int `json:"token_version"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey string
	expiry    time.Duration
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		secretKey: cfg.JWT.Secret,
		expiry:    cfg.JWT.Expiry,
	}
}

// GenerateToken signs a token for the user and returns when it expires.
func (j *JWTManager) GenerateToken(userID uuid.UUID, username, email string, tokenVersion int) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.expiry)
	claims := &Claims{
		UserID:       userID,
		Username:     username,
		Email:        email,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    issuer,
			Subject:   userID.String(),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(j.secretKey))
	return signed, expiresAt, err
}

// ValidateToken accepts only unexpired HS256 tokens this API issued. It does
// not check revocation; that needs the user's current token version.
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(*jwt.Token) (any, error) { return []byte(j.secretKey), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
