package middleware

import (
	"adminbe/internal/pkg/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// RequestLoggerMiddleware logs incoming requests to console
// Removed per-request audit logging to prevent memory allocation from JSON marshaling
// Audit logs should be created selectively in handlers for important actions only
func RequestLoggerMiddleware(db *sql.DB) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logStr := fmt.Sprintf("%s - [%s] %s %s %s %d %s %s\n",
			param.ClientIP,
			param.TimeStamp.Format("2006/01/02 15:04:05"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency.String(),
			param.Request.UserAgent(),
		)

		// Log to console only - removed audit logging for every request
		log.Print(logStr)

		return logStr
	})
}

// logRequestToAudit creates an audit log entry for the request
func logRequestToAudit(db *sql.DB, param gin.LogFormatterParams) {
	eventType := "API_ACCESS"
	if param.StatusCode >= 400 {
		eventType = "API_ERROR"
	}

	// Extract user ID from path or from context (if implemented in handlers)
	userID := (*uint64)(nil)
	if strings.HasPrefix(param.Path, "/api/users/") && param.Path != "/api/users" {
		// Could parse user ID from URL if needed, but skipping for now
	}

	requestData := map[string]interface{}{
		"method":      param.Method,
		"path":        param.Path,
		"status_code": param.StatusCode,
		"latency_ms":  param.Latency.Milliseconds(),
		"user_agent":  param.Request.UserAgent(),
		"referer":     param.Request.Referer(),
	}

	requestJSON, _ := json.Marshal(requestData)

	// Insert audit log
	db.Exec("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, new_values) VALUES (?, ?, ?, ?, ?)",
		userID, eventType, "api_requests", 0, requestJSON)
}

// CustomRecoveryMiddleware provides panic recovery with logging
func CustomRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
		} else {
			log.Printf("Panic recovered: %v", recovered)
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error occurred",
		})
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// AuthMiddleware checks JWT token and sets user ID in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		jwtSecret := utils.GetJWTSecret()

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if userIDStr, ok := claims["user_id"].(string); ok {
				userID, err := strconv.ParseUint(userIDStr, 10, 64)
				if err != nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
					c.Abort()
					return
				}
				c.Set("user_id", userID)
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}
