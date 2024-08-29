// handlers/auth_handler.go
package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yeboahd24/subscription-stripe/models"
	"github.com/yeboahd24/subscription-stripe/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

// This will be use in the route
func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		DB:        db,
		JWTSecret: jwtSecret,
	}
}

func Register(h *AuthHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.CustomUser
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if user.Email == "" || user.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
			return
		}

		if !utils.IsValidEmail(user.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}

		// Check if user already exist
		var existingUser models.CustomUser
		if err := h.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		// Hash the password

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		// utils.Log("Hashed Password during Registration:", string(hashedPassword)) // Add your logging utility
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword) // Ensure this is the correct password hash

		if err := h.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}

func Login(h *AuthHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginUser struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&loginUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.CustomUser
		if err := h.DB.Where("email = ?", loginUser.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials-(email)"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password)); err != nil {
			// Log the hashed password and the provided password for debugging
			// utils.Log("Hashed Password:", user.Password)                    // Add your logging utility
			// utils.Log("Provided Password:", loginUser.Password)             // Add your logging utility
			// utils.Log("Hash Length:", len(user.Password))                   // Log length of hashed password
			// utils.Log("Provided Password Length:", len(loginUser.Password)) // Log length of provided password
			// utils.Log("Comparison:", user.Password == loginUser.Password)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials-(password)"})
			return
		}

		token, err := utils.GenerateJWT(user.ID, h.JWTSecret, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

// PromoteToAdmin
func PromoteToAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userRequest struct {
			UserID string `json:"user_id" binding:"required"` // Keep as string for UUID
		}
		var userModel models.CustomUser

		// Change the way user is retrieved from context
		userID, exists := c.Get("user_id")
		if !exists || !isUserAdmin(db, userID) { // Check if user is admin
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		// Bind JSON to userRequest
		if err := c.ShouldBindJSON(&userRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Convert userRequest.UserID to uuid.UUID
		userIDUUID, err := uuid.Parse(userRequest.UserID) // Parse string to UUID
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
			return
		}

		// Query the user model using the userID
		if err := db.First(&userModel, userIDUUID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		userModel.IsAdmin = true

		// Save the updated user model to the database
		if err := db.Save(&userModel).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to promote user to admin"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User promoted to admin successfully"})
	}
}

func isUserAdmin(db *gorm.DB, userID interface{}) bool {
	// Convert userID to uuid.UUID
	id, ok := userID.(uuid.UUID)
	if !ok {
		return false // Return false if type assertion fails
	}

	var user models.CustomUser
	// Query the database to find the user
	if err := db.First(&user, id).Error; err != nil {
		return false // Return false if user not found
	}

	return user.IsAdmin // Return the admin status
}
