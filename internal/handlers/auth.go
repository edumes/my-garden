package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/my-garden/api/internal/database"
	"github.com/my-garden/api/internal/models"
	"github.com/my-garden/api/pkg/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *database.Database
	jwtManager *auth.JWTManager
}

func NewAuthHandler(db *database.Database, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
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
	Token     string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User      models.User `json:"user"`
	ExpiresAt time.Time   `json:"expires_at" example:"2024-01-02T00:00:00Z"`
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

	// Check if username already exists
	var existingUser models.User
	if err := h.db.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Check if email already exists
	if err := h.db.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	}

	if err := h.db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	h.db.DB.Save(&user)

	// Clear password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, AuthResponse{
		Token:     token,
		User:      user,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Default 24 hour expiry
	})
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

	// Find user by username
	var user models.User
	if err := h.db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	h.db.DB.Save(&user)

	// Clear password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		User:      user,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Default 24 hour expiry
	})
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

	// Extract token from "Bearer <token>"
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Refresh token
	newToken, err := h.jwtManager.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      newToken,
		"expires_at": time.Now().Add(24 * time.Hour), // Default 24 hour expiry
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
	// In a stateless JWT system, logout is typically handled client-side
	// by removing the token from storage. However, we can implement
	// a token blacklist if needed for additional security.

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

	var user models.User
	if err := h.db.DB.Preload("Achievements.Achievement").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Clear password hash from response
	user.PasswordHash = ""

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

	var user models.User
	if err := h.db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Timezone != "" {
		user.Timezone = req.Timezone
	}
	if req.Language != "" {
		user.Language = req.Language
	}

	if err := h.db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Clear password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}
