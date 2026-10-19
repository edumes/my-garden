package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/my-garden/api/internal/audit"
)

type AuthHandler struct {
	service      Service
	auditService *audit.AuditService
}

func NewAuthHandler(service Service, auditService *audit.AuditService) *AuthHandler {
	return &AuthHandler{
		service:      service,
		auditService: auditService,
	}
}

type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=20" example:"gardener123"`
	Email     string `json:"email" binding:"required,email" example:"gardener@example.com"`
	Password  string `json:"password" binding:"required,min=6" example:"securepassword123"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"gardener123"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

type AuthResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User      User      `json:"user"`
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-02T00:00:00Z"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with the provided information
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 409 {object} map[string]interface{} "Conflict - Username or email already exists"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Register(req)
	if err != nil {
		switch err {
		case ErrUsernameTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		case ErrEmailTaken:
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		default:
			h.auditService.LogFailure(c, audit.AuditActionUserRegister, audit.AuditResourceUser, &response.User.ID, gin.H{"error": err.Error()})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		}
		return
	}

	h.auditService.LogSuccess(c, audit.AuditActionUserRegister, audit.AuditResourceUser, &response.User.ID, gin.H{
		"username": response.User.Username,
		"email":    response.User.Email,
		"user_id":  response.User.ID,
	})

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Authenticate user
// @Description Login with username and password to get JWT token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.Login(req)
	if err != nil {
		switch err {
		case ErrUserNotFound, ErrInvalidPassword:
			h.auditService.LogFailure(c, audit.AuditActionUserLogin, audit.AuditResourceUser, nil, gin.H{"error": "Invalid credentials"})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		}
		return
	}

	h.auditService.LogSuccess(c, audit.AuditActionUserLogin, audit.AuditResourceUser, &response.User.ID, gin.H{
		"username": response.User.Username,
		"user_id":  response.User.ID,
	})

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh the current JWT token to get a new one with extended expiry
// @Tags authentication
// @Accept json
// @Produce json
// @Security bearer
// @Success 200 {object} map[string]interface{} "New token and expiry"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return
	}

	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	newToken, expiresAt, err := h.service.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      newToken,
		"expires_at": expiresAt,
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout user (client-side token removal)
// @Tags authentication
// @Accept json
// @Produce json
// @Security bearer
// @Success 200 {object} map[string]interface{} "Logout success message"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the current user's profile and achievements
// @Tags users
// @Accept json
// @Produce json
// @Security bearer
// @Success 200 {object} map[string]interface{} "User profile with achievements"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.service.GetProfile(userID.(uuid.UUID))
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the current user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security bearer
// @Param request body map[string]interface{} true "Profile update data"
// @Success 200 {object} map[string]interface{} "Updated user profile"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		FirstName string `json:"first_name" example:"Johnny"`
		LastName  string `json:"last_name" example:"Smith"`
		Avatar    string `json:"avatar" example:"https://example.com/avatar.jpg"`
		Timezone  string `json:"timezone" example:"America/New_York"`
		Language  string `json:"language" example:"en"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateProfile(
		userID.(uuid.UUID),
		req.FirstName,
		req.LastName,
		req.Avatar,
		req.Timezone,
		req.Language,
	)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
