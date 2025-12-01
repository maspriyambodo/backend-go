package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Pre-allocated user slice to avoid repeated allocations
var usersPool = make([]models.User, 0, 50) // Initial capacity of 50

// listUsersHandler GET /api/users
func listUsersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by FROM users WHERE deleted_at IS NULL")
		if err != nil {
			log.Printf("Error querying users: %v", err)
			response := utils.GetResponseObject()
			response["error"] = "Failed to retrieve users"
			c.JSON(http.StatusInternalServerError, response)
			utils.PutResponseObject(response)
			return
		}
		defer rows.Close()

		// Reset slice to zero length but keep capacity
		users := utils.ResetSliceToZeroLen(usersPool)

		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.DeletedBy); err != nil {
				log.Printf("Error scanning user row: %v", err)
				response := utils.GetResponseObject()
				response["error"] = "Failed to retrieve users"
				c.JSON(http.StatusInternalServerError, response)
				utils.PutResponseObject(response)
				return
			}
			u.PasswordHash = "" // not needed
			users = append(users, u)
		}

		response := utils.GetResponseObject()
		response["data"] = users
		c.JSON(http.StatusOK, response)
		utils.PutResponseObject(response)
	}
}

// getUserHandler GET /api/users/:id
func getUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var u models.User
		row := db.QueryRow("SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by FROM users WHERE id = ? AND deleted_at IS NULL", userID)
		err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": u})
	}
}

// createUserHandler POST /api/users
func createUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
			return
		}

		status := uint8(1) // default active
		if req.Status != nil {
			status = *req.Status
		}

		now := time.Now()
		result, err := db.Exec("INSERT INTO users (username, email, password_hash, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			req.Username, req.Email, string(hashed), status, now, now)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
			if strings.Contains(err.Error(), "1062") {
				c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			}
			return
		}

		userID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "User created", "id": userID})
		createAuditLog(db, nil, "CREATE", "users", uint64(userID), nil, req)
	}
}

// updateUserHandler PUT /api/users/:id
func updateUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var req models.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL", userID).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Build update query
		setParts := []string{}
		args := []interface{}{}

		if req.Username != "" {
			setParts = append(setParts, "username = ?")
			args = append(args, req.Username)
		}
		if req.Email != "" {
			setParts = append(setParts, "email = ?")
			args = append(args, req.Email)
		}
		if req.Password != "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
				return
			}
			setParts = append(setParts, "password_hash = ?")
			args = append(args, string(hashed))
		}
		if req.Status != nil {
			setParts = append(setParts, "status = ?")
			args = append(args, *req.Status)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		setParts = append(setParts, "updated_at = ?")
		args = append(args, time.Now())

		query := buildUpdateQuery(setParts, "users")
		args = append(args, userID)

		_, err = db.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated"})
	}
}

// deleteUserHandler DELETE /api/users/:id (soft delete)
func deleteUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Assume deleted_by is from context, e.g., JWT. For now, set to nil
		_, err = db.Exec("UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", time.Now(), time.Now(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
	}
}

// Optimized UPDATE query building
func buildUpdateQuery(setParts []string, table string) string {
	setClause := utils.JoinStrings(setParts, ", ")
	return "UPDATE " + table + " SET " + setClause + " WHERE id = ? AND deleted_at IS NULL"
}

// createAuditLog logs audit events
func createAuditLog(db *sql.DB, userID *uint64, event string, table string, recordID uint64, oldValues, newValues interface{}) {
	var oldJSON, newJSON []byte
	if oldValues != nil {
		oldJSON, _ = json.Marshal(oldValues)
	}
	if newValues != nil {
		newJSON, _ = json.Marshal(newValues)
	}
	db.Exec("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values) VALUES (?, ?, ?, ?, ?, ?)", userID, event, table, recordID, oldJSON, newJSON)
}
