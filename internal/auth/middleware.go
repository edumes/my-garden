package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/pkg/auth"
	"gorm.io/gorm"
)

// AuthMiddleware accepts requests whose bearer token is valid and carries the
// user's current token version, so tokens issued before a logout fail.
func AuthMiddleware(jwtManager *auth.JWTManager, repo Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		version, err := repo.TokenVersion(claims.UserID)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound), err == nil && version != claims.TokenVersion:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session revoked, please sign in again"})
			return
		case err != nil:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to check session"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Next()
	}
}
