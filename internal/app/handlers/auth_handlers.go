package handlers

import (
	"adminbe/internal/app/models"
	"adminbe/internal/pkg/utils"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginRequest represents login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// loginHandler POST /api/auth/login
func loginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var user models.User
		result := db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", req.Email).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				log.Printf("Login failed: user not found for email %s", req.Email)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			log.Printf("Error querying user for login: %v", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Check password
		err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			log.Printf("Login failed: incorrect password for email %s", req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Check status
		if user.Status != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Account disabled"})
			return
		}

		// Generate JWT
		jwtSecret := utils.GetJWTSecret()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  strconv.FormatUint(user.ID, 10),
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		})

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			log.Printf("Error generating JWT: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString, "user": gin.H{"id": user.ID, "username": user.Username, "email": user.Email}})
	}
}
